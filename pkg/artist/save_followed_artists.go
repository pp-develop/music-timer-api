package artist

import (
	"database/sql"

	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// SaveFollowedArtists は、Spotifyユーザーがフォローしたアーティストを取得し、データベースに保存します。
func SaveFollowedArtists(db *sql.DB, token *oauth2.Token, userId string) error {

	artists, err := spotifyApi.GetFollowedArtists(token)
	if err != nil {
		return err
	}

	err = database.DeleteArtists(db, userId)
	if err != nil {
		return err
	}
	err = saveArtists(db, artists.Artists, userId)
	if err != nil {
		return err
	}

	err = processNextArtists(db, token, artists, userId)
	if err != nil {
		return err
	}

	return nil
}

func processNextArtists(db *sql.DB, token *oauth2.Token, artists *spotify.FullArtistCursorPage, userId string) error {
	existAfter := artists.Cursor.After != ""

	for existAfter {
		artists, err := spotifyApi.GetAfterFollowedArtists(token, artists.Cursor.After)
		if err != nil {
			return err
		}

		err = saveArtists(db, artists.Artists, userId)
		if err != nil {
			return err
		}
		if artists.Cursor.After == "" {
			existAfter = false
		}
	}

	return nil
}

func saveArtists(db *sql.DB, artists []spotify.FullArtist, userId string) error {
	err := database.SaveArtists(db, artists, userId)
	if err != nil {
		return err
	}

	return nil
}
