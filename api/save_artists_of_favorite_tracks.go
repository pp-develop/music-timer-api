package api

import (
	spotifyApi "github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func SaveArtistsOfFavoriteTracks(token *oauth2.Token, userId string) error {

	tracks, err := spotifyApi.GetSavedTracks(token)
	if err != nil {
		return err
	}

	err = database.DeleteArtists(userId)
	if err != nil {
		return err
	}
	err = SaveArtist(tracks.Tracks, userId)
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
		err := spotifyApi.GetNextSavedTrakcs(token, tracks)
		if err != nil {
			return err
		}

		err = SaveArtist(tracks.Tracks, userId)
		if err != nil {
			return err
		}
		if tracks.Next == "" {
			existNextUrl = false
		}
	}

	return nil
}

func SaveArtist(track []spotify.SavedTrack, userId string) error {
	var artistName []string
	for _, v := range track {
		for _, a := range v.Artists {
			artistName = append(artistName, a.Name)
		}
	}

	for _, v := range artistName {
		err := database.SaveArtists(v, userId)
		if err != nil {
			return err
		}
	}
	return nil
}
