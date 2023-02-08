package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetMe(token *oauth2.Token) (*spotify.PrivateUser, error) {
	var currentUser *spotify.PrivateUser

	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		return currentUser, err
	}

	return currentUser, nil
}
