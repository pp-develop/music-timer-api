package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

func SaveSoundCloudFavoriteTracks(db *sql.DB, userId string, tracks []model.Track) error {
	favoriteTracksJSON, err := json.Marshal(tracks)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO soundcloud_favorite_tracks (user_id, tracks, updated_at)
        VALUES ($1, $2::jsonb, NOW())
        ON CONFLICT (user_id) DO UPDATE SET
            tracks = EXCLUDED.tracks,
            updated_at = NOW()`,
		userId, favoriteTracksJSON)
	if err != nil {
		return err
	}
	return nil
}

func GetSoundCloudFavoriteTracks(db *sql.DB, userId string) ([]model.Track, error) {
	var tracksJSON string
	var tracks []model.Track

	err := db.QueryRow(`
        SELECT tracks FROM soundcloud_favorite_tracks WHERE user_id = $1`, userId).Scan(&tracksJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.Track{}, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(tracksJSON), &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func UpdateSoundCloudFavoriteTracksUpdatedAt(db *sql.DB, userId string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE soundcloud_favorite_tracks SET updated_at = $1
        WHERE user_id = $2`,
		updatedAt, userId)
	if err != nil {
		return err
	}
	return nil
}

func ClearSoundCloudFavoriteTracks(db *sql.DB, userId string) error {
	_, err := db.Exec(`
        UPDATE soundcloud_favorite_tracks SET tracks = '[]', updated_at = NOW()
        WHERE user_id = $1`, userId)
	if err != nil {
		return err
	}
	return nil
}
