package playlist

import (
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track"
)

type RequestJson struct {
	Minute                 int      `json:"minute"`
	IncludeFavoriteArtists bool     `json:"includeFavoriteArtists"`
	ArtistIds              []string `json:"artistIds"`
}

func CreatePlaylist(c *gin.Context) (string, error) {
	var json RequestJson
	var err error
	if err := c.ShouldBindJSON(&json); err != nil {
		return "", err
	}
	// 1minute = 60000ms
	oneminuteToMsec := 60000
	specify_ms := json.Minute * oneminuteToMsec

	// sessionからuserIdを取得
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return "", model.ErrFailedGetSession
	}
	userId := v.(string)

	// DBからトラックを取得
	var tracks []model.Track
	if json.IncludeFavoriteArtists && len(json.ArtistIds) > 0 {
		tracks, err = track.GetSpecifyArtistsTracks(specify_ms, json.ArtistIds, userId)
		if err != nil {
			return "", err
		}
	} else {
		tracks, err = track.GetTracks(specify_ms)
		if err != nil {
			return "", err
		}
	}

	user, err := database.GetUser(userId)
	if err != nil {
		return "", err
	}

	playlist, err := spotify.CreatePlaylist(user, specify_ms)
	if err != nil {
		// 通常、エラーの種類はステータスコードで判定するのが望ましいが、
		// 現在使用しているフレームワークの制約により、エラーメッセージの文字列を判定する方法を採用している。
		if strings.Contains(err.Error(), "token expired") {
			return "", model.ErrAccessTokenExpired
		}
		return "", err
	}

	err = spotify.AddItemsPlaylist(string(playlist.ID), tracks, user)
	if err != nil {
		database.DeletePlaylists(string(playlist.ID), user.Id)
		// 通常、エラーの種類はステータスコードで判定するのが望ましいが、
		// 現在使用しているフレームワークの制約により、エラーメッセージの文字列を判定する方法を採用している。
		if strings.Contains(err.Error(), "token expired") {
			return "", model.ErrAccessTokenExpired
		}
		logger.LogError(spotify.UnfollowPlaylist(playlist.ID, user))
		return "", err
	}

	err = database.SavePlaylist(playlist, userId)
	if err != nil {
		return "", err
	}

	return string(playlist.ID), nil
}
