package auth

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/artist"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func SpotifyCallback(c *gin.Context) error {
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

	// フォローしてるアーティストを保存
	err = artist.SaveFollowedArtists(token, user.ID)
	if err != nil {
		return err
	}
	// err = SaveArtistsOfFavoriteTracks(token, user.ID)
	// if err != nil {
	// 	return err
	// }

	// sessionにuseridを格納
	session := sessions.Default(c)
	session.Set("userId", user.ID)
	session.Save()

	return nil
}
