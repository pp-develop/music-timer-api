package playlist

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

type RequestJson struct {
	Minute int `json:"minute"`
}

func CreatePlaylist(c *gin.Context) (string, error) {
	var json RequestJson
	if err := c.ShouldBindJSON(&json); err != nil {
		return "", err
	}
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	specify_ms := json.Minute * oneminuteToMsec

	// DBからトラックを取得
	tracks, err := track.GetTracks(specify_ms)
	if err != nil {
		return "", err
	}

	// sessionからuserIdを取得
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return "", model.ErrFailedGetSession
	}
	userId := v.(string)

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

	// oembed, err := spotify.GetOembed(string(playlist.ID))
	// if err != nil {
	// 	return "", err
	// }
	
	return string(playlist.ID), nil
}
