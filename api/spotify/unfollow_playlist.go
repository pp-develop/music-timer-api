package spotify

import (
	"context"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// UnfollowPlaylist removes a playlist from the user's library
func UnfollowPlaylist(ctx context.Context, playlistID spotify.ID, user model.User) error {
	client := NewClientWithUser(ctx, user)

	err := client.UnfollowPlaylist(ctx, playlistID)
	if err != nil {
		return WrapSpotifyError(err)
	}

	return nil
}
