package track

import (
	"log"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/json"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/logger"
)

func GetFollowedArtistsTracks(specify_ms int, userId string) ([]model.Track, error) {
	var tracks []model.Track
	var err error

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCount := 0 // 試行回数をカウントする変数

	go func() {
		followedArtistsTracks, err := getFollowedArtistsAllTracks(userId)
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

func getFollowedArtistsAllTracks(userId string) ([]model.Track, error) {
	var tracks []model.Track

	followedArtists, err := database.GetFollowedArtists(userId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	allTracks, err = json.GetAllTracks()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tracks, err = json.GetFollowedArtistsAllTracks(allTracks, followedArtists)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return tracks, nil
}
