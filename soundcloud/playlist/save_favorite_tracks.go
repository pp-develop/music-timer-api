package playlist

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/utils"
)

// SaveFavoriteTracks saves user's favorite tracks from SoundCloud to database
func SaveFavoriteTracks(c *gin.Context) error {
	// Get authenticated user
	user, err := auth.GetAuth(c)
	if err != nil {
		return err
	}

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// Get favorites from SoundCloud API
	client := soundcloud.NewClient()
	tracks, err := client.GetFavorites(user.AccessToken)
	if err != nil {
		return err
	}

	log.Printf("[SAVE-FAVORITE-TRACKS] Retrieved %d favorite tracks from SoundCloud", len(tracks))

	// Clear existing favorites and save new ones
	err = database.ClearSoundCloudFavoriteTracks(db, user.Id)
	if err != nil {
		return err
	}

	err = database.SaveSoundCloudFavoriteTracks(db, user.Id, tracks)
	if err != nil {
		return err
	}

	return nil
}
