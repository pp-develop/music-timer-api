package track

import (
	"log"
	"sync"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/json"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
)

var (
	allTracks      []model.Track
	allTracksMutex sync.Mutex // 共有リソースへのアクセスを制御
	timeout = 35
)

// GetTracks関数は、指定された総再生時間に基づいてトラックを取得します。
func GetTracks(specify_ms int) ([]model.Track, error) {
	allTracksMutex.Lock()
	localTracks := allTracks // ローカルコピーを作成
	allTracksMutex.Unlock()

	var tracks []model.Track
	var err error

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		localTracks, err = json.GetAllTracks()
		if err != nil {
			errChan <- err
			return
		}

		success := false
		for !success {
			tryCount++ // 試行回数をインクリメント
			shuffleTracks := json.ShuffleTracks(localTracks)
			success, tracks = MakeTracks(shuffleTracks, specify_ms)
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		if tracks == nil {
			return nil, <-errChan
		}
		log.Printf("試行回数: %d\n", tryCount) // 試行回数を出力
		return tracks, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(time.Duration(timeout) * time.Second):
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

	// 差分が15秒以内の場合、差分を埋めるためのトラックは必要ありません。
	allowance_ms := 15000
	if remainingTime == allowance_ms {
		return true, tracks
	}

	// 差分を埋めるためのトラックを取得します。
	var isTrackFound bool
	getTrack, _ := json.GetTrackByMsec(allTracks, remainingTime)
	if len(getTrack) > 0 {
		isTrackFound = true
		tracks = append(tracks, getTrack...)
	}
	return isTrackFound, tracks
}
