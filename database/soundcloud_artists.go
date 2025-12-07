package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pp-develop/music-timer-api/model"
)

// AddSoundCloudArtistTracks adds tracks for a SoundCloud artist (upsert)
func AddSoundCloudArtistTracks(db *sql.DB, id string, newTracks []model.Track) error {
	var existingTracks []model.Track

	var trackJSON *string
	err := db.QueryRow(`
        SELECT tracks FROM soundcloud_artists WHERE id = $1`, id).Scan(&trackJSON)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("error fetching artist tracks", slog.String("artist_id", id), slog.Any("error", err))
		return err
	}

	if trackJSON != nil && *trackJSON != "" {
		err = json.Unmarshal([]byte(*trackJSON), &existingTracks)
		if err != nil {
			slog.Error("error unmarshaling existing tracks", slog.String("artist_id", id), slog.Any("error", err))
			return err
		}
	}

	// Deduplicate by URI
	trackMap := make(map[string]bool)
	for _, track := range existingTracks {
		trackMap[track.Uri] = true
	}

	for _, newTrack := range newTracks {
		if !trackMap[newTrack.Uri] {
			existingTracks = append(existingTracks, newTrack)
		}
	}

	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		slog.Error("error marshaling updated tracks", slog.String("artist_id", id), slog.Any("error", err))
		return err
	}

	_, err = db.Exec(`
        INSERT INTO soundcloud_artists (id, tracks, updated_at)
        VALUES ($1, $2::jsonb, NOW())
        ON CONFLICT (id) DO UPDATE SET
            tracks = EXCLUDED.tracks,
            updated_at = NOW()`,
		id, updatedTrackJSON)
	if err != nil {
		slog.Error("error upserting artist tracks", slog.String("artist_id", id), slog.Any("error", err))
		return err
	}

	return nil
}

// GetSoundCloudTracksByArtistIds retrieves tracks for multiple artist IDs
func GetSoundCloudTracksByArtistIds(db *sql.DB, artistIDs []string) ([]model.Track, error) {
	if len(artistIDs) == 0 {
		return nil, fmt.Errorf("artist IDs array is empty")
	}

	placeholders := make([]string, len(artistIDs))
	for i := range artistIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
        SELECT tracks FROM soundcloud_artists WHERE id IN (%s)`,
		strings.Join(placeholders, ","))

	rows, err := db.Query(query, convertToInterfaceSlice(artistIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allTracks []model.Track
	for rows.Next() {
		var tracksJSON string
		if err := rows.Scan(&tracksJSON); err != nil {
			return nil, err
		}

		var tracks []model.Track
		if tracksJSON != "" {
			if err := json.Unmarshal([]byte(tracksJSON), &tracks); err != nil {
				slog.Error("error unmarshaling track JSON", slog.Any("error", err))
				return nil, err
			}
		}
		allTracks = append(allTracks, tracks...)
	}

	return allTracks, rows.Err()
}

// ClearSoundCloudArtistTracks clears tracks for an artist
func ClearSoundCloudArtistTracks(db *sql.DB, id string) error {
	_, err := db.Exec(`
        UPDATE soundcloud_artists SET tracks = '[]', updated_at = NOW()
        WHERE id = $1`, id)
	return err
}
