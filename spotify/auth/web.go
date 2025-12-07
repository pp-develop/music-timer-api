package auth

import (
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// SpotifyAuthzWeb generates Spotify authorization URL for web applications
func SpotifyAuthzWeb(c *gin.Context) (string, error) {
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
	if err := session.Save(); err != nil {
		log.Printf("[ERROR] Failed to save session: %v", err)
		return "", err
	}

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
		log.Printf("[ERROR] Callback - Session state is nil, query state: %s", qState)
		return model.ErrFailedGetSession
	}

	state := v.(string)
	if state != qState {
		log.Printf("[ERROR] Callback - State mismatch. Expected: %s, Got: %s", state, qState)
		return model.ErrInvalidState
	}

	// Exchange code for token
	token, err := spotify.ExchangeSpotifyCode(code, os.Getenv("SPOTIFY_REDIRECT_URI"))
	if err != nil {
		return err
	}

	// Get user info and save token
	user, err := spotify.GetMe(c.Request.Context(), token)
	if err != nil {
		return err
	}

	err = database.SaveAccessToken(db, token, user)
	if err != nil {
		return err
	}

	// Set session data
	session.Set("userId", user.ID)
	session.Set("service", "spotify")
	return session.Save()
}
