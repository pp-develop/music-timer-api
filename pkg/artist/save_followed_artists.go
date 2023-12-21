package artist

import (
	spotifyApi "github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)
// SaveFollowedArtists は、Spotifyユーザーがフォローしたアーティストを取得し、データベースに保存します。
func SaveFollowedArtists(token *oauth2.Token, userId string) error {

	artists, err := spotifyApi.GetFollowedArtists(token)
	if err != nil {
		return err
	}

	err = database.DeleteArtists(userId)
	if err != nil {
		return err
	}
	err = SaveArtistNames(artists.Artists, userId)
	if err != nil {
		return err
	}

	err = ProcessNextArtists(token, artists, userId)
	if err != nil {
		return err
	}

	return nil
}

func ProcessNextArtists(token *oauth2.Token, artists *spotify.FullArtistCursorPage, userId string) error {
	existAfter := artists.Cursor.After != ""

	for existAfter {
		artists, err := spotifyApi.GetAfterFollowedArtists(token, artists.Cursor.After)
		if err != nil {
			return err
		}

		err = SaveArtistNames(artists.Artists, userId)
		if err != nil {
			return err
		}
		if artists.Cursor.After == "" {
			existAfter = false
		}
	}

	return nil
}

func SaveArtistNames(artists []spotify.FullArtist, userId string) error {
	var artistName []string
	for _, v := range artists {
		artistName = append(artistName, v.Name)
	}

	for _, v := range artistName {
		err := database.SaveArtists(v, userId)
		if err != nil {
			return err
		}
	}
	return nil
}
