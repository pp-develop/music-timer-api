package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetFollowedArtists(token *oauth2.Token) (*spotify.FullArtistCursorPage, error) {
	var artists *spotify.FullArtistCursorPage

	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	artists, err := client.CurrentUsersFollowedArtists(ctx, spotify.Limit(50))
	if err != nil {
		return artists, err
	}

	return artists, nil
}

func GetAfterFollowedArtists(token *oauth2.Token, after string) (*spotify.FullArtistCursorPage, error) {
	var artists *spotify.FullArtistCursorPage

	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	artists, err := client.CurrentUsersFollowedArtists(ctx, spotify.Limit(50), spotify.After(after))
	if err != nil {
		return artists, err
	}

	return artists, nil
}
