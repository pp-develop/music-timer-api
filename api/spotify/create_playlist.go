package spotify

import (
	"context"
	"strconv"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func CreatePlaylist(user model.User, ms int) (*spotify.FullPlaylist, error) {
	var playlist *spotify.FullPlaylist

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		},
	))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	playlist, err := client.CreatePlaylistForUser(ctx, user.Id, strconv.Itoa(ms/ONEMINUTE_TO_MSEC)+"min", "", true, false)
	if err != nil {
		return playlist, WrapSpotifyError(err, model.ErrPlaylistCreationFailed)
	}

	return playlist, nil
}
