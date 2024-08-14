package playlist

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetPlaylists(c *gin.Context) ([]model.Playlist, error) {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return nil, model.ErrFailedGetSession
	}
	userId := v.(string)

	playlists, err := database.GetAllPlaylists(userId)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}
