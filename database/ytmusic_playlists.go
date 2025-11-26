package database

import (
	"database/sql"

	"github.com/pp-develop/music-timer-api/model"
)

// SaveYTMusicPlaylist saves a YouTube Music playlist
func SaveYTMusicPlaylist(db *sql.DB, playlistID string, userID string) error {
	_, err := db.Exec(`
        INSERT INTO ytmusic_playlists (id, user_id, created_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (id) DO NOTHING`, playlistID, userID)
	if err != nil {
		return err
	}
	return nil
}

// GetAllYTMusicPlaylists retrieves all YouTube Music playlists for a user
func GetAllYTMusicPlaylists(db *sql.DB, userID string) ([]model.Playlist, error) {
	var playlists []model.Playlist
	rows, err := db.Query("SELECT id FROM ytmusic_playlists WHERE user_id = $1", userID)
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

// DeleteYTMusicPlaylist deletes a YouTube Music playlist
func DeleteYTMusicPlaylist(db *sql.DB, playlistID string, userID string) error {
	_, err := db.Exec("DELETE FROM ytmusic_playlists WHERE id = $1 AND user_id = $2", playlistID, userID)
	if err != nil {
		return err
	}
	return nil
}
