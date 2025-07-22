package playlist

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/logger"
	"github.com/pp-develop/music-timer-api/pkg/track"
	"github.com/pp-develop/music-timer-api/utils"
)

type RequestJson struct {
	Minute                int      `json:"minute"`
	Market                string   `json:"market"`
	IncludeFavoriteTracks bool     `json:"includeFavoriteTracks"`
	ArtistIds             []string `json:"artistIds"`
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

	// セッションまたはJWTからユーザーIDを取得
	userId, err := utils.GetUserID(c)
	if err != nil {
		return "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return "", model.ErrFailedGetDB
	}

	// DBからトラックを取得
	var tracks []model.Track
	if json.IncludeFavoriteTracks {
		tracks, err = track.GetFavoriteTracks(dbInstance, specify_ms, json.ArtistIds, userId)
		if err != nil {
			log.Println(err)
			return "", err
		}
	} else if len(json.ArtistIds) > 0 {
		tracks, err = track.GetTracksFromArtists(dbInstance, specify_ms, json.ArtistIds, userId)
		if err != nil {
			log.Println(err)
			return "", err
		}
	} else {
		tracks, err = track.GetTracks(dbInstance, specify_ms, json.Market)
		if err != nil {
			log.Println(err)
			return "", err
		}
	}
	
	// トラックが見つからない場合のエラー
	if len(tracks) == 0 {
		return "", model.ErrNotEnoughTracks
	}

	user, err := database.GetUser(dbInstance, userId)
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
		if strings.Contains(err.Error(), "rate limit") {
			return "", model.ErrSpotifyRateLimit
		}
		if strings.Contains(err.Error(), "quota") {
			return "", model.ErrPlaylistQuotaExceeded
		}
		return "", model.ErrPlaylistCreationFailed
	}

	err = spotify.AddItemsPlaylist(string(playlist.ID), tracks, user)
	if err != nil {
		database.DeletePlaylists(dbInstance, string(playlist.ID), user.Id)
		// 通常、エラーの種類はステータスコードで判定するのが望ましいが、
		// 現在使用しているフレームワークの制約により、エラーメッセージの文字列を判定する方法を採用している。
		if strings.Contains(err.Error(), "token expired") {
			return "", model.ErrAccessTokenExpired
		}
		if strings.Contains(err.Error(), "rate limit") {
			return "", model.ErrSpotifyRateLimit
		}
		logger.LogError(spotify.UnfollowPlaylist(playlist.ID, user))
		return "", model.ErrTrackAdditionFailed
	}

	err = database.SavePlaylist(dbInstance, playlist, userId)
	if err != nil {
		return "", err
	}

	err = database.IncrementPlaylistCount(dbInstance, userId)
	if err != nil {
		// return "", err
	}

	return string(playlist.ID), nil
}
