package spotify

import (
	"context"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func UnfollowPlaylist(playlist []model.Playlist, user model.User) error {
	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		},
	))
	client := spotify.New(httpClient)

	for _, item := range playlist {
		err := client.UnfollowPlaylist(ctx, spotify.ID(item.ID))
		if err != nil {
			return err
		}
	}

	return nil
}
