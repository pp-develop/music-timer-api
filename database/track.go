package database

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveTrack(track *spotify.FullTrack) error {
	var artistName []string
	for _, v := range track.Album.Artists {
		artistName = append(artistName, v.Name)
	}
	artistNameJson, _ := json.Marshal(artistName)

	_, err := db.Exec("INSERT INTO tracks (uri, artists_name, duration_ms, isrc, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", track.URI, artistNameJson, track.Duration, track.ExternalIDs["isrc"])
	if err != nil {
		return err
	}
	return nil
}

// ページ番号とページサイズに基づいてトラックを取得する関数
func GetTracks(pageNumber, pageSize int) ([]model.Track, error) {
	// OFFSETを計算してLIMIT句を生成
	offset := (pageNumber - 1) * pageSize
	limit := pageSize

	// クエリ実行
	query := "SELECT uri, duration_ms, artists_name FROM tracks LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]model.Track, 0)
	for rows.Next() {
		var track model.Track
		var artistsJSON string // JSON 文字列を格納するための一時変数
		if err := rows.Scan(&track.Uri, &track.DurationMs, &artistsJSON); err != nil {
			return nil, err
		}
		// JSON 文字列を構造体に変換
		if err := json.Unmarshal([]byte(artistsJSON), &track.ArtistsName); err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

func GetAllTracks() ([]model.Track, error) {
	var AllTracks []model.Track

	// ページネーションでトラックデータを取得
	pageSize := 50000
	pageNumber := 1
	for {
		// ページ番号とページサイズに基づいてトラックを取得
		tracks, err := GetTracks(pageNumber, pageSize)
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

func DeleteTracks() error {
	const chunkSize = 100000
	// 180日更新されてないデータを削除
	thirtyDaysAgo := time.Now().AddDate(0, 0, -180).Format("2006-01-02 15:04:05")
	totalDeleted := 0

	for {
		query := `DELETE FROM tracks WHERE updated_at < ? LIMIT ?`
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
