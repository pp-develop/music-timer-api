package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/auth"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

// GetAuthzURLNative returns the Spotify authorization URL for native authentication
func GetAuthzURLNative(c *gin.Context) {
	url, err := auth.SpotifyAuthzNative(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackNative handles the Spotify OAuth callback for native authentication
func CallbackNative(c *gin.Context) {
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

// GetAuthStatusNative returns the authentication status for native users
func GetAuthStatusNative(c *gin.Context) {
	user, err := auth.Auth(c)

	if err == model.ErrFailedGetSession {
		// Return unauthenticated status as a successful response
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"reason":        "session_expired",
		})
	} else if err != nil {
		logger.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"country":       user.Country,
		})
	}
}

// RefreshTokenNative refreshes the access token for native authentication
func RefreshTokenNative(c *gin.Context) {
	tokenPair, err := auth.RefreshAccessToken(c)
	if err != nil {
		logger.LogError(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}
