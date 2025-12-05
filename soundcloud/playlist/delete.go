package playlist

import (
	"github.com/gin-gonic/gin"
	soundcloud "github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/utils"
)

// DeletePlaylists deletes all SoundCloud playlists for the user
func DeletePlaylists(c *gin.Context) error {
	// ユーザー情報を取得（SoundCloudトークンの期限切れ時は自動リフレッシュ）
	user, err := auth.GetAuth(c)
	if err != nil {
		return err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// すべてのプレイリストを取得
	playlists, err := database.GetSoundCloudPlaylists(dbInstance, user.Id)
	if err != nil {
		return err
	} else if len(playlists) == 0 {
		return model.ErrNotFoundPlaylist
	}

	client := soundcloud.NewClient()

	// SoundCloud APIで削除
	for _, playlistID := range playlists {
		err := client.DeletePlaylist(user.AccessToken, playlistID)
		if err != nil {
			return err
		}
	}

	// データベースから削除
	for _, playlistID := range playlists {
		err = database.DeleteSoundCloudPlaylist(dbInstance, playlistID, user.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
