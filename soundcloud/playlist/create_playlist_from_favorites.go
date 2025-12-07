package playlist

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
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

type CreatePlaylistFromFavoritesRequest struct {
	Minute int `json:"minute" binding:"required,min=1"`
}

// CreatePlaylistFromFavorites creates a SoundCloud playlist from user's favorite tracks
// Returns playlistID, secretToken, and error
func CreatePlaylistFromFavorites(c *gin.Context) (string, string, error) {
	var json CreatePlaylistFromFavoritesRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Error("failed to bind JSON", slog.Any("error", err))
		return "", "", err
	}

	slog.Info("creating playlist from favorites", slog.Int("duration_minutes", json.Minute))

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

	// Get favorite tracks from database
	tracks, err := getTracksFromFavorites(dbInstance, specifyMs, user.Id)
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
	description := fmt.Sprintf("Generated playlist for %d minutes from favorites", json.Minute)

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

	// Increment playlist count (non-fatal)
	if err = database.IncrementSoundCloudPlaylistCount(dbInstance, user.Id); err != nil {
		slog.Warn("failed to increment playlist count", slog.Any("error", err))
	}

	slog.Info("playlist created successfully", slog.String("playlist_id", playlistID), slog.String("secret_token", playlist.SecretToken), slog.Int("tracks", len(trackIDs)))
	return playlistID, playlist.SecretToken, nil
}

// getTracksFromFavorites retrieves and processes favorite tracks with retry mechanism
func getTracksFromFavorites(db *sql.DB, specifyMs int, userId string) ([]model.Track, error) {
	// Get favorite tracks from database
	saveTracks, err := database.GetSoundCloudFavoriteTracks(db, userId)
	if err != nil {
		slog.Error("database error", slog.Any("error", err))
		return nil, err
	}

	if len(saveTracks) == 0 {
		slog.Error("no favorite tracks in database")
		return nil, model.ErrNoFavoriteTracks
	}

	// Calculate total available duration
	totalDuration := 0
	for _, track := range saveTracks {
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
				shuffled := shuffleTracks(saveTracks)
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
				slog.Int("track_count", len(saveTracks)),
				slog.Int("try_count", finalTryCount),
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			slog.Warn("timeout: combination not found",
				slog.Int("duration_minutes", specifyMs/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(saveTracks)),
				slog.Int("total_duration_minutes", totalDuration/commontrack.MillisecondsPerMinute),
				slog.Int("try_count", finalTryCount),
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}

// shuffleTracks shuffles the track list using Fisher-Yates algorithm
func shuffleTracks(tracks []model.Track) []model.Track {
	shuffled := make([]model.Track, len(tracks))
	copy(shuffled, tracks)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}
