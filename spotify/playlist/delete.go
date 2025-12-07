package playlist

import (
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/auth"
	"github.com/pp-develop/music-timer-api/utils"
	sotifySdk "github.com/zmb3/spotify/v2"
)

func DeletePlaylists(c *gin.Context) error {
	// ユーザー情報を取得（Spotifyトークンの期限切れ時は自動リフレッシュ）
	user, err := auth.GetUserWithValidToken(c)
	if err != nil {
		return err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	playlists, err := database.GetAllPlaylists(dbInstance, user.Id)
	if err != nil {
		return err
	} else if len(playlists) == 0 {
		return model.ErrNotFoundPlaylist
	}

	ctx := c.Request.Context()
	for _, item := range playlists {
		err := spotify.UnfollowPlaylist(ctx, sotifySdk.ID(item.ID), user)
		if err != nil {
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
