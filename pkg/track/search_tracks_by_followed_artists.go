package track

import (
	"log"
	"sync"

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

func SearchTracksByFollowedArtists(userId string) error {
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

			log.Println(item.Name)
			items, err := spotify.SearchTracksByArtists(item.Name)
			if err != nil {
				errChan <- err
				return
			}
			log.Printf("%+v\n", items.Tracks)

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