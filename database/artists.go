package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
        FROM artists
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
				log.Printf("Error unmarshaling track JSON: %v", err)
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
        SELECT updated_at FROM artists WHERE id = $1`, id).Scan(&updatedAt)
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
        INSERT INTO artists (id, tracks, updated_at)
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
        SELECT tracks FROM artists WHERE id = $1`, id).Scan(&tracksJSON)
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

func AddArtistTrack(db *sql.DB, id string, newTrack model.Track) error {
	var existingTracks []model.Track

	// 既存のトラックURIリストを取得
	var trackJSON *string
	err := db.QueryRow(`
        SELECT tracks FROM artists WHERE id = $1`, id).Scan(&trackJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// レコードが存在しない場合、新規作成として処理
			return SaveArtist(db, id, newTrack)
		}
		log.Printf("Error fetching artist tracks for ID %s: %v", id, err)
		return err
	}

	if trackJSON != nil && *trackJSON != "" {
		err = json.Unmarshal([]byte(*trackJSON), &existingTracks)
		if err != nil {
			log.Printf("Error unmarshaling existing tracks for artist ID %s: %v", id, err)
			return err
		}
	}

	for _, track := range existingTracks {
		if track.Uri == newTrack.Uri {
			return fmt.Errorf("track already exists")
		}
	}

	// 新しいトラックをリストに追加
	existingTracks = append(existingTracks, newTrack)

	// 更新されたトラックリストをJSONBにエンコード
	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		log.Printf("Error marshaling updated tracks for artist ID %s: %v", id, err)
		return err
	}

	// JSONBカラムを更新
	_, err = db.Exec(`
        UPDATE artists SET tracks = $1, updated_at = NOW()
        WHERE id = $2`, updatedTrackJSON, id)
	if err != nil {
		log.Printf("Error updating artist tracks for artist ID %s: %v", id, err)
		return err
	}

	return nil
}

func UpdateArtistsUpdateAt(db *sql.DB, id string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE artists SET updated_at = $1
        WHERE id = $2`,
		updatedAt, id)
	return err
}

func ClearArtistsTracks(db *sql.DB, id string) error {
	_, err := db.Exec(`
        UPDATE artists SET tracks = '[]', updated_at = NOW()
        WHERE id = $1`, id)
	return err
}
