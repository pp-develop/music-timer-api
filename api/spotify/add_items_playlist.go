package spotify

import (
	"context"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func AddItemsPlaylist(playlistId string, tracks []model.Track, user model.User) error {
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		},
	))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	var ids []spotify.ID
	for _, item := range tracks {
		uri := strings.Replace(item.Uri, "spotify:track:", "", 1)
		ids = append(ids, spotify.ID(uri))
	}

	_, err := client.AddTracksToPlaylist(ctx, spotify.ID(playlistId), ids...)
	if err != nil {
		return err
	}

	return nil
}
