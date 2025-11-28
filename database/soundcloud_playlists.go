package database

import (
	"database/sql"
)

func SaveSoundCloudPlaylist(db *sql.DB, playlistId, userId string) error {
	_, err := db.Exec(`
        INSERT INTO soundcloud_playlists (id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (id) DO UPDATE SET
            user_id = EXCLUDED.user_id`,
		playlistId, userId)
	return err
}

func GetSoundCloudPlaylists(db *sql.DB, userId string) ([]string, error) {
	rows, err := db.Query(`
        SELECT id FROM soundcloud_playlists WHERE user_id = $1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		playlists = append(playlists, id)
	}
	return playlists, nil
}

func DeleteSoundCloudPlaylist(db *sql.DB, playlistId, userId string) error {
	_, err := db.Exec(`
        DELETE FROM soundcloud_playlists WHERE id = $1 AND user_id = $2`,
		playlistId, userId)
	return err
}

func DeleteSoundCloudPlaylists(db *sql.DB, playlistId, userId string) error {
	return DeleteSoundCloudPlaylist(db, playlistId, userId)
}
