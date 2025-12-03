package utils

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
)

// GetUserID はセッションまたはJWTコンテキストからユーザーIDを取得します
// Webベース（セッション）とネイティブ（JWT）の両方の認証をサポートします
func GetUserID(c *gin.Context) (string, error) {
	// まずJWTミドルウェアによって設定されたuserIDをチェック
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok && id != "" {
			return id, nil
		}
	}

	// Web認証のためにセッションにフォールバック
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return "", model.ErrFailedGetSession
	}

	userId, ok := v.(string)
	if !ok || userId == "" {
		return "", model.ErrFailedGetSession
	}

	return userId, nil
}

// GetService はセッションまたはJWTコンテキストからサービス名を取得します
// 戻り値: "spotify", "soundcloud" など、サービスが識別できない場合は空文字列を返します
func GetService(c *gin.Context) string {
	// まずJWTミドルウェアによって設定されたserviceをチェック
	if service, exists := c.Get("service"); exists {
		if s, ok := service.(string); ok && s != "" {
			return s
		}
	}

	// Web認証のためにセッションにフォールバック
	session := sessions.Default(c)
	if v := session.Get("service"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}

	return ""
}
