package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func main() {
	router := gin.Default()
	router.GET("/authz", authz)
	router.GET("/user", getUserProfile)
	router.GET("/playlist", getPlaylist)
	router.POST("playlist", postPlaylist)

	router.Run(":8080")
}

func authz(c *gin.Context) {
	pkg.Authz()
}

func getUserProfile(c *gin.Context) {
	pkg.GetUserProfile()
}

func getPlaylist(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	playlist := pkg.GetPlaylist(minute * ONEMINUTE_TO_MSEC)
	c.IndentedJSON(http.StatusOK, playlist)
}

func postPlaylist(c *gin.Context) {
	pkg.CreatePlaylist("", "")
}
