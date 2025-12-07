package spotify

import (
	"context"
	"strings"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// AddItemsPlaylist adds tracks to an existing playlist
func AddItemsPlaylist(ctx context.Context, playlistId string, tracks []model.Track, user model.User) error {
	client := NewClientWithUser(ctx, user)

	var ids []spotify.ID
	for _, item := range tracks {
		uri := strings.Replace(item.Uri, "spotify:track:", "", 1)
		ids = append(ids, spotify.ID(uri))
	}

	_, err := client.AddTracksToPlaylist(ctx, spotify.ID(playlistId), ids...)
	return WrapSpotifyError(err, model.ErrTrackAdditionFailed)
}
