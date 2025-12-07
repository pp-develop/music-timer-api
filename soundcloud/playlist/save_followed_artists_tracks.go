package playlist

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	soundcloud "github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/artist"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/utils"
)

const (
	maxConcurrency = 3
	timeout        = 600 * time.Second
)

var semaphore = make(chan struct{}, maxConcurrency)

// SaveTracksFromFollowedArtists fetches and saves tracks from all followed artists to database
func SaveTracksFromFollowedArtists(c *gin.Context) error {
	user, err := auth.GetAuth(c)
	if err != nil {
		return err
	}

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// Get followed artists
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		return err
	}

	slog.Info("fetching tracks from followed artists", slog.Int("artist_count", len(artists)))

	errChan := make(chan error, len(artists))
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := soundcloud.NewClient()

	for _, art := range artists {
		wg.Add(1)

		go func(art model.Artists) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			}

			// Get tracks for this artist
			tracks, err := client.GetUserTracks(user.AccessToken, art.Id)
			if err != nil {
				slog.Warn("error fetching tracks for artist", slog.String("artist_id", art.Id), slog.Any("error", err))
				return // Continue with other artists
			}

			if len(tracks) == 0 {
				slog.Debug("no tracks found for artist", slog.String("artist_id", art.Id))
				return
			}

			// Save to database
			if err := database.AddSoundCloudArtistTracks(db, art.Id, tracks); err != nil {
				slog.Error("error saving artist tracks", slog.String("artist_id", art.Id), slog.Any("error", err))
				errChan <- err
			}

			slog.Debug("saved tracks for artist", slog.String("artist_id", art.Id), slog.Int("track_count", len(tracks)))
		}(art)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	runtime.GC()

	slog.Info("finished saving tracks from followed artists")
	return nil
}
