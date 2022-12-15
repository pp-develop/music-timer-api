package database

import (
	"log"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func SavePlaylist(playlist model.CreatePlaylistResponse, userId string) {
	_, err := db.Exec("INSERT INTO playlists (id, user_id) VALUES (?, ?)", playlist.ID, userId)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllPlaylists(userId string) ([]model.Playlist, error) {
	var playlists []model.Playlist
	rows, err := db.Query("SELECT id FROM playlists WHERE user_id = ?", userId)
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
	_, err := db.Exec("DELETE FROM playlists WHERE id=? AND user_id=?", playlistId, userId)
	if err != nil {
		return err
	}
	return nil
}
