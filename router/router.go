package router

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/middleware"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/artist"
	"github.com/pp-develop/music-timer-api/pkg/auth"
	"github.com/pp-develop/music-timer-api/pkg/json"
	"github.com/pp-develop/music-timer-api/pkg/logger"
	"github.com/pp-develop/music-timer-api/pkg/playlist"
	"github.com/pp-develop/music-timer-api/pkg/search"
	"github.com/pp-develop/music-timer-api/utils"
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
			os.Getenv("NATIVE_BASE_URL"),
			os.Getenv("NATIVE_API_URL"),
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

	// Cookie configuration: Secure and SameSite based on environment
	// Development: Secure=false, SameSite=Lax (allows HTTP)
	// Production: Secure=true, SameSite=None (HTTPS only, cross-site allowed)
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	var sameSiteMode http.SameSite
	if isProduction {
		sameSiteMode = http.SameSiteNoneMode // Cross-site allowed (requires HTTPS)
	} else {
		sameSiteMode = http.SameSiteLaxMode // Normal mode (works with HTTP)
	}

	store := cookie.NewStore([]byte(os.Getenv("COOKIE_SECRET")))
	store.Options(sessions.Options{
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: sameSiteMode,
		Path:     "/",
	})

	log.Printf("[INFO] Session configuration - Environment: %s, Secure: %v, SameSite: %v",
		os.Getenv("ENVIRONMENT"), isProduction, sameSiteMode)
	router.Use(sessions.Sessions("mysession", store))
	router.Use(middleware.DBMiddleware())

	router.GET("/health", healthCheck)

	// Web authentication endpoints
	authWeb := router.Group("/auth/web")
	{
		authWeb.GET("/authz-url", getAuthzUrlWeb)
		authWeb.GET("/callback", callbackWeb)
		authWeb.DELETE("/session", deleteSession)
	}

	// Native authentication endpoints
	authNative := router.Group("/auth/native")
	{
		authNative.GET("/authz-url", getAuthzUrlNative)
		authNative.GET("/callback", callbackNative)
		authNative.POST("/refresh", refreshTokenNative)
	}

	// Common authentication endpoint
	router.GET("/auth/status", getAuthStatus)

	// Protected endpoints (support both session and JWT auth)
	router.Use(middleware.OptionalAuthMiddleware())
	router.POST("/tracks", saveTracks)
	router.POST("/tracks/init", initTrackData)
	router.POST("/reset-tracks", resetTracks)
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

// Web authentication handlers
func callbackWeb(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	err = auth.SpotifyCallbackWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

func getAuthzUrlWeb(c *gin.Context) {
	url, err := auth.SpotifyAuthzWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// Native authentication handlers
func callbackNative(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError(err)
		// エラー時はディープリンクでエラーを通知
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL_NATIVE")+"?error=config_error")
		return
	}

	tokenPair, err := auth.SpotifyCallbackNative(c)
	if err != nil {
		logger.LogError(err)
		// 認証失敗時
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL_NATIVE")+"?error=auth_failed")
		return
	}

	// クエリパラメータ（?）を使用してトークンを渡す
	redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&expires_in=%d",
		os.Getenv("AUTHZ_SUCCESS_URL_NATIVE"),
		tokenPair.AccessToken,
		tokenPair.RefreshToken,
		tokenPair.ExpiresIn,
	)

	c.Redirect(http.StatusSeeOther, redirectURL)
}

func getAuthzUrlNative(c *gin.Context) {
	url, err := auth.SpotifyAuthzNative(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func refreshTokenNative(c *gin.Context) {
	tokenPair, err := auth.RefreshAccessToken(c)
	if err != nil {
		logger.LogError(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

// Common authentication handler
func getAuthStatus(c *gin.Context) {
	user, err := auth.Auth(c)

	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Status(http.StatusSeeOther)
	} else if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"country": user.Country})
	}
}

func deleteSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Save()
	c.Status(http.StatusOK)
}

func saveTracks(c *gin.Context) {
	dbInstance, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}
	err := search.SaveTracks(c, dbInstance)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

func initTrackData(c *gin.Context) {
	err := search.SaveFavoriteTracks(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	err = search.SaveTracksFromFollowedArtists(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func resetTracks(c *gin.Context) {
	err := json.ReCreate(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

func deleteTracks(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	err := database.DeleteTracks(db)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	err = database.DeleteOldTracksIfOverLimit(db)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func updateFavoriteTracks(c *gin.Context) {
	err := search.SaveFavoriteTracks(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

func getArtists(c *gin.Context) {
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, artists)
}

func getPlaylist(c *gin.Context) {
	playlist, err := playlist.GetPlaylists(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, playlist)
}

func createPlaylist(c *gin.Context) {
	playlistId, err := playlist.CreatePlaylist(c)
	if err == nil {
		c.IndentedJSON(http.StatusCreated, playlistId)
		return
	}

	logger.LogError(err)

	// エラーケースごとに適切なレスポンスを返す
	switch err {
	case model.ErrNotFoundTracks:
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTracksNotFound,
		})
	case model.ErrNotEnoughTracks:
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutInsufficientTracks,
		})
	case model.ErrTimeoutCreatePlaylist:
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutNoMatch,
		})
	case model.ErrNoFavoriteTracks:
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeNoFavoriteTracks,
		})
	case model.ErrSpotifyRateLimit:
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
			Code: model.CodeSpotifyRateLimit,
		})
	case model.ErrPlaylistQuotaExceeded:
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
			Code: model.CodePlaylistQuotaExceeded,
		})
	case model.ErrPlaylistCreationFailed:
		c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Code: model.CodePlaylistCreationFailed,
		})
	case model.ErrAccessTokenExpired:
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code: model.CodeTokenExpired,
		})
	default:
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: model.CodeInternalError,
		})
	}
}

func gestCreatePlaylist(c *gin.Context) {
	playlistId, err := playlist.GestCreatePlaylist(c)
	if err == model.ErrNotEnoughTracks {
		logger.LogError(err)
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutInsufficientTracks,
		})
	} else if err == model.ErrTimeoutCreatePlaylist {
		logger.LogError(err)
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutNoMatch,
		})
	} else if err != nil {
		logger.LogError(err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: model.CodeInternalError,
		})
	} else {
		c.IndentedJSON(http.StatusCreated, playlistId)
	}
}

func deletePlaylists(c *gin.Context) {
	err := playlist.DeletePlaylists(c)
	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Status(http.StatusSeeOther)
	} else if err == model.ErrNotFoundPlaylist {
		logger.LogError(err)
		c.Status(http.StatusNoContent)
	} else if err == model.ErrAccessTokenExpired {
		logger.LogError(err)
		c.Status(http.StatusUnauthorized)
	} else if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}
