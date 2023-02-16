package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func CreatePlaylistWithFollowedArtists(c *gin.Context) (string, error) {
	var json RequestJson
	if err := c.ShouldBindJSON(&json); err != nil {
		return "", err
	}
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	specifyMsec := json.Minute * oneminuteToMsec

	// sessionからuserIdを取得
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return "", model.ErrFailedGetSession
	}
	userId := v.(string)

	// DBからトラックを取得
	tracks, err := GetFollowedArtistsTracksBySpecifyTime(specifyMsec, userId)
	if err != nil {
		return "", err
	}

	user, err := database.GetUser(userId)
	if err != nil {
		return "", err
	}

	playlist, err := spotify.CreatePlaylist(user, specifyMsec)
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
