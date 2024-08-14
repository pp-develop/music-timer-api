package search

import (
	spotifyApi "github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// SaveFavoriteTracks は、ユーザーの「お気に入りトラック」をデータベースに保存します。
func SaveFavoriteTracks(token *oauth2.Token, userId string) error {

	tracks, err := spotifyApi.GetSavedTracks(token)
	if err != nil {
		return err
	}

	err = database.ClearFavoriteTracks(userId)
	if err != nil {
		return err
	}

	// トラック情報を保存
	for _, item := range tracks.Tracks {
		track := convertToTrackModel(&item)
		err := database.AddFavoriteTrack(userId, track)
		if err != nil {
			return err
		}
	}

	// 次のトラックが存在する場合の処理
	err = ProcessNextTracks(token, tracks, userId)
	if err != nil {
		return err
	}

	return nil
}

func ProcessNextTracks(token *oauth2.Token, tracks *spotify.SavedTrackPage, userId string) error {
	existNextUrl := true

	for existNextUrl {
		err := spotifyApi.GetNextSavedTrakcs(token, tracks)
		if err != nil {
			return err
		}

		for _, item := range tracks.Tracks {
			track := convertToTrackModel(&item)
			err := database.AddFavoriteTrack(userId, track)
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
	return model.Track{
		Uri:        string(savedTrack.URI),
		DurationMs: int(savedTrack.Duration),
	}
}
