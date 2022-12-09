package api

import (
	"fmt"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func Authz(c *gin.Context) (bool, string) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, ""
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("REDIRECT_URI")),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistModifyPublic, spotifyauth.ScopePlaylistModifyPrivate),
		spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
	)
	state := uuid.New()
	url := auth.AuthURL(state.String())

	// sessionにstateを格納
	session := sessions.Default(c)
	session.Set("state", state.String())
	session.Save()

	fmt.Println(url)
	return true, url
}
