package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/internal"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func main() {
	router := gin.Default()
	router.GET("/authz", authz)
	router.GET("/user", getUserProfile)
	router.GET("/tracks", getTracks)
	router.POST("playlist", createPlaylist)
	router.PUT("playlist", addItemsPlaylist)
	router.Run(":8080")
}

func authz(c *gin.Context) {
	internal.Authz()
}

func getUserProfile(c *gin.Context) {
	internal.GetUserProfile()
}

func getTracks(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	tracks := internal.GetTracks(minute * ONEMINUTE_TO_MSEC)
	c.IndentedJSON(http.StatusOK, tracks)
}

func createPlaylist(c *gin.Context) {
	internal.CreatePlaylist("", "")
}

func addItemsPlaylist(c *gin.Context) {
	internal.AddItemsPlaylist("", "")
}
