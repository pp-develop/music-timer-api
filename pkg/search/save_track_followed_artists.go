package search

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/artist"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
)

const (
	maxConcurrency = 5
	timeout        = 300 * time.Second
)

var semaphore = make(chan struct{}, maxConcurrency)

func SaveTracksFromFollowedArtists(c *gin.Context) error {
	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// TODO:: 必要な関数の切り出し
	artists, err := artist.GetFollowedArtists(c)
	if err != nil {
		return err
	}

	errChan := make(chan error, len(artists))
	var wg sync.WaitGroup

	// タイムアウト付きのコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, artist := range artists {
		wg.Add(1)

		go func(artist model.Artists) {
			defer wg.Done()

			// タイムアウトを確認し、終了する
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case semaphore <- struct{}{}: // 空きがあれば取得
				defer func() { <-semaphore }() // 処理後に解放
			}

			albums, err := spotifyApi.GetArtistAlbums(artist.Id)
			if err != nil {
				// todo:: 再考慮
				errChan <- err
				return
			}

			// err = database.ClearArtistsTracks(db, artist.Id)
			// if err != nil {
			// 	log.Printf("Error clear artists table: %v", err)
			// 	errChan <- err
			// 	return
			// }

			for _, album := range albums.Albums {
				if album.ID.String() == "" {
					// todo:: 再考慮
					errChan <- fmt.Errorf("Album ID is empty")
					continue
				}

				tracks, err := spotifyApi.GetAlbumTracks(album.ID.String())
				if err != nil {
					// todo:: 再考慮
					log.Printf("Error retrieving album tracks: %v", err)
					continue
				}

				for _, item := range tracks.Tracks {
					track := convertToTrackFromSimple(&item)
					if err := database.AddArtistTrack(db, artist.Id, track); err != nil {
						// todo:: 再考慮
						log.Printf("Error add artists table: %v", err)
						errChan <- err
					}
				}
			}
		}(artist)
	}

	// Goroutineを待機し、完了後にチャンネルを閉じる
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// エラーチェック
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func convertToTrackFromSimple(savedTrack *spotify.SimpleTrack) model.Track {
	artistsId := make([]string, len(savedTrack.Artists))
	for i, artist := range savedTrack.Artists {
		artistsId[i] = artist.ID.String()
	}
	return model.Track{
		Uri:        string(savedTrack.URI),
		Isrc:       savedTrack.ExternalIDs.ISRC,
		DurationMs: int(savedTrack.Duration),
		ArtistsId:  artistsId,
	}
}
