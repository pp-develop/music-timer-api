package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

// Platform はOAuth認証のプラットフォームを表す
type Platform string

const (
	PlatformWeb    Platform = "web"
	PlatformNative Platform = "native"
)

// CallbackResult は共通callbackの結果を格納する
type CallbackResult struct {
	Platform  Platform
	TokenPair *utils.TokenPair // Nativeの場合のみ設定
}

// Callback はSoundCloud OAuth callbackを処理する共通関数
// stateパラメータのプレフィックスでプラットフォームを識別する
// - "web_*": Webフロー（セッション検証、セッション保存）
// - "native_*": Nativeフロー（JWT生成）
func Callback(c *gin.Context) (*CallbackResult, error) {
	state := c.Query("state")

	platform := detectPlatform(state)

	switch platform {
	case PlatformWeb:
		// Webフロー: セッション検証 → ユーザー保存 → セッション保存
		err := CallbackWeb(c)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{Platform: PlatformWeb}, nil

	case PlatformNative:
		// Nativeフロー: JWT生成
		tokenPair, err := CallbackNative(c)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{
			Platform:  PlatformNative,
			TokenPair: tokenPair,
		}, nil

	default:
		return nil, model.ErrInvalidState
	}
}

// detectPlatform はstateパラメータからプラットフォームを判定する
func detectPlatform(state string) Platform {
	if strings.HasPrefix(state, "web_") {
		return PlatformWeb
	}
	if strings.HasPrefix(state, "native_") {
		return PlatformNative
	}
	return ""
}
