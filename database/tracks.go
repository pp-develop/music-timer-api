package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// SaveTracksBatch saves multiple tracks in a single query (batch insert)
func SaveTracksBatch(db *sql.DB, tracks []spotify.FullTrack) error {
	if len(tracks) == 0 {
		return nil
	}

	// URIをキーにして重複を除去（同一バッチ内の重複でDBエラーを防ぐ）
	seen := make(map[string]bool)
	uniqueTracks := make([]spotify.FullTrack, 0, len(tracks))
	for _, track := range tracks {
		uri := string(track.URI)
		if !seen[uri] {
			seen[uri] = true
			uniqueTracks = append(uniqueTracks, track)
		}
	}

	valueStrings := make([]string, 0, len(uniqueTracks))
	valueArgs := make([]interface{}, 0, len(uniqueTracks)*3)

	for i, track := range uniqueTracks {
		offset := i * 3
		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, NOW(), NOW())", offset+1, offset+2, offset+3))
		valueArgs = append(valueArgs, track.URI, track.Duration, track.ExternalIDs["isrc"])
	}

	query := fmt.Sprintf(`
		INSERT INTO spotify_tracks (uri, duration_ms, isrc, created_at, updated_at)
		VALUES %s
		ON CONFLICT (uri) DO UPDATE SET
			duration_ms = EXCLUDED.duration_ms,
			isrc = EXCLUDED.isrc,
			updated_at = NOW()
	`, strings.Join(valueStrings, ","))

	_, err := db.Exec(query, valueArgs...)
	return err
}

// ページ番号とページサイズに基づいてトラックを取得する関数
func GetTracks(db *sql.DB, pageNumber, pageSize int) ([]model.Track, error) {
	// OFFSETを計算してLIMIT句を生成
	offset := (pageNumber - 1) * pageSize
	limit := pageSize

	// クエリ実行
	query := "SELECT uri, duration_ms, isrc FROM spotify_tracks LIMIT $1 OFFSET $2"
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]model.Track, 0)
	for rows.Next() {
		var track model.Track
		if err := rows.Scan(&track.Uri, &track.DurationMs, &track.Isrc); err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

func GetAllTracks(db *sql.DB) ([]model.Track, error) {
	var AllTracks []model.Track

	// ページネーションでトラックデータを取得
	pageSize := 50000
	pageNumber := 1
	for {
		// ページ番号とページサイズに基づいてトラックを取得
		tracks, err := GetTracks(db, pageNumber, pageSize)
		if err != nil {
			return nil, err
		}

		// 取得したトラックデータを処理
		AllTracks = append(AllTracks, tracks...)

		// ページネーション処理のために、ページ番号をインクリメントして次のページのトラックデータを取得
		if len(tracks) == 0 {
			break
		}

		// ページ番号をインクリメント
		pageNumber++
	}
	return AllTracks, nil
}

