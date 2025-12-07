package handlers

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
)

// GetAuthzURLWeb returns the SoundCloud authorization URL for web authentication
func GetAuthzURLWeb(c *gin.Context) {
	url, err := auth.AuthzWeb(c)
	if err != nil {
		slog.Error("failed to get authz URL", slog.Any("error", err))
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackWeb handles the SoundCloud OAuth callback for web authentication
func CallbackWeb(c *gin.Context) {
	err := auth.CallbackWeb(c)
	if err != nil {
		slog.Error("callback failed", slog.Any("error", err))
		c.Redirect(http.StatusSeeOther, os.Getenv("SOUNDCLOUD_AUTHZ_WEB_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("SOUNDCLOUD_AUTHZ_WEB_SUCCESS_URL"))
	}
}

// GetAuthStatusWeb returns the authentication status for web users
func GetAuthStatusWeb(c *gin.Context) {
	user, err := auth.CheckAuth(c)

	if err != nil {
		reason := "session_expired"
		if err != model.ErrFailedGetSession {
			slog.Error("auth check failed", slog.Any("error", err))
			reason = "server_error"
		}
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"reason":        reason,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"username":      user.Username,
	})
}

// DeleteSessionWeb deletes the SoundCloud user's session for web authentication
func DeleteSessionWeb(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Delete("service")
	session.Save()
	c.Status(http.StatusOK)
}
