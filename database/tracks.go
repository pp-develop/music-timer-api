package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveTrack(db *sql.DB, track *spotify.FullTrack) error {
	_, err := db.Exec(`
        INSERT INTO tracks (uri, duration_ms, isrc, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (uri) DO UPDATE SET 
            duration_ms = EXCLUDED.duration_ms,
            isrc = EXCLUDED.isrc,
            updated_at = NOW()`, track.URI, track.Duration, track.ExternalIDs["isrc"])
	if err != nil {
		return err
	}
	return nil
}

func SaveSimpleTrack(db *sql.DB, track *spotify.SimpleTrack) error {
	_, err := db.Exec(`
        INSERT INTO tracks (uri, duration_ms, isrc, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (uri) DO UPDATE SET 
            duration_ms = EXCLUDED.duration_ms,
            isrc = EXCLUDED.isrc,
            updated_at = NOW()`, track.URI, track.Duration, track.ExternalIDs.ISRC)
	if err != nil {
		return err
	}
	return nil
}

// ページ番号とページサイズに基づいてトラックを取得する関数
func GetTracks(db *sql.DB, pageNumber, pageSize int) ([]model.Track, error) {
	// OFFSETを計算してLIMIT句を生成
	offset := (pageNumber - 1) * pageSize
	limit := pageSize

	// クエリ実行
	query := "SELECT uri, duration_ms, isrc FROM tracks LIMIT $1 OFFSET $2"
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

func DeleteTracks(db *sql.DB) error {
	const chunkSize = 100000
	// 180日更新されてないデータを削除
	thirtyDaysAgo := time.Now().AddDate(0, 0, -180).Format("2006-01-02 15:04:05")
	totalDeleted := 0

	for {
		query := `DELETE FROM tracks WHERE updated_at < $1 LIMIT $2`
		result, err := db.Exec(query, thirtyDaysAgo, chunkSize)
		if err != nil {
			return err
		}

		// このバッチで削除された行数を取得
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		totalDeleted += int(rowsAffected)

		// もう削除すべき行がない場合、ループを抜ける
		if rowsAffected == 0 {
			break
		}
	}

	log.Printf("Total %d rows deleted\n", totalDeleted)
	return nil
}

func DeleteOldTracksIfOverLimit(db *sql.DB) error {
	const maxRows = 100000    // 10万行の上限
	const deleteChunk = 10000 // 一度に削除する行数

	// トラック数を取得するクエリ
	var rowCount int
	err := db.QueryRow("SELECT COUNT(*) FROM tracks").Scan(&rowCount)
	if err != nil {
		return err
	}

	// トラック数が10万行を超えているかチェック
	if rowCount > maxRows {
		log.Printf("Track count exceeds %d, proceeding with deletion of %d rows.\n", maxRows, deleteChunk)

		// 古いデータから1万行削除
		query := `
            DELETE FROM tracks
            WHERE ctid IN (
                SELECT ctid
                FROM tracks
                ORDER BY updated_at ASC
                LIMIT $1
            )
        `
		result, err := db.Exec(query, deleteChunk)
		if err != nil {
			return err
		}

		// 削除された行数を取得
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		log.Printf("%d rows deleted.\n", rowsAffected)
	} else {
		log.Printf("Track count is below the limit of %d rows, no deletion needed.\n", maxRows)
	}

	return nil
}
