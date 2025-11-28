package auth

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

// SoundCloudAuthzWeb generates SoundCloud authorization URL for web applications
func SoundCloudAuthzWeb(c *gin.Context) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	client := soundcloud.NewClient()
	state := uuid.New().String()
	redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI")

	url := client.GetAuthURL(redirectURI, state)

	// Store state in session for CSRF protection
	session := sessions.Default(c)
	session.Set("sc_state", state)
	err = session.Save()
	if err != nil {
		log.Printf("[ERROR] Failed to save session: %v", err)
		return "", err
	}

	return url, nil
}

// SoundCloudCallbackWeb handles SoundCloud OAuth callback for web applications
func SoundCloudCallbackWeb(c *gin.Context) error {
	code := c.Query("code")
	qState := c.Query("state")

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// Validate state for CSRF protection
	session := sessions.Default(c)
	v := session.Get("sc_state")
	if v == nil {
		log.Printf("[ERROR] SoundCloud Callback - Session state is nil, query state: %s", qState)
		return model.ErrFailedGetSession
	}

	state := v.(string)
	if state != qState {
		log.Printf("[ERROR] SoundCloud Callback - State mismatch. Expected: %s, Got: %s", state, qState)
		return model.ErrInvalidState
	}

	// Exchange code for token
	client := soundcloud.NewClient()
	redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI")

	tokenResp, err := client.ExchangeCode(code, redirectURI)
	if err != nil {
		return err
	}

	// Get user info
	userInfo, err := client.GetMe(tokenResp.AccessToken)
	if err != nil {
		return err
	}

	// Calculate token expiration
	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600 // Default 1 hour
	}

	// Save user info to database
	user := &model.SoundCloudUser{
		Id:              strconv.Itoa(userInfo.ID),
		Username:        userInfo.Username,
		AccessToken:     tokenResp.AccessToken,
		RefreshToken:    tokenResp.RefreshToken,
		TokenExpiration: int(time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()),
	}

	err = database.CreateOrUpdateSoundCloudUser(db, user)
	if err != nil {
		return err
	}

	// Set session data
	session.Set("userId", user.Id)
	session.Set("service", "soundcloud")
	return session.Save()
}

// GetSoundCloudAuthStatus returns authentication status for SoundCloud users
func GetSoundCloudAuthStatus(c *gin.Context) (*model.SoundCloudUser, error) {
	db, ok := utils.GetDB(c)
	if !ok {
		log.Println("[SC-AUTH] ERROR: Failed to get DB instance")
		return nil, model.ErrFailedGetDB
	}

	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		log.Println("[SC-AUTH] ERROR: No userId in session")
		return nil, model.ErrFailedGetSession
	}

	userId := v.(string)

	user, err := database.GetSoundCloudUser(db, userId)
	if err != nil {
		log.Printf("[SC-AUTH] ERROR: Failed to get user from database: %v", err)
		return nil, err
	}

	// Check if token is expired
	if time.Now().Unix() > int64(user.TokenExpiration) {
		// Try to refresh token
		client := soundcloud.NewClient()
		tokenResp, err := client.RefreshToken(user.RefreshToken)
		if err != nil {
			log.Printf("[SC-AUTH] ERROR: Failed to refresh token: %v", err)
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Update tokens
		expiresIn := tokenResp.ExpiresIn
		if expiresIn == 0 {
			expiresIn = 3600
		}

		err = database.UpdateSoundCloudUserTokens(
			db,
			userId,
			tokenResp.AccessToken,
			tokenResp.RefreshToken,
			int(time.Now().Add(time.Duration(expiresIn)*time.Second).Unix()),
		)
		if err != nil {
			return nil, err
		}

		user.AccessToken = tokenResp.AccessToken
		user.RefreshToken = tokenResp.RefreshToken
		user.TokenExpiration = int(time.Now().Add(time.Duration(expiresIn) * time.Second).Unix())
	}

	return user, nil
}
