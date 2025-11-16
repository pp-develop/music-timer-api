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

func GetFavoriteTracks(db *sql.DB, specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
	// Phase 1: データ取得と検証（即座にエラー判定）
	saveTracks, err := database.GetFavoriteTracks(db, userId)
	if err != nil {
		return nil, err
	}

	if len(saveTracks) == 0 {
		return nil, model.ErrNoFavoriteTracks // 即座に返す
	}

	// Phase 2: フィルタリング（即座にエラー判定）
	if len(artistIds) > 0 {
		saveTracks = filterTracksByArtistIds(saveTracks, artistIds)
		if len(saveTracks) == 0 {
			return nil, model.ErrNotEnoughTracks // 即座に返す
		}
	}

	// Phase 3: 組み合わせ計算（時間がかかる可能性がある処理）
	var tracks []model.Track
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCountChan := make(chan int, 1) // 試行回数を送信するチャネル
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		defer close(c1)
		defer close(errChan)
		defer func() {
			tryCountChan <- tryCount // goroutine終了時に試行回数を送信
			close(tryCountChan)
		}()

		success := false
		for !success {
			select {
			case <-ctx.Done(): // タイムアウトまたはキャンセル時にループを終了
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffleTracks := json.ShuffleTracks(saveTracks)
				success, tracks = MakeTracks(shuffleTracks, specify_ms)
			}
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		finalTryCount := <-tryCountChan
		log.Printf("試行回数: %d\n", finalTryCount) // 試行回数を出力
		return tracks, nil
	case err := <-errChan:
		finalTryCount := <-tryCountChan
		log.Printf("タイムアウト: 試行回数: %d\n", finalTryCount)
		return nil, err
	case <-ctx.Done(): // タイムアウト時
		finalTryCount := <-tryCountChan
		// タイムアウト時の詳細情報をログ出力
		log.Printf("[タイムアウト詳細] 関数: GetFavoriteTracks, 再生時間: %d分, 利用可能トラック数: %d, 試行回数: %d, タイムアウト: %d秒, お気に入り: true, アーティスト数: %d",
			specify_ms/60000,
			len(saveTracks),
			finalTryCount,
			timeout,
			len(artistIds),
		)
		return nil, model.ErrTimeoutCreatePlaylist
	}
}

// 関数: 特定のアーティストIDを含むトラックをフィルタリング
func filterTracksByArtistIds(tracks []model.Track, artistIds []string) []model.Track {
	var filteredTracks []model.Track

	// トラックのループ
	for _, track := range tracks {
		// アーティストIDのチェック
		if containsAny(track.ArtistsId, artistIds) {
			filteredTracks = append(filteredTracks, track)
		}
	}

	return filteredTracks
}

// 特定のアーティストIDがトラックのArtistsIdに含まれているかチェックする関数
func containsAny(trackArtistIds []string, artistIds []string) bool {
	// アーティストIDのセットを作成して効率化
	artistIdSet := make(map[string]struct{})
	for _, id := range artistIds {
		artistIdSet[id] = struct{}{}
	}

	// トラックのアーティストIDに特定のIDが含まれているかチェック
	for _, id := range trackArtistIds {
		if _, exists := artistIdSet[id]; exists {
			return true
		}
	}

	return false
}
