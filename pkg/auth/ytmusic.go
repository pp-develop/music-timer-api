package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ytmusicOAuthConfig *oauth2.Config

func init() {
	godotenv.Load()
	ytmusicOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("YTMUSIC_CLIENT_ID"),
		ClientSecret: os.Getenv("YTMUSIC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("YTMUSIC_REDIRECT_URI"),
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// YTMusicAuthzWeb generates YouTube Music authorization URL for web applications
func YTMusicAuthzWeb(c *gin.Context) (string, error) {
	state := uuid.New().String()
	url := ytmusicOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Store state in session for CSRF protection
	session := sessions.Default(c)
	session.Set("ytmusic_state", state)
	err := session.Save()
	if err != nil {
		log.Printf("[ERROR] Failed to save session: %v", err)
		return "", err
	}

	return url, nil
}

// YTMusicCallbackWeb handles OAuth callback for YouTube Music web flow
func YTMusicCallbackWeb(c *gin.Context, db *sql.DB) error {
	// Verify state for CSRF protection
	session := sessions.Default(c)
	savedState := session.Get("ytmusic_state")
	if savedState == nil || savedState != c.Query("state") {
		return fmt.Errorf("state mismatch")
	}

	// Exchange authorization code for token
	code := c.Query("code")
	token, err := ytmusicOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("[ERROR] Token exchange failed: %v", err)
		return err
	}

	// Get user info from Google
	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to get user info: %v", err)
		return err
	}

	// Save token to database
	err = database.SaveYTMusicToken(db, token, userInfo)
	if err != nil {
		log.Printf("[ERROR] Failed to save token: %v", err)
		return err
	}

	// Store user ID in session
	session.Set("ytmusic_user_id", userInfo.ID)
	session.Save()

	log.Printf("[INFO] YouTube Music authentication successful for user: %s", userInfo.Email)
	return nil
}

// getGoogleUserInfo retrieves user information from Google OAuth2 API
func getGoogleUserInfo(accessToken string) (*model.GoogleUser, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var user model.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
