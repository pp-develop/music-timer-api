package api

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

type RequestJson struct {
	Minute int `json:"minute"`
}

func CreatePlaylistBySpecifyTime(c *gin.Context) (string, error) {
	var json RequestJson
	if err := c.ShouldBindJSON(&json); err != nil {
		return "", err
	}
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	specify_ms := json.Minute * oneminuteToMsec

	// DBからトラックリストを取得
	tracks, err := GetTracks(specify_ms)
	if err != nil {
		return "", err
	}

	// sessionからuserIdを取得
	session := sessions.Default(c)
	var userId string
	v := session.Get("userId")
	if v == nil {
		return "", errors.New("session: Failed to get userid")
	}
	userId = v.(string)

	user, err := database.GetUser(userId)
	if err != nil {
		return "", err
	}

	playlist, err := spotify.CreatePlaylist(user, specify_ms)
	if err != nil {
		return "", err
	}

	err = spotify.AddItemsPlaylist(string(playlist.ID), tracks, user)
	if err != nil {
		database.DeletePlaylists(string(playlist.ID), user.Id)
		return "", err
	}

	err = database.SavePlaylist(playlist, userId)
	if err != nil {
		return "", err
	}

	return string(playlist.ID), nil
}
