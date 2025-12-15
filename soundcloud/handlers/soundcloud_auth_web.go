package handlers

import (
	"log/slog"
	"net/http"

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
// Fail-safe: セッション削除失敗時もログアウト成功として扱う（クライアント側でセッションクリア）
func DeleteSessionWeb(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Delete("service")
	if err := session.Save(); err != nil {
		slog.Error("failed to save session on delete", slog.Any("error", err))
	}
	c.Status(http.StatusOK)
}
