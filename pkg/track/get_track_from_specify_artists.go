package track

import (
	"log"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/json"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
)

func GetSpecifyArtistsTracks(specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
	var tracks []model.Track
	var err error

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		var artists []model.Artists
		for _, id := range artistIds {
			artists = append(artists, model.Artists{Id: id})
		}
		followedArtistsTracks, err := getSpecifyArtistsAllTracks(artists)
		if err != nil {
			errChan <- err
			return
		}

		success := false
		for !success {
			tryCount++ // 試行回数をインクリメント
			shuffleTracks := json.ShuffleTracks(followedArtistsTracks)
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
	case <-time.After(60 * time.Second):
		if err != nil {
			logger.LogError(err)
		}
		return nil, model.ErrTimeoutCreatePlaylist
	}
}

func getSpecifyArtistsAllTracks(artists []model.Artists) ([]model.Track, error) {
	var tracks []model.Track

	tracks, err := json.GetTracksByArtistsFromAllFiles(artists)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, model.ErrNotFoundTracks
	}
	return tracks, nil
}
