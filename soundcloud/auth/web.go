package auth

import (
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

// AuthzWeb generates SoundCloud authorization URL for web applications
func AuthzWeb(c *gin.Context) (string, error) {
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
		log.Printf("[SC-AUTH-WEB] ERROR: Failed to save session: %v", err)
		return "", err
	}

	return url, nil
}

// CallbackWeb handles SoundCloud OAuth callback for web applications
func CallbackWeb(c *gin.Context) error {
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
		log.Printf("[SC-AUTH-WEB] ERROR: Session state is nil, query state: %s", qState)
		return model.ErrFailedGetSession
	}

	state := v.(string)
	if state != qState {
		log.Printf("[SC-AUTH-WEB] ERROR: State mismatch. Expected: %s, Got: %s", state, qState)
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
