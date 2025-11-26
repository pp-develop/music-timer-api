package ytmusic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	pythonScriptPath = "pkg/ytmusic/ytmusic_client.py"
	oauthTempDir     = "/tmp/ytmusic_oauth"
)

// Client manages YouTube Music API operations
type Client struct {
	db     *sql.DB
	userID string
}

// NewClient creates a new YouTube Music client
func NewClient(db *sql.DB, userID string) *Client {
	return &Client{
		db:     db,
		userID: userID,
	}
}

// getOAuthConfig returns the OAuth2 config for YouTube Music
func getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
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

// createOAuthFile creates a temporary OAuth credentials file for ytmusicapi
func (c *Client) createOAuthFile() (string, error) {
	// Get user from database
	user, err := database.GetYTMusicUser(c.db, c.userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Create OAuth token
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       time.Unix(user.TokenExpiration, 0),
		TokenType:    "Bearer",
	}

	// Refresh token if expired
	config := getOAuthConfig()
	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update token in database if refreshed
	if newToken.AccessToken != token.AccessToken {
		err = database.UpdateYTMusicToken(c.db, newToken, c.userID)
		if err != nil {
			log.Printf("[WARN] Failed to update refreshed token: %v", err)
		}
	}

	// Create temp directory if not exists
	if err := os.MkdirAll(oauthTempDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create OAuth credentials in ytmusicapi format
	oauthData := map[string]interface{}{
		"access_token":  newToken.AccessToken,
		"refresh_token": newToken.RefreshToken,
		"expires_at":    newToken.Expiry.Unix(),
		"token_type":    "Bearer",
		"scope":         "https://www.googleapis.com/auth/youtube https://www.googleapis.com/auth/youtube.readonly",
	}

	// Write to temporary file
	oauthFile := filepath.Join(oauthTempDir, fmt.Sprintf("oauth_%s.json", c.userID))
	file, err := os.Create(oauthFile)
	if err != nil {
		return "", fmt.Errorf("failed to create oauth file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(oauthData); err != nil {
		return "", fmt.Errorf("failed to write oauth file: %w", err)
	}

	return oauthFile, nil
}

// cleanupOAuthFile removes the temporary OAuth file
func (c *Client) cleanupOAuthFile(oauthFile string) {
	if err := os.Remove(oauthFile); err != nil {
		log.Printf("[WARN] Failed to cleanup oauth file: %v", err)
	}
}

// GetLikedSongs retrieves user's liked songs from YouTube Music
func (c *Client) GetLikedSongs() ([]model.YouTubeTrack, error) {
	// Create temporary OAuth file
	oauthFile, err := c.createOAuthFile()
	if err != nil {
		return nil, err
	}
	defer c.cleanupOAuthFile(oauthFile)

	// Execute Python script
	cmd := exec.Command("python3", pythonScriptPath, "get_liked_songs", oauthFile)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python script error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute python script: %w", err)
	}

	// Parse response
	var response struct {
		Tracks []model.YouTubeTrack `json:"tracks"`
		Count  int                  `json:"count"`
		Error  string               `json:"error,omitempty"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse python output: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("ytmusic error: %s", response.Error)
	}

	log.Printf("[INFO] Retrieved %d liked songs for user %s", response.Count, c.userID)
	return response.Tracks, nil
}

// CreatePlaylist creates a new playlist with specified tracks
func (c *Client) CreatePlaylist(title, description string, videoIDs []string) (string, error) {
	// Create temporary OAuth file
	oauthFile, err := c.createOAuthFile()
	if err != nil {
		return "", err
	}
	defer c.cleanupOAuthFile(oauthFile)

	// Convert video IDs to JSON
	videoIDsJSON, err := json.Marshal(videoIDs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal video IDs: %w", err)
	}

	// Execute Python script
	cmd := exec.Command("python3", pythonScriptPath, "create_playlist", oauthFile, title, description, string(videoIDsJSON))
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("python script error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute python script: %w", err)
	}

	// Parse response
	var response struct {
		PlaylistID string `json:"playlist_id"`
		Success    bool   `json:"success"`
		Error      string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return "", fmt.Errorf("failed to parse python output: %w", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("ytmusic error: %s", response.Error)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to create playlist")
	}

	log.Printf("[INFO] Created playlist %s for user %s", response.PlaylistID, c.userID)
	return response.PlaylistID, nil
}
