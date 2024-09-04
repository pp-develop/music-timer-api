package playlist

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

func GetPlaylists(c *gin.Context) ([]model.Playlist, error) {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return nil, model.ErrFailedGetSession
	}
	userId := v.(string)

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
