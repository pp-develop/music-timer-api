package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/spotify/playlist"
)

// GetPlaylists returns the user's playlists
func GetPlaylists(c *gin.Context) {
	playlist, err := playlist.GetPlaylists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, playlist)
}

// CreatePlaylist creates a new playlist
func CreatePlaylist(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylist(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusCreated, playlistId)
}

// GestCreatePlaylist creates a guest playlist
func GestCreatePlaylist(c *gin.Context) {
	playlistId, err := playlist.GestCreatePlaylist(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusCreated, playlistId)
}

// DeletePlaylists deletes the user's playlists
func DeletePlaylists(c *gin.Context) {
	err := playlist.DeletePlaylists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// CreatePlaylistFromFavorites creates a playlist from user's favorite tracks
func CreatePlaylistFromFavorites(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylistFromFavorites(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusCreated, playlistId)
}

// CreatePlaylistFromArtists creates a playlist from specified artists' tracks
func CreatePlaylistFromArtists(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylistFromArtists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusCreated, playlistId)
}
