package api

import (
	spotifyApi "github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func SaveUserFavoriteArtists(token *oauth2.Token, userId string) error {

	tracks, err := spotifyApi.GetUserSavesTrakcs(token)
	if err != nil {
		return err
	}

	err = database.DeleteUserFavoriteArtists(userId)
	if err != nil {
		return err
	}
	err = database.SaveUserFavoriteArtists(tracks.Tracks, userId)
	if err != nil {
		return err
	}

	err = NextPage(token, tracks, userId)
	if err != nil {
		return err
	}

	return nil
}

func NextPage(token *oauth2.Token, tracks *spotify.SavedTrackPage, userId string) error {
	existNextUrl := true

	for existNextUrl {
		err := spotifyApi.GetNextUserSavesTrakcs(token, tracks)
		if err != nil {
			return err
		}

		err = database.SaveUserFavoriteArtists(tracks.Tracks, userId)
		if err != nil {
			return err
		}
		if (tracks.Next == ""){
			existNextUrl = false
		}
	}

	return nil
}
