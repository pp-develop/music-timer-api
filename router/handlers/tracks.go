package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/json"
	"github.com/pp-develop/music-timer-api/pkg/search"
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

// InitTrackData initializes track data by saving favorite tracks and tracks from followed artists
func InitTrackData(c *gin.Context) {
	err := search.SaveFavoriteTracks(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = search.SaveTracksFromFollowedArtists(c)
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
