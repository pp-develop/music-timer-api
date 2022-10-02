package pkg

import (
	"fmt"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistModifyPublic),
		spotifyauth.WithClientID(""),
		spotifyauth.WithClientSecret(""),
	)
	state = "abc123"
)

func Authz() {
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
}
