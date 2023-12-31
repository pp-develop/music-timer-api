package track

import (
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

// SearchTracksByFollowedArtists は、spotifyユーザーがフォローしたアーティストをデータベースから取得し、
// それらのアーティスト名に基づいてSpotifyでトラックを検索します。
// 検索された各アーティストについて、関連するトラックが見つかった場合、それらのトラックを保存します。
// func SearchTracksByFollowedArtists(userId string) error {
// 	artists, err := database.GetFollowedArtists(userId)
// 	log.Printf("artists:")
// 	log.Printf("%+v\n", artists)
// 	if err != nil {
// 		return err
// 	}
// 	for _, item := range artists {
// 		items, err := spotify.SearchTracksByArtists(item.Name)
// 		if err != nil {
// 			return err
// 		}

// 		err = SaveTracks(items)
// 		if err != nil {
// 			return err
// 		}

// 		if items.Tracks.Next == "" {
// 			continue
// 		}
// 		err = NextSearchTracks(items)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func SearchTracksByFollowedArtists(c *gin.Context) error {
	// sessionからuserIdを取得
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	artist, err := database.GetFollowedArtists(userId)
	if err != nil {
		return err
	}

	errChan := make(chan error, len(artist))
	var wg sync.WaitGroup

	for _, item := range artist {
		wg.Add(1)
		go func(item model.Artists) {
			defer wg.Done()

			items, err := spotify.SearchTracksByArtists(item.Name)
			if err != nil {
				errChan <- err
				return
			}

			if err := SaveTracks(items); err != nil {
				errChan <- err
				return
			}

			if items.Tracks.Next != "" {
				if err := NextSearchTracks(items); err != nil {
					errChan <- err
					return
				}
			}
		}(item)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
