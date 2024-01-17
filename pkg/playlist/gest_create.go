package playlist

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track"
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

	// DBからトラックを取得
	tracks, err := track.GetTracks(specify_ms)
	if err != nil {
		return "", err
	}

	user, err := database.GetUser(os.Getenv("SPOTIFY_GEST_ACCOUNT"))
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
	err = database.SavePlaylist(playlist, user.Id)
	if err != nil {
		return "", err
	}

	return string(playlist.ID), nil
}
