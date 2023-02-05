package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
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
	} else if(len(playlists) == 0){
		return model.ErrNotFoundPlaylist
	}

	user, err := database.GetUser(userId)
	if err != nil {
		return err
	}

	err = spotify.UnfollowPlaylist(playlists, user)
	if err != nil {
		return err
	}

	for _, item := range playlists {
		err = database.DeletePlaylists(item.ID, user.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
