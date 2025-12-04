package track

import (
	"github.com/pp-develop/music-timer-api/model"
)

// MakeTracks は指定された総再生時間に合うようにトラックを選択する。
// 成功したかどうかと、選択されたトラックを返す。
func MakeTracks(allTracks []model.Track, totalPlayTimeMs int) (bool, []model.Track) {
	var tracks []model.Track
	var totalDuration int

	// 合計時間が指定時間を超えるまでトラックを追加
	for _, v := range allTracks {
		tracks = append(tracks, v)
		totalDuration += v.DurationMs
		if totalDuration > totalPlayTimeMs {
			break
		}
	}

	// オーバーフローを引き起こした最後のトラックを削除
	if len(tracks) > 0 {
		tracks = tracks[:len(tracks)-1]
	}

	// 残り時間を計算
	totalDuration = 0
	var remainingTime int
	for _, v := range tracks {
		totalDuration += v.DurationMs
	}
	remainingTime = totalPlayTimeMs - totalDuration

	// 残り時間が許容誤差（15秒）内かつプレイリストが十分長い（10分以上）場合、
	// ギャップを埋める必要なし。短いプレイリストでは誤差の影響が大きいため許容しない。
	// 例: 30分のプレイリストで残り10秒 → 成功（追加曲不要）
	// 例: 5分のプレイリストで残り10秒 → 追加曲を探す
	if remainingTime <= AllowanceMs && totalPlayTimeMs >= MinPlaylistDurationForAllowanceMs {
		return true, tracks
	}

	// ギャップを埋めるトラックを探す
	var isTrackFound bool
	getTrack := GetTrackByDuration(allTracks, remainingTime)
	if len(getTrack) > 0 {
		isTrackFound = true
		tracks = append(tracks, getTrack...)
	}

	return isTrackFound, tracks
}

// GetTrackByDuration は指定された再生時間に最も近い曲を許容誤差内で探す。
// durationMs ± AllowanceMs（15秒）の範囲内で、目標時間との差が最小の曲を返す。
// 例: durationMs が 20000ms の場合、5000ms〜35000ms の範囲で探索する。
func GetTrackByDuration(allTracks []model.Track, durationMs int) []model.Track {
	var bestTrack *model.Track
	bestDiff := AllowanceMs + 1 // 許容誤差を超える初期値

	for i := range allTracks {
		diff := abs(allTracks[i].DurationMs - durationMs)
		// 許容誤差内かつ、これまでで最も近い曲を選択
		if diff <= AllowanceMs && diff < bestDiff {
			bestTrack = &allTracks[i]
			bestDiff = diff
			if diff == 0 {
				break // 完全一致なら即終了
			}
		}
	}

	if bestTrack != nil {
		return []model.Track{*bestTrack}
	}
	return []model.Track{}
}

// abs は整数の絶対値を返す
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
