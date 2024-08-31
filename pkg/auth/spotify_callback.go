package auth

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/artist"
)

func SpotifyCallback(c *gin.Context) error {
	code := c.Query("code")
	qState := c.Query("state")
	log.Println(qState)

	session := sessions.Default(c)
	v := session.Get("state")
	if v == nil {
		return model.ErrFailedGetSession
	}
	state := v.(string)
	log.Println(state)
	if state != qState {
		return model.ErrInvalidState
	}

	token, err := spotify.ExchangeSpotifyCode(code)
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

	// フォローしてるアーティストを保存
	err = artist.SaveFollowedArtists(token, user.ID)
	if err != nil {
		return err
	}

	// sessionにuseridを格納
	session = sessions.Default(c)
	session.Set("userId", user.ID)
	session.Save()

	return nil
}
