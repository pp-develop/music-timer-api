package database

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveTrack(track *spotify.FullTrack) error {
	var artistName []string
	for _, v := range track.Album.Artists {
		artistName = append(artistName, v.Name)
	}

	_, err := db.Exec("INSERT INTO tracks (uri, artists_name, duration_ms, isrc, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", track.URI, strings.Join(artistName, " "), track.Duration, track.ExternalIDs["isrc"])
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
	query := "SELECT uri, duration_ms, artists_name FROM tracks WHERE isrc like '%JP%' LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]model.Track, 0)
	for rows.Next() {
		var track model.Track
		if err := rows.Scan(&track.Uri, &track.DurationMs, &track.ArtistsName); err != nil {
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
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Format("2000-01-01 00:00:00")

	// 30日以上更新されていないレコードを削除するSQLクエリの実行
	query := `DELETE FROM tracks WHERE updated_at < ?`
	result, err := db.Exec(query, thirtyDaysAgo)
	if err != nil {
		return err
	}

	// 削除されたレコードの数を取得
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("%d rows deleted\n", rowsAffected)
	return nil
}
