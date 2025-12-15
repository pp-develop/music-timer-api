package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
)

// GetAuthzURLNative returns the SoundCloud authorization URL for native authentication
func GetAuthzURLNative(c *gin.Context) {
	url, err := auth.AuthzNative(c)
	if err != nil {
		slog.Error("failed to get authz URL", slog.Any("error", err))
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// GetAuthStatusNative returns the authentication status for native users
func GetAuthStatusNative(c *gin.Context) {
	user, err := auth.CheckAuth(c)

	if err != nil {
		reason := "session_expired"
		if err != model.ErrFailedGetSession {
			// セッション取得以外のエラー（サーバーエラー）
			slog.Error("auth check failed", slog.Any("error", err))
			reason = "server_error"
		} else {
			// トークン期限切れまたはセッションが見つからない（通常の動作）
			slog.Info("token expired or session not found", slog.Any("error", err))
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

// RefreshTokenNative refreshes the access token for native authentication
func RefreshTokenNative(c *gin.Context) {
	tokenPair, err := auth.RefreshAccessToken(c)
	if err != nil {
		slog.Error("failed to refresh token", slog.Any("error", err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}
