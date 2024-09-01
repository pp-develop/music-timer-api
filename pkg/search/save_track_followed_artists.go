package search

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

const maxConcurrency = 5

var semaphore = make(chan struct{}, maxConcurrency)

func SaveTracksByFollowedArtists(c *gin.Context) error {
	// sessionからuserIdを取得
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	artists, err := database.GetFollowedArtists(dbInstance, userId)
	if err != nil {
		return err
	}

	errChan := make(chan error, len(artists))
	var wg sync.WaitGroup

	for _, artist := range artists {
		wg.Add(1)

		go func(artist model.Artists) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			albums, err := spotify.GetArtistAlbums(artist.Id)
			if err != nil {
				// todo:: 再考慮
				errChan <- err
				return
			}

			for _, album := range albums.Albums {
				if album.ID.String() == "" {
					// todo:: 再考慮
					errChan <- fmt.Errorf("Album ID is empty")
					continue
				}

				tracks, err := spotify.GetAlbumTracks(album.ID.String())
				if err != nil {
					// todo:: 再考慮
					log.Printf("Error retrieving album tracks: %v", err)
					continue
				}

				for _, track := range tracks.Tracks {
					if err := database.SaveSimpleTrack(dbInstance, &track); err != nil {
						// todo:: 再考慮
						errChan <- err
					}
				}
			}
		}(artist)
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
