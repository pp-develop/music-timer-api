package track

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/spotify/json"
)

// GetTracks関数は、指定された総再生時間に基づいてトラックを取得します。
func GetTracks(db *sql.DB, specify_ms int, market string) ([]model.Track, error) {
	// Phase 1: データ取得と検証（即座にエラー判定）
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(commontrack.DefaultTimeoutSeconds)*time.Second)
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

		var tracks []model.Track
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
		slog.Info("track selection completed", slog.Int("try_count", finalTryCount))
		return tracks, nil
	case err := <-errChan:
		finalTryCount := <-tryCountChan
		slog.Warn("track selection timeout", slog.Int("try_count", finalTryCount))
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
			slog.Warn("not enough tracks",
				slog.Int("required_min", specify_ms/commontrack.MillisecondsPerMinute),
				slog.Int("available_min", totalAvailableDuration/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(tracksToProcess)),
				slog.Int("try_count", finalTryCount),
				slog.String("market", market))
			return nil, model.ErrNotEnoughTracks
		} else {
			slog.Warn("combination not found",
				slog.Int("duration_min", specify_ms/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(tracksToProcess)),
				slog.Int("total_duration_min", totalAvailableDuration/commontrack.MillisecondsPerMinute),
				slog.Int("try_count", finalTryCount),
				slog.String("market", market))
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}

// filterByISRC はISRCの国コードプレフィックスに基づいてトラックをフィルタリングする
// ISRCの形式: CC-XXX-YY-NNNNN（CCが国コード、例: JP, US, GB）
// countryCode: 2文字の国コード（例: "JP"）
func filterByISRC(tracks []model.Track, countryCode string) []model.Track {
	if len(countryCode) < 2 {
		return tracks // 無効な国コードの場合は全トラックを返す
	}

	// 大文字に正規化
	countryCode = strings.ToUpper(countryCode)

	var filteredTracks []model.Track
	for _, track := range tracks {
		// ISRCの先頭2文字が国コード
		if len(track.Isrc) >= 2 && strings.HasPrefix(strings.ToUpper(track.Isrc), countryCode) {
			filteredTracks = append(filteredTracks, track)
		}
	}

	return filteredTracks
}
