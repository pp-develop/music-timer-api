package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

type RequestJson struct {
	Minute int `json:"minute"`
}

func CreatePlaylistBySpecifyTime(c *gin.Context) (bool, string) {
	var json RequestJson
	if err := c.ShouldBindJSON(&json); err != nil {
		return false, ""
	}
	specify_ms := json.Minute * ONEMINUTE_TO_MSEC

	// DBからトラックリストを取得
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

	user := database.GetUser(userId)

	isCreate, playlist := spotify.CreatePlaylist(user.Id, specify_ms, user.AccessToken)
	if !isCreate {
		return false, ""
	}

	isAddItems := spotify.AddItemsPlaylist(playlist.ID, tracks, user.AccessToken)
	if !isAddItems {
		// TODO 作成したプレイリストを削除
		return false, playlist.ID
	}
	database.SavePlaylist(playlist, userId)
	return true, playlist.ID
}
