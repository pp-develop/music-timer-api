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
