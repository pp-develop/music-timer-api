package spotify

import (
	"context"
	"strconv"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// MillisecondsPerMinute は1分あたりのミリ秒数
const MillisecondsPerMinute = 60000

// CreatePlaylist creates a new playlist for the user with the specified duration
func CreatePlaylist(ctx context.Context, user model.User, ms int) (*spotify.FullPlaylist, error) {
	client := NewClientWithUser(ctx, user)

	playlist, err := client.CreatePlaylistForUser(ctx, user.Id, strconv.Itoa(ms/MillisecondsPerMinute)+"min", "", true, false)
	if err != nil {
		return playlist, WrapSpotifyError(err, model.ErrPlaylistCreationFailed)
	}

	return playlist, nil
}
