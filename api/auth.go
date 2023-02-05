package api

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func Auth(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return errors.New("session: Failed to get userid")
	}
	userId := v.(string)

	user, err := database.GetUser(userId)
	if err != nil {
		return err
	}

	token, err := spotify.RefreshToken(user)
	if err != nil {
		return err
	}

	err = database.UpdateAccessToken(token, user.Id)
	if err != nil {
		return err
	}

	session.Set("userId", user.Id)
	session.Save()

	return nil
}
