package soundcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

const (
	SoundCloudAPIBase  = "https://api.soundcloud.com"
	SoundCloudAuthBase = "https://secure.soundcloud.com"
)

type Client struct {
	HTTPClient   *http.Client
	ClientID     string
	ClientSecret string
}

func NewClient() *Client {
	return &Client{
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		ClientID:     os.Getenv("SOUNDCLOUD_CLIENT_ID"),
		ClientSecret: os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
	}
}

// OAuth Token Response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// Get Authorization URL
func (c *Client) GetAuthURL(redirectURI, state string) string {
	params := url.Values{}
	params.Add("client_id", c.ClientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "non-expiring")
	params.Add("state", state)

	return fmt.Sprintf("%s/connect?%s", SoundCloudAuthBase, params.Encode())
}

// Exchange authorization code for access token
func (c *Client) ExchangeCode(code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", SoundCloudAPIBase), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Refresh access token
func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", SoundCloudAPIBase), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// SoundCloud User Info
type SCUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// Get current user info
func (c *Client) GetMe(accessToken string) (*SCUser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/me", SoundCloudAPIBase), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get me failed: %s", string(body))
	}

	var user SCUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SoundCloud Track
type SCTrack struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"` // in milliseconds
	PermalinkURL string `json:"permalink_url"`
	User        struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// Get user's favorite tracks
func (c *Client) GetFavorites(accessToken string, limit int) ([]model.Track, error) {
	url := fmt.Sprintf("%s/me/favorites?limit=%d", SoundCloudAPIBase, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get favorites failed: %s", string(body))
	}

	var scTracks []SCTrack
	if err := json.NewDecoder(resp.Body).Decode(&scTracks); err != nil {
		return nil, err
	}

	// Convert to model.Track
	tracks := make([]model.Track, 0, len(scTracks))
	for _, scTrack := range scTracks {
		tracks = append(tracks, model.Track{
			Uri:        scTrack.PermalinkURL,
			DurationMs: scTrack.Duration,
			Isrc:       "",
			ArtistsId:  []string{fmt.Sprintf("%d", scTrack.User.ID)},
			ID:         fmt.Sprintf("%d", scTrack.ID),
		})
	}

	return tracks, nil
}

// SoundCloud Playlist
type SCPlaylist struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	PermalinkURL string `json:"permalink_url"`
}

// Create playlist
func (c *Client) CreatePlaylist(accessToken, title, description string) (*SCPlaylist, error) {
	data := url.Values{}
	data.Set("playlist[title]", title)
	data.Set("playlist[description]", description)
	data.Set("playlist[sharing]", "public")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/playlists", SoundCloudAPIBase), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create playlist failed: %s", string(body))
	}

	var playlist SCPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	return &playlist, nil
}

// Add tracks to playlist
func (c *Client) AddTracksToPlaylist(accessToken string, playlistID int, trackIDs []string) error {
	data := url.Values{}
	for i, trackID := range trackIDs {
		// Use track ID instead of permalink URL
		data.Add(fmt.Sprintf("playlist[tracks][%d][id]", i), trackID)
	}

	apiURL := fmt.Sprintf("%s/playlists/%d", SoundCloudAPIBase, playlistID)

	req, err := http.NewRequest("PUT", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("[SC-API] Failed to create request: %v", err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[SC-API] HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add tracks failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
