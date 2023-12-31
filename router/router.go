package router

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/auth"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/playlist"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track"
)

func Create() *gin.Engine {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			os.Getenv("BASE_URL"),
			os.Getenv("API_URL"),
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
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

	store := cookie.NewStore([]byte(os.Getenv("COOKIE_SECRET")))
	store.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/callback", callback)
	router.GET("/auth", getAuth)
	router.GET("/authz-url", getAuthzUrl)
	router.DELETE("/session", deleteSession)
	router.POST("/tracks", saveTracks)
	router.DELETE("/tracks", deleteTracks)
	router.POST("/playlist", createPlaylist)
	router.DELETE("/playlist", deletePlaylists)
	return router
}

func callback(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		c.Redirect(http.StatusSeeOther, "/")
	}

	err = auth.SpotifyCallback(c)
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

func getAuth(c *gin.Context) {
	err := auth.Auth(c)

	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("BASE_URL"))
	} else if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}

func getAuthzUrl(c *gin.Context) {
	url, err := auth.SpotifyAuthz(c)
	if err != nil {
		logger.LogError(err)
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

func saveTracks(c *gin.Context) {
	var json track.RequestJson
	var err error
	c.ShouldBindJSON(&json)
	if json.IncludeFavoriteArtists {
		err = track.SearchTracksByFollowedArtists(c)
	} else {
		err = track.SearchAndSaveTracks()
	}

	if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}

func deleteTracks(c *gin.Context) {
	err := database.DeleteTracks()
	if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}

func createPlaylist(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylist(c)
	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("BASE_URL"))
	} else if err == model.ErrTimeoutCreatePlaylist {
		logger.LogError(err)
		c.IndentedJSON(http.StatusNotFound, "")
	} else if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
}

func deletePlaylists(c *gin.Context) {
	err := playlist.DeletePlaylists(c)
	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("BASE_URL"))
	} else if err == model.ErrNotFoundPlaylist {
		logger.LogError(err)
		c.IndentedJSON(http.StatusNoContent, "")
	} else if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}
