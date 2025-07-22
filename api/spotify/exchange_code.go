package spotify

import (
	"context"

	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func ExchangeSpotifyCode(code string, redirectURI string) (*oauth2.Token, error) {
	var token *oauth2.Token

	err := godotenv.Load()
	if err != nil {
		return token, err
	}

	ctx := context.Background()
	auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI))
	token, err = auth.Exchange(ctx, code)
	if err != nil {
		return token, err
	}

	return token, nil
}
