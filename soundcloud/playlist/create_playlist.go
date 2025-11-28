package playlist

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"

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
		return "", err
	}

	// Convert minutes to milliseconds
	specify_ms := json.Minute * commontrack.MillisecondsPerMinute

	// Get authenticated user
	user, err := auth.GetSoundCloudAuthStatus(c)
	if err != nil {
		return "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return "", model.ErrFailedGetDB
	}

	// Get favorite tracks from database
	tracks, err := getSoundCloudTracks(dbInstance, specify_ms, user.Id)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if len(tracks) == 0 {
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

	// Extract track URLs
	trackURLs := make([]string, len(tracks))
	for i, track := range tracks {
		trackURLs[i] = track.Uri
	}

	// Add tracks to playlist
	err = client.AddTracksToPlaylist(user.AccessToken, playlist.ID, trackURLs)
	if err != nil {
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

// getSoundCloudTracks retrieves and processes tracks using existing MakeTracks logic
func getSoundCloudTracks(db *sql.DB, specify_ms int, userId string) ([]model.Track, error) {
	// Get favorite tracks from database
	saveTracks, err := database.GetSoundCloudFavoriteTracks(db, userId)
	if err != nil {
		return nil, err
	}

	if len(saveTracks) == 0 {
		return nil, model.ErrNoFavoriteTracks
	}

	// Use existing track selection logic
	// Shuffle tracks using the json package
	shuffledTracks := shuffleTracks(saveTracks)

	// Use MakeTracks to select tracks that fit the specified duration
	success, tracks := commontrack.MakeTracks(shuffledTracks, specify_ms)

	if !success {
		return nil, model.ErrNotEnoughTracks
	}

	return tracks, nil
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
