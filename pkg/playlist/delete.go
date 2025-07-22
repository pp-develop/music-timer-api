package playlist

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	sotifySdk "github.com/zmb3/spotify/v2"
)

func DeletePlaylists(c *gin.Context) error {
	// セッションまたはJWTからユーザーIDを取得
	userId, err := utils.GetUserID(c)
	if err != nil {
		return err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	playlists, err := database.GetAllPlaylists(dbInstance, userId)
	if err != nil {
		return err
	} else if len(playlists) == 0 {
		return model.ErrNotFoundPlaylist
	}

	user, err := database.GetUser(dbInstance, userId)
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
		err = database.DeletePlaylists(dbInstance, item.ID, user.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
