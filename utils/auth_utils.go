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
