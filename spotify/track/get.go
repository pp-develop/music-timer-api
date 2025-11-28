package track

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/json"
)

var (
	allTracks      []model.Track
	allTracksMutex sync.Mutex // 共有リソースへのアクセスを制御
	timeout        = commontrack.DefaultTimeoutSeconds
)

// GetTracks関数は、指定された総再生時間に基づいてトラックを取得します。
func GetTracks(db *sql.DB, specify_ms int, market string) ([]model.Track, error) {
	// Phase 1: データ取得と検証（即座にエラー判定）
	allTracksMutex.Lock()
	localTracks := allTracks // ローカルコピーを作成
	allTracksMutex.Unlock()

	localTracks, err := json.GetAllTracks(db)
	if err != nil {
		return nil, err
	}

	if len(localTracks) == 0 {
		// 全トラックが空の場合
		return nil, model.ErrNotFoundTracks // 即座に返す
	}

	// Phase 2: マーケットフィルタリング（必要な場合）
	var tracksToProcess []model.Track
	if market != "" {
		tracksToProcess = filterByISRC(localTracks, market)
		if len(tracksToProcess) == 0 {
			return nil, model.ErrNotEnoughTracks // フィルタ後にトラックがない
		}
	} else {
		tracksToProcess = localTracks
	}

	// Phase 3: 組み合わせ計算（時間がかかる可能性がある処理）
	var tracks []model.Track
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCountChan := make(chan int, 1) // 試行回数を送信するチャネル
	tryCount := 0                     // 試行回数をカウントする変数

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
				shuffleTracks := json.ShuffleTracks(tracksToProcess)
				success, tracks = commontrack.MakeTracks(shuffleTracks, specify_ms)
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

		// トラックの総再生時間を計算
		totalAvailableDuration := 0
		for _, track := range tracksToProcess {
			totalAvailableDuration += track.DurationMs
		}

		hasEnoughDuration := totalAvailableDuration >= specify_ms

		if !hasEnoughDuration {
			log.Printf("[タイムアウト] GetTracks: トラック不足 - 必要=%d分, 利用可能=%d分, トラック数=%d, 試行回数=%d, マーケット=%s",
				specify_ms/commontrack.MillisecondsPerMinute,
				totalAvailableDuration/commontrack.MillisecondsPerMinute,
				len(tracksToProcess),
				finalTryCount,
				market,
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			log.Printf("[タイムアウト] GetTracks: 組み合わせ未発見 - 再生時間=%d分, トラック数=%d, 総再生時間=%d分, 試行回数=%d, マーケット=%s",
				specify_ms/commontrack.MillisecondsPerMinute,
				len(tracksToProcess),
				totalAvailableDuration/commontrack.MillisecondsPerMinute,
				finalTryCount,
				market,
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}

// 特定のISRCに基づいてトラックをフィルタリングする関数
func filterByISRC(tracks []model.Track, targetISRC string) []model.Track {
	var filteredTracks []model.Track

	for _, track := range tracks {
		if track.Isrc == targetISRC {
			filteredTracks = append(filteredTracks, track)
		}
	}

	return filteredTracks
}
