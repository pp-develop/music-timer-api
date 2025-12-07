package spotify

import (
	"context"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// NewClientWithUser creates a Spotify client authenticated with the user's token
func NewClientWithUser(ctx context.Context, user model.User) *spotify.Client {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		},
	))
	return spotify.New(httpClient, spotify.WithRetry(true))
}

// NewClientWithToken creates a Spotify client authenticated with an OAuth token
func NewClientWithToken(ctx context.Context, token *oauth2.Token) *spotify.Client {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	return spotify.New(httpClient, spotify.WithRetry(true))
}
