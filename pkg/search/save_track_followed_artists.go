package search

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/artist"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

const (
	maxConcurrency = 3
	timeout        = 360 * time.Second
)

var semaphore = make(chan struct{}, maxConcurrency)

func SaveTracksFromFollowedArtists(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	user, err := database.GetUser(db, userId)
	if err != nil {
		return err
	}
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
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

			albums, err := spotifyApi.GetArtistAlbums(token, artist.Id)
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

			var allTracks []model.Track
			if artist.Id == "3ssBfPaamcxmTrzSXcc2cb" {
				log.Println("name")
				log.Println(artist.Name)
				log.Println(len(albums))
			}

			for _, album := range albums {
				if album.ID.String() == "" {
					// todo:: 再考慮
					errChan <- fmt.Errorf("album ID is empty")
					continue
				}

				albumTracks, err := spotifyApi.GetAlbumTracks(token, album.ID.String())
				if err != nil {
					// todo:: 再考慮
					log.Printf("Error retrieving album tracks: %v", err)
					continue
				}

				for _, albumTrack := range albumTracks {
					for _, albumArtist := range albumTrack.Artists {
						if albumArtist.ID.String() == artist.Id {
							track := convertToTrackFromSimple(albumTrack)
							allTracks = append(allTracks, track)
						}
					}
				}
			}

			if artist.Id == "3ssBfPaamcxmTrzSXcc2cb" {
				log.Println("name")
				log.Println(artist.Name)
				log.Println(allTracks)
			}
			// 一度に全てのトラックを追加
			if err := database.AddArtistTracks(db, artist.Id, allTracks); err != nil {
				log.Printf("Error updating artist tracks for artist ID %s: %v", artist.Id, err)
				errChan <- err
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

func convertToTrackFromSimple(savedTrack spotify.SimpleTrack) model.Track {
	return model.Track{
		Uri:        string(savedTrack.URI),
		DurationMs: int(savedTrack.Duration),
	}
}
