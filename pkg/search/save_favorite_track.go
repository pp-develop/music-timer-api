package search

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// SaveFavoriteTracks は、ユーザーの「お気に入りトラック」をデータベースに保存します。
func SaveFavoriteTracks(c *gin.Context) error {
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

	savedTracks, err := spotifyApi.GetSavedTracks(token)
	if err != nil {
		return err
	}

	err = database.ClearFavoriteTracks(db, userId)
	if err != nil {
		return err
	}

	// トラック情報を保存
	for _, item := range savedTracks {
		track := convertToTrackFromSaved(&item)
		err := database.AddFavoriteTrack(db, userId, track)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertToTrackFromSaved(savedTrack *spotify.SavedTrack) model.Track {
	artistsId := make([]string, len(savedTrack.Artists))
	for i, artist := range savedTrack.Artists {
		artistsId[i] = artist.ID.String()
	}
	return model.Track{
		Uri:        string(savedTrack.URI),
		Isrc:       savedTrack.ExternalIDs["isrc"],
		DurationMs: int(savedTrack.Duration),
		ArtistsId:  artistsId,
	}
}
