package track

import (
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetTracks(specify_ms int) ([]model.Track, error) {
	var tracks []model.Track

	c1 := make(chan []model.Track, 1)
	go func() {
		success := false
		for !success {
			tracks, _ = database.GetAllTracks()
			success, tracks = MakeTracks(tracks, specify_ms)
		}
		c1 <- tracks
	}()

	select {
	case <-c1:
		return tracks, nil
	case <-time.After(30 * time.Second):
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
	getTrack := database.GetTrackByMsec(remainingTime)
	if len(getTrack) > 0 {
		isTrackFound = true
		tracks = append(tracks, getTrack...)
	}
	return isTrackFound, tracks
}
