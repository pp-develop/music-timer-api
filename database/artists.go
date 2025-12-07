package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

// GetTracksByArtistIds は、複数のアーティスト ID に基づいてトラック情報を取得します。
// すべてのトラックをまとめて返します。
func GetTracksByArtistIds(db *sql.DB, artistIDs []string) ([]model.Track, error) {
	// アーティスト ID の配列が空の場合は、エラーを返します。
	if len(artistIDs) == 0 {
		return nil, fmt.Errorf("artist IDs array is empty")
	}

	// IN クエリのプレースホルダーを作成
	placeholders := make([]string, len(artistIDs))
	for i := range artistIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
        SELECT tracks
        FROM spotify_artists
        WHERE id IN (%s)`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, convertToInterfaceSlice(artistIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allTracks []model.Track
	for rows.Next() {
		var tracksJSON string
		// 各アーティストの tracks JSON を取得
		err := rows.Scan(&tracksJSON)
		if err != nil {
			return nil, err
		}

		// トラックの JSON をデコード
		var tracks []model.Track
		if tracksJSON != "" {
			err = json.Unmarshal([]byte(tracksJSON), &tracks)
			if err != nil {
				slog.Error("error unmarshaling track JSON", slog.Any("error", err))
				return nil, err
			}
		}
		allTracks = append(allTracks, tracks...)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return allTracks, nil
}

// convertToInterfaceSlice は、string スライスをインターフェーススライスに変換します。
// db.Query などで可変長引数を使う際に利用します。
func convertToInterfaceSlice(ids []string) []interface{} {
	interfaceSlice := make([]interface{}, len(ids))
	for i, id := range ids {
		interfaceSlice[i] = id
	}
	return interfaceSlice
}

func GetArtistsUpdatedAt(db *sql.DB, id string) (time.Time, error) {
	var updatedAt time.Time
	err := db.QueryRow(`
        SELECT updated_at FROM spotify_artists WHERE id = $1`, id).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return updatedAt, nil
}

func SaveArtist(db *sql.DB, id string, track model.Track) error {
	trackJSON, err := json.Marshal([]model.Track{track}) // トラックを配列に
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO spotify_artists (id, tracks, updated_at)
        VALUES ($1, $2::jsonb, NOW())
        ON CONFLICT (id) DO UPDATE SET
            tracks = EXCLUDED.tracks,
            updated_at = NOW()`,
		id, trackJSON)
	if err != nil {
		return err
	}
	return nil
}

func GetArtistTracks(db *sql.DB, id string) ([]model.Track, error) {
	var tracksJSON string

	// データベースからJSONB形式のトラックURIリストを取得
	err := db.QueryRow(`
        SELECT tracks FROM spotify_artists WHERE id = $1`, id).Scan(&tracksJSON)
	if err != nil {
		return nil, err
	}

	// JSONB形式を[]model.Trackにデコード
	var tracks []model.Track
	err = json.Unmarshal([]byte(tracksJSON), &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func AddArtistTracks(db *sql.DB, id string, newTracks []model.Track) error {
	var existingTracks []model.Track

	// 既存のトラックURIリストを取得
	var trackJSON *string
	err := db.QueryRow(`
        SELECT tracks FROM spotify_artists WHERE id = $1`, id).Scan(&trackJSON)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("error fetching artist tracks", slog.String("artist_id", id), slog.Any("error", err))
		return err
	}

	// 既存トラックがあれば、それをパースする
	if trackJSON != nil && *trackJSON != "" {
		err = json.Unmarshal([]byte(*trackJSON), &existingTracks)
		if err != nil {
			slog.Error("error unmarshaling existing tracks", slog.String("artist_id", id), slog.Any("error", err))
			return err
		}
	}

	// 新しいトラックを既存のトラックに追加（重複を避ける）
	trackMap := make(map[string]bool)
	for _, track := range existingTracks {
		trackMap[track.Uri] = true
	}

	for _, newTrack := range newTracks {
		if !trackMap[newTrack.Uri] {
			existingTracks = append(existingTracks, newTrack)
		}
	}

	// 更新されたトラックリストをJSONBにエンコード
	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		slog.Error("error marshaling updated tracks", slog.String("artist_id", id), slog.Any("error", err))
		return err
	}

	// ON CONFLICT を使って、既存レコードがあれば更新、なければ挿入
	_, err = db.Exec(`
        INSERT INTO spotify_artists (id, tracks, updated_at)
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

func UpdateArtistsUpdateAt(db *sql.DB, id string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE spotify_artists SET updated_at = $1
        WHERE id = $2`,
		updatedAt, id)
	return err
}

func ClearArtistsTracks(db *sql.DB, id string) error {
	_, err := db.Exec(`
        UPDATE spotify_artists SET tracks = '[]', updated_at = NOW()
        WHERE id = $1`, id)
	return err
}
