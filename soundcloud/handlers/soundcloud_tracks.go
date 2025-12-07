package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/soundcloud/playlist"
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
