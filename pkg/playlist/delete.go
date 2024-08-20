package playlist

import (
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	sotifySdk "github.com/zmb3/spotify/v2"
)

func DeletePlaylists(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	playlists, err := database.GetAllPlaylists(userId)
	if err != nil {
		return err
	} else if len(playlists) == 0 {
		return model.ErrNotFoundPlaylist
	}

	user, err := database.GetUser(userId)
	if err != nil {
		return err
	}

	for _, item := range playlists {
		err := spotify.UnfollowPlaylist(sotifySdk.ID(item.ID), user)
		if err != nil {
			// 通常、エラーの種類はステータスコードで判定するのが望ましいが、
			// 現在使用しているフレームワークの制約により、エラーメッセージの文字列を判定する方法を採用している。
			if strings.Contains(err.Error(), "token expired") {
				return model.ErrAccessTokenExpired
			}
			return err
		}
	}

	for _, item := range playlists {
		err = database.DeletePlaylists(item.ID, user.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
