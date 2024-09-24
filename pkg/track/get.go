package track

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/json"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

var (
	allTracks      []model.Track
	allTracksMutex sync.Mutex // 共有リソースへのアクセスを制御
	timeout        = 15       // 15秒
)

// GetTracks関数は、指定された総再生時間に基づいてトラックを取得します。
func GetTracks(db *sql.DB, specify_ms int, market string) ([]model.Track, error) {
	allTracksMutex.Lock()
	localTracks := allTracks // ローカルコピーを作成
	allTracksMutex.Unlock()

	var tracks []model.Track
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel() // タイムアウト後にキャンセル

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		defer close(c1)
		defer close(errChan)

		localTracks, err = json.GetAllTracks(db)
		if err != nil {
			errChan <- err
			return
		}

		if len(localTracks) == 0 {
			// 全トラックが空の場合
			errChan <- fmt.Errorf("tracks table is empty")
			return
		}

		success := false
		for !success {
			select {
			case <-ctx.Done(): // タイムアウトまたはキャンセル時にループを終了
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffleTracks := json.ShuffleTracks(localTracks)
				if market != "" {
					shuffleTracks = filterByISRC(shuffleTracks, market)
				}
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
		if err != nil {
			logger.LogError(err)
		}
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
