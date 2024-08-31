package spotify

import (
	"context"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func UnfollowPlaylist(playlistID spotify.ID, user model.User) error {
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		},
	))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	err := client.UnfollowPlaylist(ctx, playlistID)
	if err != nil {
		return err
	}

	return nil
}
