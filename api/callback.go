package api

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func Callback(c *gin.Context) bool {
	code := c.Query("code")
	state := c.Query("state")
	// api/tokenエンドポイントへhttpリクエスト
	success, response := spotify.GetApiToken(code)
	if !success {
		return false
	}

	// TODO stateの検証
	log.Println(state)

	// userid取得
	isGet, user := spotify.GetMe(response.AccessToken)
	if !isGet {
		return false
	}

	// リフレッシュトークンをDBに保存
	success = database.SaveRefreshToken(response, user.Id)
	if !success {
		return false
	}

	// sessionにuseridを格納
	session := sessions.Default(c)
	session.Set("userId", user.Id)
	session.Save()

	return true
}
