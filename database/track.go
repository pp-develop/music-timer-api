package database

import (
	"log"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
)

func GetAllTracks() ([]model.Track, error) {
	var tracks []model.Track
	rows, err := db.Query("SELECT uri, duration_ms FROM tracks WHERE isrc like 'JP%' ORDER BY rand()")
	if err != nil {
		return tracks, err
	}
	defer rows.Close()

	for rows.Next() {
		var track model.Track
		if err := rows.Scan(&track.Uri, &track.DurationMs); err != nil {
			return tracks, err
		}
		tracks = append(tracks, track)
	}
	if err = rows.Err(); err != nil {
		return tracks, err
	}
	return tracks, nil
}

func GetTrackByMsec(ms int) []model.Track {
	var tracks []model.Track
	var track model.Track

	if err := db.QueryRow("SELECT uri, duration_ms FROM tracks WHERE duration_ms BETWEEN ?-30000 AND ?+30000 AND isrc LIKE 'JP%' ORDER BY rand()", ms, ms).Scan(&track.Uri, &track.DurationMs); err != nil {
		return tracks
	}

	tracks = append(tracks, track)
	return tracks
}

func SaveTrack(track *spotify.FullTrack) {
	_, err := db.Exec("INSERT INTO tracks (uri, duration_ms, isrc, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", track.URI, track.Duration, track.ExternalIDs["isrc"])
	if err != nil {
		log.Fatal(err)
	}
}
