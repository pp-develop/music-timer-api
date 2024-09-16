package search

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// SaveFavoriteTracks は、ユーザーの「お気に入りトラック」をデータベースに保存します。
func SaveFavoriteTracks(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	user, err := database.GetUser(db, userId)
	if err != nil {
		return err
	}

	// 24時間以上経過している場合はトークンを取得して処理を続行
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}

	tracks, err := spotifyApi.GetSavedTracks(token)
	if err != nil {
		return err
	}

	err = database.ClearFavoriteTracks(db, userId)
	if err != nil {
		return err
	}

	// トラック情報を保存
	for _, item := range tracks.Tracks {
		track := convertToTrackModel(&item)
		err := database.AddFavoriteTrack(db, userId, track)
		if err != nil {
			return err
		}
	}

	// 次のトラックが存在する場合の処理
	err = ProcessNextTracks(db, token, tracks, userId)
	if err != nil {
		return err
	}

	// 更新日時を現在の時間に更新
	err = database.UpdateFavoriteTracksUpdateAt(db, userId, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func ProcessNextTracks(db *sql.DB, token *oauth2.Token, tracks *spotify.SavedTrackPage, userId string) error {
	existNextUrl := true

	for existNextUrl {
		err := spotifyApi.GetNextSavedTrakcs(token, tracks)
		if err != nil {
			return err
		}

		for _, item := range tracks.Tracks {
			track := convertToTrackModel(&item)
			err := database.AddFavoriteTrack(db, userId, track)
			if err != nil {
				return err
			}
		}

		if tracks.Next == "" {
			existNextUrl = false
		}
	}

	return nil
}

func convertToTrackModel(savedTrack *spotify.SavedTrack) model.Track {
	artistsId := make([]string, len(savedTrack.Artists))
	for i, artist := range savedTrack.Artists {
		artistsId[i] = artist.ID.String()
	}
	return model.Track{
		Uri:        string(savedTrack.URI),
		DurationMs: int(savedTrack.Duration),
		ArtistsId:  artistsId,
	}
}
