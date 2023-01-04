package spotify

import (
	"context"
	"os"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"github.com/joho/godotenv"
)

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func GetApiTokenForAuthzCode(code string) (*oauth2.Token, error) {
	var token *oauth2.Token

	err := godotenv.Load()
	if err != nil {
		return token, err
	}

	ctx := context.Background()
	auth := spotifyauth.New(spotifyauth.WithRedirectURL(os.Getenv("SPOTIFY_REDIRECT_URI")))
	token, err = auth.Exchange(ctx, code)
	if err != nil {
		return token, err
	}

	return token, nil
}
