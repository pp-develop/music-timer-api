package track

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/json"
)

func GetTracksFromArtists(db *sql.DB, specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
	// Phase 1: データ取得と検証（即座にエラー判定）
	var artists []model.Artists
	for _, id := range artistIds {
		artists = append(artists, model.Artists{Id: id})
	}

	followedArtistsTracks, err := getSpecifyArtistsAllTracks(db, artists)
	if err != nil {
		return nil, err // ErrNotFoundTracksも含む
	}

	// Phase 2: 組み合わせ計算（時間がかかる可能性がある処理）
	var tracks []model.Track
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		defer close(c1)
		defer close(errChan)

		success := false
		for !success {
			select {
			case <-ctx.Done(): // タイムアウトまたはキャンセル時にループを終了
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffleTracks := json.ShuffleTracks(followedArtistsTracks)
				success, tracks = MakeTracks(shuffleTracks, specify_ms)
			}
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		log.Printf("試行回数: %d\n", tryCount)
		return tracks, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done(): // タイムアウト時
		return nil, model.ErrTimeoutCreatePlaylist
	}
}

func getSpecifyArtistsAllTracks(db *sql.DB, artists []model.Artists) ([]model.Track, error) {
	var tracks []model.Track

	artistsIds := ConvertArtistsToIDs(artists)
	tracks, err := database.GetTracksByArtistIds(db, artistsIds)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, model.ErrNotFoundTracks
	}
	return tracks, nil
}

// Artists の各要素から ID フィールドを抽出します。
func ConvertArtistsToIDs(artists []model.Artists) []string {
	artistIDs := make([]string, len(artists)) // アーティストの数だけ string スライスを作成

	for i, artist := range artists {
		artistIDs[i] = artist.Id // 各アーティストの ID を抽出
	}

	return artistIDs
}
