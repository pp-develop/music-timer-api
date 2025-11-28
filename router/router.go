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
	// Health check (non-Spotify specific)
	router.GET("/health", handlers.HealthCheck)

	// Spotify API endpoints
	spotify := router.Group("/spotify")
	{
		// Authentication endpoints
		auth := spotify.Group("/auth")
		{
			// Web authentication
			authWeb := auth.Group("/web")
			{
				authWeb.GET("/authz-url", handlers.GetAuthzURLWeb)
				authWeb.GET("/callback", handlers.CallbackWeb)
				authWeb.GET("/status", handlers.GetAuthStatusWeb)
				authWeb.DELETE("/session", handlers.DeleteSession)
			}

			// Native authentication
			authNative := auth.Group("/native")
			{
				authNative.GET("/authz-url", handlers.GetAuthzURLNative)
				authNative.GET("/callback", handlers.CallbackNative)
				authNative.GET("/status", handlers.GetAuthStatusNative)
				authNative.POST("/refresh", handlers.RefreshTokenNative)
			}
		}

		// Protected endpoints (require authentication)
		spotify.Use(middleware.OptionalAuthMiddleware())

		// Track endpoints
		tracks := spotify.Group("/tracks")
		{
			tracks.POST("", handlers.SaveTracks)
			tracks.DELETE("", handlers.DeleteTracks)
			tracks.POST("/reset", handlers.ResetTracks)

			// Track initialization endpoints
			tracksInit := tracks.Group("/init")
			{
				tracksInit.POST("/favorites", handlers.InitFavoriteTracks)
				tracksInit.POST("/followed-artists", handlers.InitFollowedArtistsTracks)
			}
		}

		// Artist endpoints
		spotify.GET("/artists", handlers.GetArtists)

		// Playlist endpoints
		playlists := spotify.Group("/playlists")
		{
			playlists.GET("", handlers.GetPlaylists)
			playlists.POST("", handlers.CreatePlaylist)
			playlists.DELETE("", handlers.DeletePlaylists)
			playlists.POST("/guest", handlers.GestCreatePlaylist)
		}
	}
}
