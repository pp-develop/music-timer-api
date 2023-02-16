package database

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func SaveFavoriteArtists(artistName string, userId string) error {
	_, err := db.Exec("INSERT IGNORE INTO artists (user_id, name) VALUES (?, ?)", userId, artistName)
	if err != nil {
		return err
	}
	return nil
}

func GetFavoriteAllArtists(userId string) ([]model.Artists, error) {
	var artists []model.Artists
	rows, err := db.Query("SELECT name FROM artists WHERE user_id = ?", userId)
	if err != nil {
		return artists, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist model.Artists
		if err := rows.Scan(&artist.Name); err != nil {
			return artists, err
		}
		artists = append(artists, artist)
	}
	if err = rows.Err(); err != nil {
		return artists, err
	}
	return artists, nil
}

func DeleteUserFavoriteArtists(userId string) error {
	_, err := db.Exec("DELETE FROM artists WHERE user_id=?", userId)
	if err != nil {
		return err
	}
	return nil
}
