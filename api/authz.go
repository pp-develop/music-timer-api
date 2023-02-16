package api

import (
	"fmt"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func Authz(c *gin.Context) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("SPOTIFY_REDIRECT_URI")),
		spotifyauth.WithScopes(
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserFollowRead,
		),
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
	)
	state := uuid.New()
	url := auth.AuthURL(state.String())

	// sessionにstateを格納
	session := sessions.Default(c)
	session.Set("state", state.String())
	session.Save()

	fmt.Println(url)
	return url, nil
}
