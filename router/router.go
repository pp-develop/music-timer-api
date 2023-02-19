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
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/auth", getAuth)
	router.GET("/authz-url", getAuthzUrl)
	router.DELETE("/session", deleteSession)
	router.GET("/callback", callback)
	router.GET("/tracks", getTracks)
	router.POST("/tracks", saveTracks)
	router.POST("/playlist", createPlaylist)
	router.POST("/playlist-with-favorite-artists", createPlaylistWithFavoriteArtists)
	router.DELETE("/playlist", deletePlaylists)
	return router
}

func getAuth(c *gin.Context) {
	err := api.Auth(c)

	if err == model.ErrFailedGetSession {
		log.Println(err)
		c.Redirect(http.StatusUnauthorized, os.Getenv("APP_URL"))
	} else if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
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
		c.Redirect(http.StatusInternalServerError, "/")
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
	playlistId, err := api.CreatePlaylist(c)
	if err == model.ErrFailedGetSession {
		log.Println(err)
		c.Redirect(http.StatusUnauthorized, os.Getenv("APP_URL"))
	} else if err == model.ErrTimeoutCreatePlaylist {
		log.Println(err)
		c.IndentedJSON(http.StatusNotFound, "")
	} else if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
}

func createPlaylistWithFavoriteArtists(c *gin.Context) {
	playlistId, err := api.CreatePlaylistWithFollowedArtists(c)
	if err == model.ErrFailedGetSession {
		log.Println(err)
		c.Redirect(http.StatusUnauthorized, os.Getenv("APP_URL"))
	} else if err == model.ErrTimeoutCreatePlaylist {
		log.Println(err)
		c.IndentedJSON(http.StatusNotFound, "")
	} else if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
}

func saveTracks(c *gin.Context) {
	err := api.SearchTracks()
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}

func getTracks(c *gin.Context) {
	// 1minute = 60000ms
	oneminuteToMsec := 60000

	minute, _ := strconv.Atoi(c.Query("minute"))
	tracks, err := api.GetTracksBySpecifyTime(minute * oneminuteToMsec)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, tracks)
	}
}

func deletePlaylists(c *gin.Context) {
	err := api.DeletePlaylists(c)
	if err == model.ErrFailedGetSession {
		log.Println(err)
		c.Redirect(http.StatusUnauthorized, os.Getenv("APP_URL"))
	} else if err == model.ErrNotFoundPlaylist {
		log.Println(err)
		c.IndentedJSON(http.StatusNotFound, "")
	} else if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, "")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
}
