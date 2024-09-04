package track

import (
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

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		saveTracks, err := database.GetFavoriteTracks(db, userId)
		if err != nil {
			errChan <- err
			return
		}

		success := false
		for !success {
			tryCount++ // 試行回数をインクリメント
			shuffleTracks := json.ShuffleTracks(saveTracks)
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
