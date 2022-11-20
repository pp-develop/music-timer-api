package router

import (
	// "fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func Create() *gin.Engine {
	router := gin.Default()

	// sessionの発行
	store := cookie.NewStore([]byte("secret")) // TODO envを参照する
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/authz-url", getAuthzUrl)
	router.GET("/callback", callback)
	router.GET("/user", getUserProfile)
	router.GET("/tracks", getTracks)
	router.POST("playlist", createPlaylist)
	return router
}

func getAuthzUrl(c *gin.Context) {
	success, url := api.Authz(c)
	if success {
		c.JSON(http.StatusOK, gin.H{"url": url})
	} else {
		c.JSON(http.StatusInternalServerError, "")
	}
}

func callback(c *gin.Context) {
	success := api.Callback(c)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if success {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	} else {
		c.Redirect(http.StatusInternalServerError, os.Getenv("AUTHZ_ERROR_URL"))
	}
}

func getUserProfile(c *gin.Context) {
	api.GetUserProfile()
}

func getTracks(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	success, tracks := api.GetTracks(minute * ONEMINUTE_TO_MSEC)
	if success {
		c.IndentedJSON(http.StatusOK, tracks)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, "")
	}
}

func createPlaylist(c *gin.Context) {
	isCreate, playlistId := api.CreatePlaylistBySpecifyTime(c)
	if isCreate {
		c.IndentedJSON(http.StatusCreated, playlistId)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, "")
	}
}
