package internal

import (
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func Authz() (bool, string) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, ""
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("REDIRECT_URI")),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistModifyPublic),
		spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
	)
	state := uuid.New()
	url := auth.AuthURL(state.String())
	return true, url
}
