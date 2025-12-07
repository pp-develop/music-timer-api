package soundcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	ID        int    `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
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

	slog.Info("starting to fetch all favorite tracks")

	for nextURL != "" {
		pageCount++
		slog.Debug("fetching page", slog.Int("page", pageCount))

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
			slog.Debug("reached last page")
		}

		// If no new tracks were added, all were duplicates - stop
		if len(allTracks) == prevCount {
			slog.Debug("all tracks in batch were duplicates, stopping")
			break
		}
	}

	slog.Info("successfully fetched favorite tracks", slog.Int("track_count", len(allTracks)), slog.Int("page_count", pageCount))
	return allTracks, nil
}

// Get tracks by artist (user) ID with pagination support
func (c *Client) GetUserTracks(accessToken string, userID string) ([]model.Track, error) {
	var allTracks []model.Track
	trackIDSet := make(map[string]bool)
	pageCount := 0

	nextURL := fmt.Sprintf("%s/users/%s/tracks?linked_partitioning=true&limit=50",
		SoundCloudAPIBase, userID)

	slog.Debug("starting to fetch tracks for user", slog.String("user_id", userID))

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

	slog.Debug("successfully fetched tracks for user", slog.Int("track_count", len(allTracks)), slog.String("user_id", userID), slog.Int("page_count", pageCount))
	return allTracks, nil
}

// SoundCloud Playlist
type SCPlaylist struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	PermalinkURL string `json:"permalink_url"`
	SecretToken  string `json:"secret_token"`
}

// Delete playlist by ID
func (c *Client) DeletePlaylist(accessToken string, playlistID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/playlists/%s", SoundCloudAPIBase, playlistID), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// SoundCloud returns 200 OK on successful deletion
	// 404 means playlist already deleted on SoundCloud - treat as success
	if resp.StatusCode == http.StatusNotFound {
		slog.Debug("playlist already deleted on SoundCloud, skipping", slog.String("playlist_id", playlistID))
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete playlist failed (status %d): %s", resp.StatusCode, string(body))
	}

	slog.Info("playlist deleted successfully", slog.String("playlist_id", playlistID))
	return nil
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

	slog.Debug("creating playlist", slog.Int("track_count", len(trackIDs)))

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
		slog.Error("create playlist failed", slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
		return nil, fmt.Errorf("create playlist failed (status %d): %s", resp.StatusCode, string(body))
	}

	var playlist SCPlaylist
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("failed to decode playlist response: %v", err)
	}

	slog.Info("playlist created", slog.Int("playlist_id", playlist.ID), slog.Int("track_count", len(trackIDs)))
	return &playlist, nil
}

// Get users that the current user is following (artists)
func (c *Client) GetFollowings(accessToken string) ([]SCUser, error) {
	var allUsers []SCUser
	userIDSet := make(map[int]bool)
	pageCount := 0

	nextURL := fmt.Sprintf("%s/me/followings?linked_partitioning=true&limit=50", SoundCloudAPIBase)

	slog.Debug("starting to fetch followed users")

	for nextURL != "" {
		pageCount++
		slog.Debug("fetching followings page", slog.Int("page", pageCount))

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
			return nil, fmt.Errorf("get followings failed: %s", string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var paginatedResp struct {
			Collection []SCUser `json:"collection"`
			NextHref   string   `json:"next_href"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		if len(paginatedResp.Collection) == 0 {
			break
		}

		prevCount := len(allUsers)
		for _, user := range paginatedResp.Collection {
			if userIDSet[user.ID] {
				continue
			}
			userIDSet[user.ID] = true
			allUsers = append(allUsers, user)
		}

		nextURL = paginatedResp.NextHref

		if len(allUsers) == prevCount {
			break
		}
	}

	slog.Info("successfully fetched followed users", slog.Int("user_count", len(allUsers)), slog.Int("page_count", pageCount))
	return allUsers, nil
}
