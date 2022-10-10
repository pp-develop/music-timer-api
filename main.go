package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/internal"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func main() {
	router := gin.Default()
	router.GET("/authz-url", getAuthzUrl)
	router.GET("/callback", callback)
	router.GET("/user", getUserProfile)
	router.GET("/tracks", getTracks)
	router.POST("playlist", createPlaylist)
	router.Run(":8080")
}

func getAuthzUrl(c *gin.Context) {
	success, url := internal.Authz()
	if success {
		c.JSON(http.StatusOK, gin.H{"url": url})
	} else {
		c.JSON(http.StatusInternalServerError, "")
	}
}

func callback(c *gin.Context) {
	code := c.Query("code")
	success := internal.Callback(code)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if !success {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_ERROR_URL"))
	}
}

func getUserProfile(c *gin.Context) {
	internal.GetUserProfile()
}

func getTracks(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	success, tracks := internal.GetTracks(minute * ONEMINUTE_TO_MSEC)
	if success {
		c.IndentedJSON(http.StatusOK, tracks)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, "")
	}
}

func createPlaylist(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	isCreate, playlistId := internal.CreatePlaylistBySpecifyTime(minute * ONEMINUTE_TO_MSEC)
	if isCreate {
		c.IndentedJSON(http.StatusCreated, playlistId)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, "")
	}
}
