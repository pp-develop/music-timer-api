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
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/artist"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/auth"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/playlist"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/search"
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

	router.GET("/health", healthCheck)
	router.GET("/callback", callback)
	router.GET("/auth", getAuth)
	router.GET("/authz-url", getAuthzUrl)
	router.DELETE("/session", deleteSession)
	router.POST("/tracks", saveTracks)
	router.DELETE("/tracks", deleteTracks)
	router.GET("/artists", getArtists)
	router.POST("/gest-playlist", gestCreatePlaylist)
	router.GET("/playlist", getPlaylist)
	router.POST("/playlist", createPlaylist)
	router.DELETE("/playlist", deletePlaylists)
	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func callback(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError(err)
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
	var json search.RequestJson
	var err error
	c.ShouldBindJSON(&json)
	if json.IncludeFavoriteArtists {
		err = search.SaveTracksByFollowedArtists(c)
	} else {
		err = search.SaveTracks()
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

func getArtists(c *gin.Context) {
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}
	c.IndentedJSON(http.StatusOK, artists)
}

func getPlaylist(c *gin.Context) {
	playlist, err := playlist.GetPlaylists(c)
	if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}
	c.IndentedJSON(http.StatusOK, playlist)
}

func createPlaylist(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylist(c)
	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("BASE_URL"))
	} else if err == model.ErrAccessTokenExpired {
		logger.LogError(err)
		c.IndentedJSON(http.StatusUnauthorized, "")
	} else if err == model.ErrTimeoutCreatePlaylist || err == model.ErrNotFoundTracks {
		logger.LogError(err)
		c.IndentedJSON(http.StatusNotFound, "")
	} else if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
}

func gestCreatePlaylist(c *gin.Context) {
	playlistId, err := playlist.GestCreatePlaylist(c)
	if err == model.ErrTimeoutCreatePlaylist {
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
	} else if err == model.ErrAccessTokenExpired {
		logger.LogError(err)
		c.IndentedJSON(http.StatusUnauthorized, "")
	} else if err != nil {
		logger.LogError(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}
