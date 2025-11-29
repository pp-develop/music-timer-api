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
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Duration     int    `json:"duration"` // in milliseconds
	PermalinkURL string `json:"permalink_url"`
	User         struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// Get user's favorite tracks with pagination support
// Fetches all favorite tracks from SoundCloud
func (c *Client) GetFavorites(accessToken string) ([]model.Track, error) {
	var allTracks []model.Track
	trackIDSet := make(map[string]bool) // To detect duplicates
	pageCount := 0

	// Initial URL with linked_partitioning (limit=50 is SoundCloud recommended)
	nextURL := fmt.Sprintf("%s/me/likes/tracks?linked_partitioning=true&limit=50",
		SoundCloudAPIBase)

	log.Printf("[SC-API] Starting to fetch all favorite tracks")

	for nextURL != "" {
		pageCount++
		log.Printf("[SC-API] Fetching page %d", pageCount)

		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("get favorites failed: %s", string(body))
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Parse paginated response
		var paginatedResp struct {
			Collection []SCTrack `json:"collection"`
			NextHref   string    `json:"next_href"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		// If no tracks returned, we've reached the end
		if len(paginatedResp.Collection) == 0 {
			break
		}

		// Convert to model.Track and append (skip duplicates)
		prevCount := len(allTracks)
		for _, scTrack := range paginatedResp.Collection {
			trackID := fmt.Sprintf("%d", scTrack.ID)
			if trackIDSet[trackID] {
				continue
			}
			trackIDSet[trackID] = true
			allTracks = append(allTracks, model.Track{
				Uri:        scTrack.PermalinkURL,
				DurationMs: scTrack.Duration,
				Isrc:       "",
				ArtistsId:  []string{fmt.Sprintf("%d", scTrack.User.ID)},
				ID:         trackID,
			})
		}

		// Update nextURL for next iteration
		nextURL = paginatedResp.NextHref

		if nextURL == "" {
			log.Printf("[SC-API] Reached last page (next_href is empty)")
		}

		// If no new tracks were added, all were duplicates - stop
		if len(allTracks) == prevCount {
			log.Printf("[SC-API] All tracks in batch were duplicates, stopping")
			break
		}
	}

	log.Printf("[SC-API] Successfully fetched %d favorite tracks in %d pages", len(allTracks), pageCount)
	return allTracks, nil
}

// Get tracks by artist (user) ID with pagination support
func (c *Client) GetUserTracks(accessToken string, userID string) ([]model.Track, error) {
	var allTracks []model.Track
	trackIDSet := make(map[string]bool)
	pageCount := 0

	nextURL := fmt.Sprintf("%s/users/%s/tracks?linked_partitioning=true&limit=50",
		SoundCloudAPIBase, userID)

	log.Printf("[SC-API] Starting to fetch tracks for user %s", userID)

	for nextURL != "" {
		pageCount++

		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("get user tracks failed: %s", string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var paginatedResp struct {
			Collection []SCTrack `json:"collection"`
			NextHref   string    `json:"next_href"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		if len(paginatedResp.Collection) == 0 {
			break
		}

		prevCount := len(allTracks)
		for _, scTrack := range paginatedResp.Collection {
			trackID := fmt.Sprintf("%d", scTrack.ID)
			if trackIDSet[trackID] {
				continue
			}
			trackIDSet[trackID] = true
			allTracks = append(allTracks, model.Track{
				Uri:        scTrack.PermalinkURL,
				DurationMs: scTrack.Duration,
				Isrc:       "",
				ArtistsId:  []string{fmt.Sprintf("%d", scTrack.User.ID)},
				ID:         trackID,
			})
		}

		nextURL = paginatedResp.NextHref

		if len(allTracks) == prevCount {
			break
		}
	}

	log.Printf("[SC-API] Successfully fetched %d tracks for user %s in %d pages", len(allTracks), userID, pageCount)
	return allTracks, nil
}

// SoundCloud Playlist
type SCPlaylist struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	PermalinkURL string `json:"permalink_url"`
}

// Create playlist with tracks
func (c *Client) CreatePlaylist(accessToken, title, description string, trackIDs []string) (*SCPlaylist, error) {
	// Convert track IDs to the format required by SoundCloud API
	tracks := make([]map[string]interface{}, len(trackIDs))
	for i, trackID := range trackIDs {
		tracks[i] = map[string]interface{}{
			"id": trackID,
		}
	}

	// Build JSON request body
	playlistData := map[string]interface{}{
		"playlist": map[string]interface{}{
			"title":       title,
			"description": description,
			"sharing":     "public",
			"tracks":      tracks,
		},
	}

	jsonData, err := json.Marshal(playlistData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal playlist data: %v", err)
	}

	log.Printf("[SC-API] Creating playlist with %d tracks, payload: %s", len(trackIDs), string(jsonData))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/playlists", SoundCloudAPIBase), strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("[SC-API] Create playlist failed: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("create playlist failed (status %d): %s", resp.StatusCode, string(body))
	}

	var playlist SCPlaylist
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("failed to decode playlist response: %v", err)
	}

	log.Printf("[SC-API] Playlist created: id=%d, tracks=%d", playlist.ID, len(trackIDs))
	return &playlist, nil
}
