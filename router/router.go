package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/middleware"
	"github.com/pp-develop/music-timer-api/router/handlers"
)

// Create initializes and configures the Gin router
func Create() *gin.Engine {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	router := gin.Default()

	// Setup middleware
	setupCORS(router)
	setupSession(router)
	router.Use(middleware.DBMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Setup routes
	setupRoutes(router)

	return router
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", handlers.HealthCheck)

	// Web authentication endpoints
	authWeb := router.Group("/auth/web")
	{
		authWeb.GET("/authz-url", handlers.GetAuthzURLWeb)
		authWeb.GET("/callback", handlers.CallbackWeb)
		authWeb.GET("/status", handlers.GetAuthStatusWeb)
		authWeb.DELETE("/session", handlers.DeleteSession)
	}

	// Native authentication endpoints
	authNative := router.Group("/auth/native")
	{
		authNative.GET("/authz-url", handlers.GetAuthzURLNative)
		authNative.GET("/callback", handlers.CallbackNative)
		authNative.GET("/status", handlers.GetAuthStatusNative)
		authNative.POST("/refresh", handlers.RefreshTokenNative)
	}

	// Protected endpoints (support both session and JWT auth)
	router.Use(middleware.OptionalAuthMiddleware())

	// Track endpoints
	router.POST("/tracks", handlers.SaveTracks)
	router.POST("/tracks/init", handlers.InitTrackData)
	router.POST("/reset-tracks", handlers.ResetTracks)
	router.DELETE("/tracks", handlers.DeleteTracks)

	// Artist endpoints
	router.GET("/artists", handlers.GetArtists)

	// Playlist endpoints
	router.GET("/playlist", handlers.GetPlaylists)
	router.POST("/playlist", handlers.CreatePlaylist)
	router.DELETE("/playlist", handlers.DeletePlaylists)
	router.POST("/gest-playlist", handlers.GestCreatePlaylist)
}
