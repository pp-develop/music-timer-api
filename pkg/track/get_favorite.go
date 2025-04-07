package track

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/json"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

func GetFavoriteTracks(db *sql.DB, specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
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

		saveTracks, err := database.GetFavoriteTracks(db, userId)
		if err != nil {
			errChan <- err
			return
		}

		if len(saveTracks) == 0 {
			errChan <- model.ErrNotFoundTracks
			return
		}

		if len(artistIds) > 0 {
			saveTracks = filterTracksByArtistIds(saveTracks, artistIds)
		}

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
		if tracks == nil {
			return nil, <-errChan
		}
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
