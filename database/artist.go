package database

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveArtists(artists []spotify.FullArtist, userId string) error {
	for _, v := range artists {
		_, err := db.Exec(`
            INSERT INTO artists (user_id, name, id)
            VALUES ($1, $2, $3)
            ON CONFLICT (id) DO NOTHING`,
			userId, v.Name, string(v.ID))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetFollowedArtists(userId string) ([]model.Artists, error) {
	var artists []model.Artists
	rows, err := db.Query(`
        SELECT id, name FROM artists
        WHERE user_id = $1`, userId)
	if err != nil {
		return artists, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist model.Artists
		if err := rows.Scan(&artist.Id, &artist.Name); err != nil {
			return artists, err
		}
		artists = append(artists, artist)
	}
	if err = rows.Err(); err != nil {
		return artists, err
	}
	return artists, nil
}

func DeleteArtists(userId string) error {
	_, err := db.Exec(`
        DELETE FROM artists WHERE user_id = $1`, userId)
	if err != nil {
		return err
	}
	return nil
}
