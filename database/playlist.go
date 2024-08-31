package database

import (
	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

func SavePlaylist(playlist *spotify.FullPlaylist, userId string) error {
	_, err := db.Exec(`
        INSERT INTO playlists (id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (id) DO NOTHING`, string(playlist.ID), userId)
	if err != nil {
		return err
	}
	return nil
}

func GetAllPlaylists(userId string) ([]model.Playlist, error) {
	var playlists []model.Playlist
	rows, err := db.Query("SELECT id FROM playlists WHERE user_id = $1", userId)
	if err != nil {
		return playlists, err
	}
	defer rows.Close()

	for rows.Next() {
		var playlist model.Playlist
		if err := rows.Scan(&playlist.ID); err != nil {
			return playlists, err
		}
		playlists = append(playlists, playlist)
	}
	if err = rows.Err(); err != nil {
		return playlists, err
	}
	return playlists, nil
}

func DeletePlaylists(playlistId string, userId string) error {
	_, err := db.Exec("DELETE FROM playlists WHERE id = $1 AND user_id = $2", playlistId, userId)
	if err != nil {
		return err
	}
	return nil
}
