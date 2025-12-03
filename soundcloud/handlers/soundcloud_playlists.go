package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/playlist"
	"github.com/pp-develop/music-timer-api/utils"
)

// CreatePlaylistFromFavorites creates a SoundCloud playlist from user's favorite tracks
func CreatePlaylistFromFavorites(c *gin.Context) {
	playlistId, secretToken, err := playlist.CreatePlaylistFromFavorites(c)
	if err != nil {
		log.Printf("[HANDLER] Error creating playlist from favorites: %v", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"playlist_id": playlistId, "secret_token": secretToken})
}

// CreatePlaylistFromArtists creates a SoundCloud playlist from specified artists
func CreatePlaylistFromArtists(c *gin.Context) {
	playlistId, secretToken, err := playlist.CreatePlaylistFromArtists(c)
	if err != nil {
		log.Printf("[HANDLER] Error creating playlist from artists: %v", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"playlist_id": playlistId, "secret_token": secretToken})
}

// GetPlaylistsSoundCloud retrieves user's SoundCloud playlists
func GetPlaylistsSoundCloud(c *gin.Context) {
	userId, err := utils.GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		c.Error(model.ErrFailedGetDB)
		return
	}

	playlists, err := database.GetSoundCloudPlaylists(dbInstance, userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// DeletePlaylistsSoundCloud deletes a SoundCloud playlist
func DeletePlaylistsSoundCloud(c *gin.Context) {
	var json struct {
		PlaylistId string `json:"playlist_id"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.Error(err)
		return
	}

	userId, err := utils.GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		c.Error(model.ErrFailedGetDB)
		return
	}

	err = database.DeleteSoundCloudPlaylists(dbInstance, json.PlaylistId, userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
