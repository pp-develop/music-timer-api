package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
)

// Callback handles the SoundCloud OAuth callback for both web and native authentication
// stateパラメータのプレフィックスでプラットフォームを識別し、適切なフローを実行する
// - "web_*": Webフロー → Webフロントエンドにリダイレクト
// - "native_*": Nativeフロー → ネイティブアプリにディープリンクでリダイレクト
func Callback(c *gin.Context) {
	result, err := auth.Callback(c)
	if err != nil {
		slog.Error("callback failed", slog.Any("error", err))
		// エラー時はstateからプラットフォームを判定してリダイレクト
		handleCallbackError(c)
		return
	}

	switch result.Platform {
	case auth.PlatformWeb:
		// Webフロー: Webフロントエンドにリダイレクト
		c.Redirect(http.StatusMovedPermanently, os.Getenv("SOUNDCLOUD_AUTHZ_WEB_SUCCESS_URL"))

	case auth.PlatformNative:
		// Nativeフロー: ネイティブアプリにトークン付きでリダイレクト
		// Spotifyと共通の環境変数を使用（同じネイティブアプリに戻るため）
		redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&expires_in=%d",
			os.Getenv("AUTHZ_SUCCESS_URL_NATIVE"),
			result.TokenPair.AccessToken,
			result.TokenPair.RefreshToken,
			result.TokenPair.ExpiresIn,
		)
		c.Redirect(http.StatusSeeOther, redirectURL)
	}
}

// handleCallbackError はcallbackエラー時のリダイレクトを処理する
func handleCallbackError(c *gin.Context) {
	state := c.Query("state")

	// stateからプラットフォームを判定
	if len(state) >= 4 && state[:4] == "web_" {
		c.Redirect(http.StatusSeeOther, os.Getenv("SOUNDCLOUD_AUTHZ_WEB_ERROR_URL"))
	} else if len(state) >= 7 && state[:7] == "native_" {
		// Spotifyと共通の環境変数を使用（同じネイティブアプリに戻るため）
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL_NATIVE")+"?error=auth_failed")
	} else {
		// 不明なstate: デフォルトでWebエラーURLにリダイレクト
		slog.Warn("unknown state prefix in callback error", slog.String("state", state))
		c.Redirect(http.StatusSeeOther, os.Getenv("SOUNDCLOUD_AUTHZ_WEB_ERROR_URL"))
	}
}
