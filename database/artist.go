package database

import (
	"fmt"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveArtists(artists []spotify.FullArtist, userId string) error {
	// トランザクションを開始
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// バルクインサートのためのクエリを構築
	queryPrefix := `
        INSERT INTO artists (user_id, name, id)
        VALUES `
	querySuffix := `
        ON CONFLICT (id) DO NOTHING`
	valueStrings := make([]string, 0, len(artists))
	valueArgs := make([]interface{}, 0, len(artists)*3)

	for i, v := range artists {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, userId, v.Name, string(v.ID))
	}

	query := queryPrefix + strings.Join(valueStrings, ",") + querySuffix

	_, err = tx.Exec(query, valueArgs...)
	if err != nil {
		tx.Rollback()
		return err
	}

	// トランザクションをコミット
	err = tx.Commit()
	if err != nil {
		return err
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
