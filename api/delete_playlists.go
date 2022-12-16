package api

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func DeletePlaylists(c *gin.Context) error {
	session := sessions.Default(c)
	var userId string
	v := session.Get("userId")
	if v == nil {
		return errors.New("session: Failed to get userid")
	}
	userId = v.(string)

	playlists, err := database.GetAllPlaylists(userId)
	if err != nil {
		return err
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
