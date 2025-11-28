package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/json"
	"github.com/pp-develop/music-timer-api/spotify/search"
	"github.com/pp-develop/music-timer-api/utils"
)

// SaveTracks saves tracks to the database
func SaveTracks(c *gin.Context) {
	dbInstance, ok := utils.GetDB(c)
	if !ok {
		c.Error(model.ErrFailedGetDB)
		return
	}
	err := search.SaveTracks(c, dbInstance)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// InitFavoriteTracks initializes track data by saving favorite tracks
func InitFavoriteTracks(c *gin.Context) {
	err := search.SaveFavoriteTracks(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// InitFollowedArtistsTracks initializes track data by saving tracks from followed artists
func InitFollowedArtistsTracks(c *gin.Context) {
	err := search.SaveTracksFromFollowedArtists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// ResetTracks recreates the track data
func ResetTracks(c *gin.Context) {
	err := json.ReCreate(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// DeleteTracks deletes tracks from the database
func DeleteTracks(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Error(model.ErrFailedGetDB)
		return
	}

	err := database.DeleteTracks(db)
	if err != nil {
		c.Error(err)
		return
	}
	err = database.DeleteOldTracksIfOverLimit(db)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}
