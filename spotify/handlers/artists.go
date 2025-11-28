package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/spotify/artist"
)

// GetArtists returns the user's followed artists
func GetArtists(c *gin.Context) {
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, artists)
}
