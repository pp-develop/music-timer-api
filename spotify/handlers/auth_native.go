package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/auth"
)

// GetAuthzURLNative returns the Spotify authorization URL for native authentication
func GetAuthzURLNative(c *gin.Context) {
	url, err := auth.SpotifyAuthzNative(c)
	if err != nil {
		slog.Error("failed to get authz URL", slog.Any("error", err))
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackNative handles the Spotify OAuth callback for native authentication
func CallbackNative(c *gin.Context) {
	tokenPair, err := auth.SpotifyCallbackNative(c)
	if err != nil {
		slog.Error("callback failed", slog.Any("error", err))
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

// GetAuthStatusNative returns the authentication status for native users
func GetAuthStatusNative(c *gin.Context) {
	user, err := auth.GetUserWithValidToken(c)

	if err != nil {
		reason := "token_expired"
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
		"country":       user.Country,
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
