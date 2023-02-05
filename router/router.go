package router

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func Create() *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		// TODO envを参照
		AllowOrigins: []string{
			"http://localhost:19006",
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	store := cookie.NewStore([]byte("secret")) // TODO envを参照する
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/auth", getAuth)
	router.GET("/authz-url", getAuthzUrl)
	router.DELETE("/session", deleteSession)
	router.GET("/callback", callback)
	router.GET("/tracks", getTracks)
	router.POST("/playlist", createPlaylist)
	router.DELETE("/playlist", deletePlaylists)
	return router
}

func getAuth(c *gin.Context) {
	err := api.Auth(c)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else if err == model.ErrFailedGetSession {
		log.Println(err)
		c.IndentedJSON(http.StatusUnauthorized, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
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

func deleteSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Save()
	c.JSON(http.StatusOK, "")
}

func callback(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		c.IndentedJSON(http.StatusInternalServerError, "")
	}

	err = api.Callback(c)
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusInternalServerError, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

func createPlaylist(c *gin.Context) {
	playlistId, err := api.CreatePlaylistBySpecifyTime(c)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else if err == model.ErrFailedGetSession {
		log.Println(err)
		c.IndentedJSON(http.StatusUnauthorized, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
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

func deletePlaylists(c *gin.Context) {
	err := api.DeletePlaylists(c)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else if err == model.ErrFailedGetSession {
		log.Println(err)
		c.IndentedJSON(http.StatusUnauthorized, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}
