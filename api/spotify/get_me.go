package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// GetMe retrieves the current user's profile
func GetMe(ctx context.Context, token *oauth2.Token) (*spotify.PrivateUser, error) {
	client := NewClientWithToken(ctx, token)

	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		return currentUser, WrapSpotifyError(err)
	}

	return currentUser, nil
}
