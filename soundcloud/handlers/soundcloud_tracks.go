package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/playlist"
	"github.com/pp-develop/music-timer-api/utils"
)

// InitFavoriteTracksSoundCloud initializes SoundCloud favorite tracks
func InitFavoriteTracksSoundCloud(c *gin.Context) {
	err := playlist.SaveFavoriteTracks(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// InitFollowedArtistsTracksSoundCloud initializes tracks from followed artists
func InitFollowedArtistsTracksSoundCloud(c *gin.Context) {
	err := playlist.SaveTracksFromFollowedArtists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

// GetFavoriteTracksExistsSoundCloud checks if favorite tracks exist for the user
func GetFavoriteTracksExistsSoundCloud(c *gin.Context) {
	dbInstance, ok := utils.GetDB(c)
	if !ok {
		c.Error(model.ErrFailedGetDB)
		return
	}

	userId, err := utils.GetUserID(c)
	if err != nil {
		c.Error(model.ErrFailedGetSession)
		return
	}

	exists, err := database.ExistsSoundCloudFavoriteTracks(dbInstance, userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}
