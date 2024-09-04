package playlist

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/track"
	"github.com/pp-develop/music-timer-api/utils"
)

func GestCreatePlaylist(c *gin.Context) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	var json RequestJson
	if err = c.ShouldBindJSON(&json); err != nil {
		return "", err
	}
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	specify_ms := json.Minute * oneminuteToMsec

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return "", model.ErrFailedGetDB
	}

	// DBからトラックを取得
	tracks, err := track.GetTracks(dbInstance, specify_ms)
	if err != nil {
		return "", err
	}

	user, err := database.GetUser(dbInstance, os.Getenv("SPOTIFY_GEST_ACCOUNT"))
	if err != nil {
		return "", err
	}
	token, err := spotify.RefreshToken(user)
	if err != nil {
		return "", err
	}
	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.TokenExpiration = token.Expiry.Second()

	playlist, err := spotify.CreatePlaylist(user, specify_ms)
	if err != nil {
		return "", err
	}

	err = spotify.AddItemsPlaylist(string(playlist.ID), tracks, user)
	if err != nil {
		return "", err
	}

	// TODO:: delete
	err = database.SavePlaylist(dbInstance, playlist, user.Id)
	if err != nil {
		return "", err
	}

	return string(playlist.ID), nil
}
