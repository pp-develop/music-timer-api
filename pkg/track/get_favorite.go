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

func GetFavoriteTracks(db *sql.DB, specify_ms int, userId string) ([]model.Track, error) {
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
