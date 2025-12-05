package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/middleware"
	"github.com/pp-develop/music-timer-api/router/handlers"
	soundcloudHandlers "github.com/pp-develop/music-timer-api/soundcloud/handlers"
	spotifyHandlers "github.com/pp-develop/music-timer-api/spotify/handlers"
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
				authWeb.GET("/authz-url", spotifyHandlers.GetAuthzURLWeb)
				authWeb.GET("/callback", spotifyHandlers.CallbackWeb)
				authWeb.GET("/status", spotifyHandlers.GetAuthStatusWeb)
				authWeb.DELETE("/session", spotifyHandlers.DeleteSession)
			}

			// Native authentication
			authNative := auth.Group("/native")
			{
				authNative.GET("/authz-url", spotifyHandlers.GetAuthzURLNative)
				authNative.GET("/callback", spotifyHandlers.CallbackNative)
				authNative.GET("/status", spotifyHandlers.GetAuthStatusNative)
				authNative.POST("/refresh", spotifyHandlers.RefreshTokenNative)
			}
		}

		// Protected endpoints (require authentication)
		spotify.Use(middleware.OptionalAuthMiddleware())

		// Track endpoints
		tracks := spotify.Group("/tracks")
		{
			tracks.POST("", spotifyHandlers.SaveTracks)
			tracks.POST("/reset", spotifyHandlers.ResetTracks)

			// Track initialization endpoints
			tracksInit := tracks.Group("/init")
			{
				tracksInit.POST("/favorites", spotifyHandlers.InitFavoriteTracks)
				tracksInit.POST("/followed-artists", spotifyHandlers.InitFollowedArtistsTracks)
			}
		}

		// Artist endpoints
		spotify.GET("/artists", spotifyHandlers.GetArtists)

		// Playlist endpoints
		playlists := spotify.Group("/playlists")
		{
			playlists.GET("", spotifyHandlers.GetPlaylists)
			playlists.POST("", spotifyHandlers.CreatePlaylist)
			playlists.DELETE("", spotifyHandlers.DeletePlaylists)
			playlists.POST("/guest", spotifyHandlers.GestCreatePlaylist)
		}
	}

	// SoundCloud API endpoints
	soundcloud := router.Group("/soundcloud")
	{
		// Authentication endpoints
		auth := soundcloud.Group("/auth")
		{
			// Web authentication
			authWeb := auth.Group("/web")
			{
				authWeb.GET("/authz-url", soundcloudHandlers.GetAuthzURLWeb)
				authWeb.GET("/callback", soundcloudHandlers.CallbackWeb)
				authWeb.GET("/status", soundcloudHandlers.GetAuthStatusWeb)
				authWeb.DELETE("/session", soundcloudHandlers.DeleteSessionWeb)
			}

			// Native authentication
			authNative := auth.Group("/native")
			{
				authNative.GET("/authz-url", soundcloudHandlers.GetAuthzURLNative)
				authNative.GET("/callback", soundcloudHandlers.CallbackNative)
				authNative.GET("/status", soundcloudHandlers.GetAuthStatusNative)
				authNative.POST("/refresh", soundcloudHandlers.RefreshTokenNative)
			}
		}

		// Protected endpoints (require authentication)
		soundcloud.Use(middleware.OptionalAuthMiddleware())

		// Track endpoints
		tracks := soundcloud.Group("/tracks")
		{
			tracksInit := tracks.Group("/init")
			{
				tracksInit.POST("/favorites", soundcloudHandlers.InitFavoriteTracksSoundCloud)
			}
		}

		// Artist endpoints
		soundcloud.GET("/artists", soundcloudHandlers.GetArtistsSoundCloud)

		// Playlist endpoints
		playlists := soundcloud.Group("/playlists")
		{
			playlists.GET("", soundcloudHandlers.GetPlaylistsSoundCloud)
			playlists.DELETE("", soundcloudHandlers.DeletePlaylistsSoundCloud)
			playlists.POST("/from-favorites", soundcloudHandlers.CreatePlaylistFromFavorites)
			playlists.POST("/from-artists", soundcloudHandlers.CreatePlaylistFromArtists)
		}
	}
}
