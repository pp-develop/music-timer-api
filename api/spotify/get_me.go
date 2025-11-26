package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetMe(token *oauth2.Token) (*spotify.PrivateUser, error) {
	var currentUser *spotify.PrivateUser

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		return currentUser, WrapSpotifyError(err)
	}

	return currentUser, nil
}
