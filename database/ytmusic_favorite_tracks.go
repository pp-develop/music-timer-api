package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

// GetYTMusicTracksUpdatedAt gets the last update time of YouTube Music favorite tracks
func GetYTMusicTracksUpdatedAt(db *sql.DB, userID string) (time.Time, error) {
	var updatedAt time.Time

	err := db.QueryRow(`
        SELECT updated_at FROM ytmusic_favorite_tracks WHERE user_id = $1`, userID).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}

	return updatedAt, nil
}

// SaveYTMusicFavoriteTracks saves YouTube Music favorite tracks (overwrites existing)
func SaveYTMusicFavoriteTracks(db *sql.DB, userID string, tracks []model.YouTubeTrack) error {
	tracksJSON, err := json.Marshal(tracks)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO ytmusic_favorite_tracks (user_id, tracks, updated_at)
        VALUES ($1, $2::jsonb, NOW())
        ON CONFLICT (user_id) DO UPDATE SET
            tracks = EXCLUDED.tracks,
            updated_at = NOW()`,
		userID, tracksJSON)
	if err != nil {
		return err
	}
	return nil
}

// GetYTMusicFavoriteTracks retrieves YouTube Music favorite tracks
func GetYTMusicFavoriteTracks(db *sql.DB, userID string) ([]model.YouTubeTrack, error) {
	var tracksJSON string
	var tracks []model.YouTubeTrack

	err := db.QueryRow(`
        SELECT tracks FROM ytmusic_favorite_tracks WHERE user_id = $1`, userID).Scan(&tracksJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.YouTubeTrack{}, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(tracksJSON), &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

// AddYTMusicFavoriteTrack adds a single track to YouTube Music favorites
func AddYTMusicFavoriteTrack(db *sql.DB, userID string, newTrack model.YouTubeTrack) error {
	var existingTracks []model.YouTubeTrack

	var trackJSON *string
	err := db.QueryRow(`
        SELECT tracks FROM ytmusic_favorite_tracks WHERE user_id = $1`, userID).Scan(&trackJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// レコードが存在しない場合、新規作成
			return SaveYTMusicFavoriteTracks(db, userID, []model.YouTubeTrack{newTrack})
		}
		return err
	}

	// JSONBデータをデコード
	if trackJSON != nil && *trackJSON != "" {
		err = json.Unmarshal([]byte(*trackJSON), &existingTracks)
		if err != nil {
			return err
		}
	}

	// 新しいトラックをリストに追加
	existingTracks = append(existingTracks, newTrack)

	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        UPDATE ytmusic_favorite_tracks SET tracks = $1, updated_at = NOW()
        WHERE user_id = $2`, updatedTrackJSON, userID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateYTMusicFavoriteTracksUpdatedAt updates the last modified timestamp
func UpdateYTMusicFavoriteTracksUpdatedAt(db *sql.DB, userID string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE ytmusic_favorite_tracks SET updated_at = $1
        WHERE user_id = $2`,
		updatedAt, userID)
	if err != nil {
		return err
	}
	return nil
}

// ClearYTMusicFavoriteTracks clears all YouTube Music favorite tracks
func ClearYTMusicFavoriteTracks(db *sql.DB, userID string) error {
	_, err := db.Exec(`
        UPDATE ytmusic_favorite_tracks SET tracks = '[]', updated_at = NOW()
        WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	return nil
}
