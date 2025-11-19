package auth

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// SpotifyAuthzWeb generates Spotify authorization URL for web applications
func SpotifyAuthzWeb(c *gin.Context) (string, error) {
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

	// Store state in session for CSRF protection
	session := sessions.Default(c)
	session.Set("state", state.String())
	session.Save()

	return url, nil
}

// SpotifyCallbackWeb handles Spotify OAuth callback for web applications
func SpotifyCallbackWeb(c *gin.Context) error {
	code := c.Query("code")
	qState := c.Query("state")

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// Validate state for CSRF protection
	session := sessions.Default(c)
	v := session.Get("state")
	if v == nil {
		return model.ErrFailedGetSession
	}

	state := v.(string)
	if state != qState {
		return model.ErrInvalidState
	}

	// Exchange code for token
	token, err := spotify.ExchangeSpotifyCode(code, os.Getenv("SPOTIFY_REDIRECT_URI"))
	if err != nil {
		return err
	}

	// Get user info and save token
	user, err := spotify.GetMe(token)
	if err != nil {
		return err
	}

	err = database.SaveAccessToken(db, token, user)
	if err != nil {
		return err
	}

	// Set session data
	session.Set("userId", user.ID)
	return session.Save()
}
