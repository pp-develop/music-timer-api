package api

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

type RequestJson struct {
	Minute int `json:"minute"`
}

func CreatePlaylistBySpecifyTime(c *gin.Context) (bool, string) {
	var json RequestJson
	if err := c.ShouldBindJSON(&json); err != nil {
		return false, ""
	}
	specify_ms := json.Minute * ONEMINUTE_TO_MSEC

	// トラックを取得
	isGetTracks, tracks := GetTracks(specify_ms)
	if !isGetTracks {
		return false, ""
	}

	// sessionからuserIdを取得
	session := sessions.Default(c)
	var userId string
	v := session.Get("userId")
	if v == nil {
		return false, ""
	}
	userId = v.(string)

	// userIdからtokenを取得
	user := database.GetUser(userId)

	// プレイリスト作成
	isCreate, playlist := CreatePlaylist(user.Id, specify_ms, user.AccessToken)
	if !isCreate {
		return false, ""
	}

	// プレイリストにトラックを追加
	isAddItems := AddItemsPlaylist(playlist.ID, tracks, user.AccessToken)
	if !isAddItems {
		// TODO 作成したプレイリストを削除
		return false, playlist.ID
	}
	return isAddItems, playlist.ID
}
