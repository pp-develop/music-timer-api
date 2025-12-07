package auth

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// SpotifyAuthzNative generates Spotify authorization URL for native applications
func SpotifyAuthzNative(c *gin.Context) (string, error) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("SPOTIFY_REDIRECT_URI_NATIVE")),
		spotifyauth.WithScopes(
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserFollowRead,
		),
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
	)

	// For native apps, we don't need CSRF protection with state
	// as the response is returned directly as JSON, not via redirect
	url := auth.AuthURL("native")

	return url, nil
}

// SpotifyCallbackNative handles Spotify OAuth callback for native applications
func SpotifyCallbackNative(c *gin.Context) (*utils.TokenPair, error) {
	code := c.Query("code")

	db, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	// Exchange code for token
	token, err := spotify.ExchangeSpotifyCode(code, os.Getenv("SPOTIFY_REDIRECT_URI_NATIVE"))
	if err != nil {
		return nil, err
	}

	// Get user info and save token
	user, err := spotify.GetMe(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}

	err = database.SaveAccessToken(db, token, user)
	if err != nil {
		return nil, err
	}

	// Generate JWT token pair
	jti := uuid.New().String()
	tokenPair, err := utils.GenerateTokenPair(user.ID, "spotify", jti)
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	err = database.SaveSpotifyRefreshToken(db, jti, user.ID, expiresAt)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func RefreshAccessToken(c *gin.Context) (*utils.TokenPair, error) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, model.ErrInvalidRequest
	}

	// Validate refresh token
	userID, service, jti, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	db, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	// Check if refresh token is valid in database
	valid, err := database.IsSpotifyRefreshTokenValid(db, jti)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, model.ErrInvalidRefreshToken
	}

	// Generate new token pair
	newJTI := uuid.New().String()
	tokenPair, err := utils.GenerateTokenPair(userID, service, newJTI)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	err = database.RevokeSpotifyRefreshToken(db, jti)
	if err != nil {
		return nil, err
	}

	// Save new refresh token
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	err = database.SaveSpotifyRefreshToken(db, newJTI, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}
