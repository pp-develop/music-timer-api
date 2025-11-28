package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/playlist"
	"github.com/pp-develop/music-timer-api/utils"
)

// CreatePlaylistSoundCloud creates a SoundCloud playlist
func CreatePlaylistSoundCloud(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylist(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"playlist_id": playlistId})
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
