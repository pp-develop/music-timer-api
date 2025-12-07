package handlers

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/auth"
)

// GetAuthzURLWeb returns the Spotify authorization URL for web authentication
func GetAuthzURLWeb(c *gin.Context) {
	url, err := auth.SpotifyAuthzWeb(c)
	if err != nil {
		slog.Error("failed to get authz URL", slog.Any("error", err))
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackWeb handles the Spotify OAuth callback for web authentication
func CallbackWeb(c *gin.Context) {
	err := auth.SpotifyCallbackWeb(c)
	if err != nil {
		slog.Error("callback failed", slog.Any("error", err))
		c.Redirect(http.StatusSeeOther, os.Getenv("SPOTIFY_AUTHZ_WEB_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("SPOTIFY_AUTHZ_WEB_SUCCESS_URL"))
	}
}

// GetAuthStatusWeb returns the authentication status for web users
func GetAuthStatusWeb(c *gin.Context) {
	user, err := auth.Auth(c)

	if err == model.ErrFailedGetSession {
		// Return unauthenticated status as a successful response
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"reason":        "session_expired",
		})
	} else if err != nil {
		slog.Error("failed to check auth status", slog.Any("error", err))
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"country":       user.Country,
		})
	}
}

// DeleteSession deletes the user's session
func DeleteSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Delete("service")
	session.Save()
	c.Status(http.StatusOK)
}
