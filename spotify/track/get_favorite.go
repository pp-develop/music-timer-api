package track

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/spotify/json"
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

		success := false
		for !success {
			select {
			case <-ctx.Done(): // タイムアウトまたはキャンセル時にループを終了
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffleTracks := json.ShuffleTracks(saveTracks)
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
		for _, track := range saveTracks {
			totalAvailableDuration += track.DurationMs
		}

		hasEnoughDuration := totalAvailableDuration >= specify_ms

		if !hasEnoughDuration {
			slog.Warn("not enough favorite tracks",
				slog.Int("required_min", specify_ms/commontrack.MillisecondsPerMinute),
				slog.Int("available_min", totalAvailableDuration/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(saveTracks)),
				slog.Int("try_count", finalTryCount),
				slog.Int("artist_count", len(artistIds)))
			return nil, model.ErrNotEnoughTracks
		} else {
			slog.Warn("favorite tracks combination not found",
				slog.Int("duration_min", specify_ms/commontrack.MillisecondsPerMinute),
				slog.Int("track_count", len(saveTracks)),
				slog.Int("total_duration_min", totalAvailableDuration/commontrack.MillisecondsPerMinute),
				slog.Int("try_count", finalTryCount),
				slog.Int("artist_count", len(artistIds)))
			return nil, model.ErrTimeoutCreatePlaylist
		}
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
