package playlist

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/spotify/auth"
	"github.com/pp-develop/music-timer-api/spotify/track"
	"github.com/pp-develop/music-timer-api/utils"
)

type CreatePlaylistFromArtistsRequest struct {
	Minute    int      `json:"minute" binding:"required,min=1"`
	ArtistIds []string `json:"artistIds" binding:"required,min=1"`
}

// CreatePlaylistFromArtists creates a playlist from specified artists' tracks
func CreatePlaylistFromArtists(c *gin.Context) (string, error) {
	var json CreatePlaylistFromArtistsRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		return "", err
	}

	specifyMs := json.Minute * commontrack.MillisecondsPerMinute

	user, err := auth.GetUserWithValidToken(c)
	if err != nil {
		return "", err
	}

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return "", model.ErrFailedGetDB
	}

	tracks, err := track.GetTracksFromArtists(dbInstance, specifyMs, json.ArtistIds, user.Id)
	if err != nil {
		slog.Error("failed to get tracks from artists", slog.Any("error", err))
		return "", err
	}

	if len(tracks) == 0 {
		return "", model.ErrNotEnoughTracks
	}

	ctx := c.Request.Context()
	playlist, err := spotify.CreatePlaylist(ctx, user, specifyMs)
	if err != nil {
		return "", err
	}

	err = spotify.AddItemsPlaylist(ctx, string(playlist.ID), tracks, user)
	if err != nil {
		database.DeletePlaylists(dbInstance, string(playlist.ID), user.Id)
		if unfollowErr := spotify.UnfollowPlaylist(ctx, playlist.ID, user); unfollowErr != nil {
			slog.Error("failed to unfollow playlist", slog.Any("error", unfollowErr))
		}
		return "", err
	}

	err = database.SavePlaylist(dbInstance, playlist, user.Id)
	if err != nil {
		return "", err
	}

	if err = database.IncrementPlaylistCount(dbInstance, user.Id); err != nil {
		slog.Warn("failed to increment playlist count",
			slog.String("user_id", user.Id),
			slog.Any("error", err))
	}

	return string(playlist.ID), nil
}
