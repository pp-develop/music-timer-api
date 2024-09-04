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

func GetTracksFromArtists(db *sql.DB, specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
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

		var artists []model.Artists
		for _, id := range artistIds {
			artists = append(artists, model.Artists{Id: id})
		}

		followedArtistsTracks, err := getSpecifyArtistsAllTracks(db, artists)
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
				shuffleTracks := json.ShuffleTracks(followedArtistsTracks)
				success, tracks = MakeTracks(shuffleTracks, specify_ms)
			}
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		log.Printf("試行回数: %d\n", tryCount)
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

func getSpecifyArtistsAllTracks(db *sql.DB, artists []model.Artists) ([]model.Track, error) {
	var tracks []model.Track

	tracks, err := json.GetTracksByArtistsFromAllFiles(db, artists)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		tracks, err = database.GetTracksByArtists(db, artists)
		if err != nil {
			return nil, err
		}
	}
	if len(tracks) == 0 {
		return nil, model.ErrNotFoundTracks
	}
	return tracks, nil
}
