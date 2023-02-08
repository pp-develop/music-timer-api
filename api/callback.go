package api

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func Callback(c *gin.Context) error {
	code := c.Query("code")
	state := c.Query("state")
	// TODO stateの検証
	log.Println(state)

	token, err := spotify.GetApiTokenForAuthzCode(code)
	if err != nil {
		return err
	}

	user, err := spotify.GetMe(token)
	if err != nil {
		return err
	}
	err = database.SaveAccessToken(token, user.ID)
	if err != nil {
		return err
	}

	// userのお気に入りのtrackを取得
	err = SaveUserFavoriteArtists(token, user.ID)
	if err != nil {
		return err
	}

	// sessionにuseridを格納
	session := sessions.Default(c)
	session.Set("userId", user.ID)
	session.Save()

	return nil
}
