package playlist

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/utils"
)

type CreatePlaylistFromArtistsRequest struct {
	Minute    int      `json:"minute" binding:"required,min=1"`
	ArtistIds []string `json:"artistIds" binding:"required,min=1"`
}

// CreatePlaylistFromArtists creates a SoundCloud playlist from specified artists' tracks
// Returns playlistID, secretToken, and error
func CreatePlaylistFromArtists(c *gin.Context) (string, string, error) {
	var json CreatePlaylistFromArtistsRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Error("failed to bind JSON", slog.Any("error", err))
		return "", "", err
	}

	slog.Info("creating playlist from artists", slog.Int("duration_minutes", json.Minute), slog.Any("artist_ids", json.ArtistIds))

	// Convert minutes to milliseconds
	specifyMs := json.Minute * commontrack.MillisecondsPerMinute

	// Get authenticated user
	user, err := auth.GetAuth(c)
	if err != nil {
		slog.Error("authentication failed", slog.Any("error", err))
		return "", "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		slog.Error("failed to get DB instance")
		return "", "", model.ErrFailedGetDB
	}

	// Get tracks from specified artists
	tracks, err := getTracksFromArtists(user.AccessToken, specifyMs, json.ArtistIds)
	if err != nil {
		slog.Error("failed to get tracks", slog.Any("error", err))
		return "", "", err
	}

	if len(tracks) == 0 {
		slog.Error("no tracks available for playlist creation")
		return "", "", model.ErrNotEnoughTracks
	}

	// Extract track IDs
	trackIDs := make([]string, len(tracks))
	for i, track := range tracks {
		trackIDs[i] = track.ID
	}

	// Create playlist on SoundCloud with tracks included
	client := soundcloud.NewClient()
	title := fmt.Sprintf("Playlist %d min", json.Minute)
	description := fmt.Sprintf("Generated playlist for %d minutes from artists", json.Minute)

	playlist, err := client.CreatePlaylist(user.AccessToken, title, description, trackIDs)
	if err != nil {
		slog.Error("failed to create playlist", slog.Any("error", err))
		return "", "", err
	}

	// Save playlist to database
	playlistID := strconv.Itoa(playlist.ID)
	err = database.SaveSoundCloudPlaylist(dbInstance, playlistID, user.Id)
	if err != nil {
		slog.Error("failed to save playlist to database", slog.Any("error", err))
		return "", "", err
	}

	// Increment playlist count
	err = database.IncrementSoundCloudPlaylistCount(dbInstance, user.Id)
	if err != nil {
		slog.Warn("failed to increment playlist count", slog.Any("error", err))
	}

	slog.Info("playlist created successfully", slog.String("playlist_id", playlistID), slog.String("secret_token", playlist.SecretToken), slog.Int("tracks", len(trackIDs)))
	return playlistID, playlist.SecretToken, nil
}

// getTracksFromArtists fetches tracks from all specified artists and selects tracks to match duration
func getTracksFromArtists(accessToken string, specifyMs int, artistIds []string) ([]model.Track, error) {
	client := soundcloud.NewClient()

	// Collect all tracks from all artists
	var allTracks []model.Track
	trackIDSet := make(map[string]bool)

	for _, artistID := range artistIds {
		tracks, err := client.GetUserTracks(accessToken, artistID)
		if err != nil {
			slog.Warn("failed to get tracks for artist", slog.String("artist_id", artistID), slog.Any("error", err))
			continue // Skip this artist and continue with others
		}

		// Add tracks, avoiding duplicates
		for _, track := range tracks {
			if !trackIDSet[track.ID] {
				trackIDSet[track.ID] = true
				allTracks = append(allTracks, track)
			}
		}
	}

	if len(allTracks) == 0 {
		slog.Error("no tracks found from specified artists")
		return nil, model.ErrNotEnoughTracks
	}

	slog.Info("found tracks from artists", slog.Int("track_count", len(allTracks)), slog.Int("artist_count", len(artistIds)))

	// Calculate total available duration
	totalDuration := 0
	for _, track := range allTracks {
		totalDuration += track.DurationMs
	}

	// Retry mechanism with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(commontrack.DefaultTimeoutSeconds)*time.Second)
	defer cancel()

	tracksChan := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCountChan := make(chan int, 1)
	tryCount := 0

	go func() {
		defer close(tracksChan)
		defer close(errChan)
		defer func() {
			tryCountChan <- tryCount
			close(tryCountChan)
		}()

		var tracks []model.Track
		success := false
		for !success {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffled := shuffleTracks(allTracks)
				success, tracks = commontrack.MakeTracks(shuffled, specifyMs)
			}
		}
		tracksChan <- tracks
	}()

	select {
	case tracks := <-tracksChan:
		<-tryCountChan
		return tracks, nil
	case err := <-errChan:
		<-tryCountChan
		return nil, err
	case <-ctx.Done():
		finalTryCount := <-tryCountChan

		hasEnoughDuration := totalDuration >= specifyMs

		if !hasEnoughDuration {
			slog.Warn("timeout: not enough tracks",
				slog.Int("required_minutes", specifyMs/commontrack.MillisecondsPerMinute),
				slog.Int("available_minutes", totalDuration/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(allTracks)),
				slog.Int("try_count", finalTryCount),
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			slog.Warn("timeout: combination not found",
				slog.Int("duration_minutes", specifyMs/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(allTracks)),
				slog.Int("total_duration_minutes", totalDuration/commontrack.MillisecondsPerMinute),
				slog.Int("try_count", finalTryCount),
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}
