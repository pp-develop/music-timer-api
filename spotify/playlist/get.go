package playlist

import (
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

func GetPlaylists(c *gin.Context) ([]model.Playlist, error) {
	// セッションまたはJWTからユーザーIDを取得
	userId, err := utils.GetUserID(c)
	if err != nil {
		return nil, err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	playlists, err := database.GetAllPlaylists(dbInstance, userId)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}
