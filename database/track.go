package database

import (
	"strconv"
	"strings"

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
	query := "SELECT uri, duration_ms, artists_name FROM tracks WHERE isrc like '%JP%' LIMIT " + strconv.Itoa(offset) + "," + strconv.Itoa(limit)
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
		pageNumber++
		if len(tracks) < pageSize {
			// ページネーション処理が終了した場合はループを終える
			break
		}
	}
	return AllTracks, nil
}

func GetTracksByArtistsName(artistsName string) ([]model.Track, error) {
	var tracks []model.Track
	rows, err := db.Query("SELECT uri, duration_ms FROM tracks WHERE isrc LIKE 'JP%' AND " + artistsName + "ORDER BY rand()")
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

	if err := db.QueryRow("SELECT uri, duration_ms FROM tracks WHERE duration_ms = ? AND isrc LIKE 'JP%' ORDER BY rand()", ms).Scan(&track.Uri, &track.DurationMs); err != nil {
		return tracks
	}

	tracks = append(tracks, track)
	return tracks
}
