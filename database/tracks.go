package database

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

func SaveTrack(db *sql.DB, track *spotify.FullTrack) error {
	var artistId []string
	var artistName []string
	for _, v := range track.Album.Artists {
		artistId = append(artistId, string(v.ID))
		artistName = append(artistName, v.Name)
	}
	artistIdJson, _ := json.Marshal(artistId)
	artistNameJson, _ := json.Marshal(artistName)

	_, err := db.Exec(`
        INSERT INTO tracks (uri, artists_id, artists_name, duration_ms, isrc, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (uri) DO UPDATE SET 
            artists_id = EXCLUDED.artists_id,
            artists_name = EXCLUDED.artists_name,
            duration_ms = EXCLUDED.duration_ms,
            isrc = EXCLUDED.isrc,
            updated_at = NOW()`, track.URI, artistIdJson, artistNameJson, track.Duration, track.ExternalIDs["isrc"])
	if err != nil {
		return err
	}
	return nil
}

func SaveSimpleTrack(db *sql.DB, track *spotify.SimpleTrack) error {
	var artistId []string
	var artistName []string
	for _, v := range track.Artists {
		artistId = append(artistId, string(v.ID))
		artistName = append(artistName, v.Name)
	}
	artistIdJson, _ := json.Marshal(artistId)
	artistNameJson, _ := json.Marshal(artistName)

	_, err := db.Exec(`
        INSERT INTO tracks (uri, artists_id, artists_name, duration_ms, isrc, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (uri) DO UPDATE SET 
            artists_id = EXCLUDED.artists_id,
            artists_name = EXCLUDED.artists_name,
            duration_ms = EXCLUDED.duration_ms,
            isrc = EXCLUDED.isrc,
            updated_at = NOW()`, track.URI, artistIdJson, artistNameJson, track.Duration, track.ExternalIDs.ISRC)
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
	query := "SELECT uri, duration_ms, isrc, artists_name, artists_id FROM tracks LIMIT $1 OFFSET $2"
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]model.Track, 0)
	for rows.Next() {
		var track model.Track
		var artistsNameJSON string // JSON 文字列を格納するための一時変数
		var artistsIdJSON string   // JSON 文字列を格納するための一時変数
		if err := rows.Scan(&track.Uri, &track.DurationMs, &track.Isrc, &artistsNameJSON, &artistsIdJSON); err != nil {
			return nil, err
		}
		// JSON 文字列を構造体に変換
		if err := json.Unmarshal([]byte(artistsNameJSON), &track.ArtistsName); err != nil {
			return nil, err
		}
		// artistsIdJSONのデシリアライズを追加
		if err := json.Unmarshal([]byte(artistsIdJSON), &track.ArtistsId); err != nil {
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

func GetTracksByArtists(db *sql.DB, artists []model.Artists) ([]model.Track, error) {
	// Extract artist IDs from the artists slice
	var artistIds []string
	for _, artist := range artists {
		artistIds = append(artistIds, artist.Id)
	}

	// Convert artistIds slice to a JSON string for the SQL query
	artistIdsJson, err := json.Marshal(artistIds)
	if err != nil {
		return nil, err
	}

	// Query to select tracks where the artists_id JSON contains any of the provided artist IDs
	query := `SELECT uri, duration_ms, isrc, artists_name, artists_id FROM tracks WHERE artists_id @> $1`

	// Executing the query
	rows, err := db.Query(query, artistIdsJson)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []model.Track
	for rows.Next() {
		var track model.Track
		var artistsNameJSON, artistsIdJSON string

		if err := rows.Scan(&track.Uri, &track.DurationMs, &track.Isrc, &artistsNameJSON, &artistsIdJSON); err != nil {
			return nil, err
		}

		// Deserialize JSON fields into the respective struct fields
		if err := json.Unmarshal([]byte(artistsNameJSON), &track.ArtistsName); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(artistsIdJSON), &track.ArtistsId); err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}
