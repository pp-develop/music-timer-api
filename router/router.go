package router

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
)

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
	router.DELETE("playlist", deletePlaylists)
	return router
}

func getAuthzUrl(c *gin.Context) {
	url, err := api.Authz(c)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "")
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func callback(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = api.Callback(c)
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusInternalServerError, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

func getUserProfile(c *gin.Context) {
	spotify.GetUserProfile()
}

func getTracks(c *gin.Context) {
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	
	minute, _ := strconv.Atoi(c.Query("minute"))
	tracks, err := api.GetTracks(minute * oneminuteToMsec)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, tracks)
	}
}

func createPlaylist(c *gin.Context) {
	playlistId, err := api.CreatePlaylistBySpecifyTime(c)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}

}
func deletePlaylists(c *gin.Context) {
	err := api.DeletePlaylists(c)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}
