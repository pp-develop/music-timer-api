package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func Auth(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
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
