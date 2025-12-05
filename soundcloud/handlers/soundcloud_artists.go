package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/soundcloud/artist"
)

// GetArtistsSoundCloud retrieves the artists (users) that the SoundCloud user is following
func GetArtistsSoundCloud(c *gin.Context) {
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, artists)
}
