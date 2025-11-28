package playlist

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/utils"
)

type CreatePlaylistRequest struct {
	Minute int `json:"minute"`
}

// CreatePlaylist creates a SoundCloud playlist with specified duration
func CreatePlaylist(c *gin.Context) (string, error) {
	var json CreatePlaylistRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Printf("[PLAYLIST-CREATE] Failed to bind JSON: %v", err)
		return "", err
	}

	// Convert minutes to milliseconds
	specify_ms := json.Minute * commontrack.MillisecondsPerMinute

	// Get authenticated user
	user, err := auth.GetSoundCloudAuthStatus(c)
	if err != nil {
		log.Printf("[PLAYLIST-CREATE] Auth failed: %v", err)
		return "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		log.Println("[PLAYLIST-CREATE] Failed to get DB instance")
		return "", model.ErrFailedGetDB
	}

	// Get favorite tracks from database
	tracks, err := getSoundCloudTracks(dbInstance, specify_ms, user.Id)
	if err != nil {
		log.Printf("[PLAYLIST-CREATE] Failed to get tracks: %v", err)
		return "", err
	}

	if len(tracks) == 0 {
		log.Println("[PLAYLIST-CREATE] ERROR: No tracks returned!")
		return "", model.ErrNotEnoughTracks
	}

	// Create playlist on SoundCloud
	client := soundcloud.NewClient()
	title := fmt.Sprintf("Playlist %d min", json.Minute)
	description := fmt.Sprintf("Generated playlist for %d minutes", json.Minute)

	playlist, err := client.CreatePlaylist(user.AccessToken, title, description)
	if err != nil {
		return "", err
	}

	// Extract track IDs
	trackIDs := make([]string, len(tracks))
	for i, track := range tracks {
		trackIDs[i] = track.ID
	}

	// Add tracks to playlist
	err = client.AddTracksToPlaylist(user.AccessToken, playlist.ID, trackIDs)
	if err != nil {
		log.Printf("[PLAYLIST-CREATE] Failed to add tracks: %v", err)
		return "", err
	}

	// Save playlist to database
	playlistID := strconv.Itoa(playlist.ID)
	err = database.SaveSoundCloudPlaylist(dbInstance, playlistID, user.Id)
	if err != nil {
		return "", err
	}

	// Increment playlist count
	err = database.IncrementSoundCloudPlaylistCount(dbInstance, user.Id)
	if err != nil {
		log.Printf("Failed to increment playlist count: %v", err)
	}

	return playlistID, nil
}

// getSoundCloudTracks retrieves and processes tracks using existing MakeTracks logic with retry mechanism
func getSoundCloudTracks(db *sql.DB, specify_ms int, userId string) ([]model.Track, error) {
	// Get favorite tracks from database
	saveTracks, err := database.GetSoundCloudFavoriteTracks(db, userId)
	if err != nil {
		log.Printf("[GET-SC-TRACKS] Database error: %v", err)
		return nil, err
	}

	if len(saveTracks) == 0 {
		log.Println("[GET-SC-TRACKS] ERROR: No favorite tracks in database!")
		return nil, model.ErrNoFavoriteTracks
	}

	// Calculate total available duration
	totalDuration := 0
	for _, track := range saveTracks {
		totalDuration += track.DurationMs
	}

	// Phase: Retry mechanism with timeout (same as Spotify version)
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
				success, tracks = commontrack.MakeTracks(shuffled, specify_ms)
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

		hasEnoughDuration := totalDuration >= specify_ms

		if !hasEnoughDuration {
			log.Printf("[タイムアウト] GetSoundCloudTracks: トラック不足 - 必要=%d分, 利用可能=%d分, トラック数=%d, 試行回数=%d",
				specify_ms/commontrack.MillisecondsPerMinute,
				totalDuration/commontrack.MillisecondsPerMinute,
				len(saveTracks),
				finalTryCount,
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			log.Printf("[タイムアウト] GetSoundCloudTracks: 組み合わせ未発見 - 再生時間=%d分, トラック数=%d, 総再生時間=%d分, 試行回数=%d",
				specify_ms/commontrack.MillisecondsPerMinute,
				len(saveTracks),
				totalDuration/commontrack.MillisecondsPerMinute,
				finalTryCount,
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}

// shuffleTracks shuffles the track list using Fisher-Yates algorithm
func shuffleTracks(tracks []model.Track) []model.Track {
	shuffled := make([]model.Track, len(tracks))
	copy(shuffled, tracks)

	// Use rand package for shuffling
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}
