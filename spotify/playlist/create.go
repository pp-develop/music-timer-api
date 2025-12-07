package playlist

import (
	"log"

	"github.com/gin-gonic/gin"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/logger"
	"github.com/pp-develop/music-timer-api/spotify/auth"
	"github.com/pp-develop/music-timer-api/spotify/track"
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
	specify_ms := json.Minute * commontrack.MillisecondsPerMinute

	// ユーザー情報を取得（Spotifyトークンの期限切れ時は自動リフレッシュ）
	user, err := auth.GetUserWithValidToken(c)
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
		tracks, err = track.GetFavoriteTracks(dbInstance, specify_ms, json.ArtistIds, user.Id)
		if err != nil {
			log.Println(err)
			return "", err
		}
	} else if len(json.ArtistIds) > 0 {
		tracks, err = track.GetTracksFromArtists(dbInstance, specify_ms, json.ArtistIds, user.Id)
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

	ctx := c.Request.Context()
	playlist, err := spotify.CreatePlaylist(ctx, user, specify_ms)
	if err != nil {
		return "", err
	}

	err = spotify.AddItemsPlaylist(ctx, string(playlist.ID), tracks, user)
	if err != nil {
		database.DeletePlaylists(dbInstance, string(playlist.ID), user.Id)
		logger.LogError(spotify.UnfollowPlaylist(ctx, playlist.ID, user))
		return "", err
	}

	err = database.SavePlaylist(dbInstance, playlist, user.Id)
	if err != nil {
		return "", err
	}

	if err = database.IncrementPlaylistCount(dbInstance, user.Id); err != nil {
		// Non-fatal: log the error but don't fail the playlist creation
		log.Printf("[WARN] Failed to increment playlist count for user %s: %v", user.Id, err)
	}

	return string(playlist.ID), nil
}
