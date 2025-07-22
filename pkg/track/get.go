package track

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/json"
)

var (
	allTracks      []model.Track
	allTracksMutex sync.Mutex // 共有リソースへのアクセスを制御
	timeout        = 15       // 15秒
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
				shuffleTracks := json.ShuffleTracks(tracksToProcess)
				success, tracks = MakeTracks(shuffleTracks, specify_ms)
			}
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		log.Printf("試行回数: %d\n", tryCount) // 試行回数を出力
		return tracks, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done(): // タイムアウト時
		return nil, model.ErrTimeoutCreatePlaylist
	}
}

// MakeTracksは、指定された総再生時間を超過しないように、与えられた曲リストから曲を選択し、
// 総再生時間を計算して返します。
func MakeTracks(allTracks []model.Track, totalPlayTimeMs int) (bool, []model.Track) {
	var tracks []model.Track
	var totalDuration int

	// 全てのトラックを追加し、トラックの合計再生時間が指定された再生時間を超える場合は、ループを停止します。
	for _, v := range allTracks {
		tracks = append(tracks, v)
		totalDuration += v.DurationMs
		if totalDuration > totalPlayTimeMs {
			break
		}
	}

	// 最後に追加したトラックを削除します。
	tracks = tracks[:len(tracks)-1]

	// トラックの合計再生時間と指定された再生時間の差分を求めます。
	totalDuration = 0
	var remainingTime int
	for _, v := range tracks {
		totalDuration += v.DurationMs
	}
	remainingTime = totalPlayTimeMs - totalDuration

	// 差分が15秒以内で、指定された再生時間が10分（600秒）以上の場合、
	// 差分を埋めるためのトラックは必要ないと見なします。
	allowance_ms := 15000
	if remainingTime == allowance_ms && totalPlayTimeMs >= 600000 {
		return true, tracks
	}

	// 差分を埋めるためのトラックを取得します。
	var isTrackFound bool
	getTrack := json.GetTrackByMsec(allTracks, remainingTime)
	if len(getTrack) > 0 {
		isTrackFound = true
		tracks = append(tracks, getTrack...)
	}
	return isTrackFound, tracks
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
