package database

import (
	"github.com/zmb3/spotify/v2"
)

func DeleteUserFavoriteArtists(userId string) error {
	_, err := db.Exec("DELETE FROM artists WHERE user_id=?", userId)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserFavoriteArtists(track []spotify.SavedTrack, userId string) error {
	var artistName []string
	for _, v := range track {
		for _, a := range v.Artists {
			artistName = append(artistName, a.Name)
		}
	}

	for _, v := range artistName {
		_, err := db.Exec("INSERT IGNORE INTO artists (user_id, name) VALUES (?, ?)", userId, v)
		if err != nil {
			return err
		}
	}

	return nil
}
