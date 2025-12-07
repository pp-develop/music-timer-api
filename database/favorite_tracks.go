package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

func GetTracksUpdatedAt(db *sql.DB, userId string) (time.Time, error) {
	var updatedAt time.Time

	// データベースから updated_at を取得
	err := db.QueryRow(`
        SELECT updated_at FROM spotify_favorite_tracks WHERE user_id = $1`, userId).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err // エラーの場合はゼロ値の time.Time を返す
	}

	return updatedAt, nil
}

func SaveFavoriteTrack(db *sql.DB, userId string, track model.Track) error {
	// 配列として保存（GetFavoriteTracksが[]model.Trackを期待するため）
	favoriteTrackJSON, err := json.Marshal([]model.Track{track})
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO spotify_favorite_tracks (user_id, tracks, updated_at)
        VALUES ($1, $2::jsonb, NOW())
        ON CONFLICT (user_id) DO UPDATE SET
            tracks = EXCLUDED.tracks,
            updated_at = NOW()`,
		userId, favoriteTrackJSON)
	if err != nil {
		return err
	}
	return nil
}

func GetFavoriteTracks(db *sql.DB, userId string) ([]model.Track, error) {
	var tracksJSON string
	var tracks []model.Track

	// データベースからJSONB形式のトラックURIリストを取得
	err := db.QueryRow(`
        SELECT tracks FROM spotify_favorite_tracks WHERE user_id = $1`, userId).Scan(&tracksJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// データが存在しない場合は、空のリストを返す
			return []model.Track{}, nil
		}
		return nil, err
	}

	// JSONB形式を[]model.Trackにデコード
	err = json.Unmarshal([]byte(tracksJSON), &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func AddFavoriteTrack(db *sql.DB, userId string, newTrack model.Track) error {
	var existingTracks []model.Track

	// 既存のトラックURIリストを取得
	var trackJSON *string
	err := db.QueryRow(`
        SELECT tracks FROM spotify_favorite_tracks WHERE user_id = $1`, userId).Scan(&trackJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// レコードが存在しない場合、新規作成として処理
			return SaveFavoriteTrack(db, userId, newTrack)
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

	// 更新されたトラックリストをJSONBにエンコード
	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		return err
	}

	// JSONBカラムを更新
	_, err = db.Exec(`
        UPDATE spotify_favorite_tracks SET tracks = $1, updated_at = NOW()
        WHERE user_id = $2`, updatedTrackJSON, userId)
	if err != nil {
		return err
	}

	return nil
}

func UpdateFavoriteTracksUpdateAt(db *sql.DB, userId string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE spotify_favorite_tracks SET updated_at = $1
        WHERE user_id = $2`,
		updatedAt, userId)
	if err != nil {
		return err
	}
	return nil
}

func ClearFavoriteTracks(db *sql.DB, userId string) error {
	_, err := db.Exec(`
        DELETE FROM spotify_favorite_tracks WHERE user_id = $1`, userId)
	if err != nil {
		return err
	}
	return nil
}

func ExistsFavoriteTracks(db *sql.DB, userId string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM spotify_favorite_tracks
			WHERE user_id = $1
		)`, userId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
