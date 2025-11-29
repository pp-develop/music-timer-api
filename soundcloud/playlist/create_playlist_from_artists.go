package playlist

import (
	"context"
	"fmt"
	"log"
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
func CreatePlaylistFromArtists(c *gin.Context) (string, error) {
	var json CreatePlaylistFromArtistsRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Printf("[PLAYLIST-FROM-ARTISTS] Failed to bind JSON: %v", err)
		return "", err
	}

	log.Printf("[PLAYLIST-FROM-ARTISTS] Creating playlist: duration=%d minutes, artists=%v", json.Minute, json.ArtistIds)

	// Convert minutes to milliseconds
	specifyMs := json.Minute * commontrack.MillisecondsPerMinute

	// Get authenticated user
	user, err := auth.GetAuth(c)
	if err != nil {
		log.Printf("[PLAYLIST-FROM-ARTISTS] Authentication failed: %v", err)
		return "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		log.Println("[PLAYLIST-FROM-ARTISTS] Failed to get DB instance")
		return "", model.ErrFailedGetDB
	}

	// Get tracks from specified artists
	tracks, err := getTracksFromArtists(user.AccessToken, specifyMs, json.ArtistIds)
	if err != nil {
		log.Printf("[PLAYLIST-FROM-ARTISTS] Failed to get tracks: %v", err)
		return "", err
	}

	if len(tracks) == 0 {
		log.Println("[PLAYLIST-FROM-ARTISTS] No tracks available for playlist creation")
		return "", model.ErrNotEnoughTracks
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
		log.Printf("[PLAYLIST-FROM-ARTISTS] Failed to create playlist: %v", err)
		return "", err
	}

	// Save playlist to database
	playlistID := strconv.Itoa(playlist.ID)
	err = database.SaveSoundCloudPlaylist(dbInstance, playlistID, user.Id)
	if err != nil {
		log.Printf("[PLAYLIST-FROM-ARTISTS] Failed to save playlist to database: %v", err)
		return "", err
	}

	// Increment playlist count
	err = database.IncrementSoundCloudPlaylistCount(dbInstance, user.Id)
	if err != nil {
		log.Printf("[PLAYLIST-FROM-ARTISTS] Failed to increment playlist count: %v", err)
	}

	log.Printf("[PLAYLIST-FROM-ARTISTS] Playlist created successfully: id=%s, tracks=%d", playlistID, len(trackIDs))
	return playlistID, nil
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
			log.Printf("[GET-TRACKS-FROM-ARTISTS] Failed to get tracks for artist %s: %v", artistID, err)
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
		log.Println("[GET-TRACKS-FROM-ARTISTS] ERROR: No tracks found from specified artists!")
		return nil, model.ErrNotEnoughTracks
	}

	log.Printf("[GET-TRACKS-FROM-ARTISTS] Found %d tracks from %d artists", len(allTracks), len(artistIds))

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
			log.Printf("[タイムアウト] GetTracksFromArtists: トラック不足 - 必要=%d分, 利用可能=%d分, トラック数=%d, 試行回数=%d",
				specifyMs/commontrack.MillisecondsPerMinute,
				totalDuration/commontrack.MillisecondsPerMinute,
				len(allTracks),
				finalTryCount,
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			log.Printf("[タイムアウト] GetTracksFromArtists: 組み合わせ未発見 - 再生時間=%d分, トラック数=%d, 総再生時間=%d分, 試行回数=%d",
				specifyMs/commontrack.MillisecondsPerMinute,
				len(allTracks),
				totalDuration/commontrack.MillisecondsPerMinute,
				finalTryCount,
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}
