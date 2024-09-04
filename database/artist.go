package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveArtists(db *sql.DB, artists []spotify.FullArtist, userId string) error {
	// トランザクションを開始
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// バルクインサートのためのクエリを構築
	queryPrefix := `
        INSERT INTO artists (user_id, name, id, image_url)
        VALUES `
	querySuffix := `
        ON CONFLICT (id) DO NOTHING`
	valueStrings := make([]string, 0, len(artists))
	valueArgs := make([]interface{}, 0, len(artists)*4)

	for i, v := range artists {
		imageUrl := ""
		if len(v.Images) > 0 {
			imageUrl = v.Images[0].URL
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, userId, v.Name, string(v.ID), imageUrl)
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

func GetFollowedArtists(db *sql.DB, userId string) ([]model.Artists, error) {
	var artists []model.Artists
	rows, err := db.Query(`
        SELECT id, name, image_url FROM artists
        WHERE user_id = $1`, userId)
	if err != nil {
		return artists, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist model.Artists
		if err := rows.Scan(&artist.Id, &artist.Name, &artist.ImageUrl); err != nil {
			return artists, err
		}
		artists = append(artists, artist)
	}
	if err = rows.Err(); err != nil {
		return artists, err
	}
	return artists, nil
}

func DeleteArtists(db *sql.DB, userId string) error {
	_, err := db.Exec(`
        DELETE FROM artists WHERE user_id = $1`, userId)
	if err != nil {
		return err
	}
	return nil
}
