package model

// YouTubeTrack represents a track from YouTube Music
type YouTubeTrack struct {
	VideoID      string   `json:"video_id"`
	DurationMs   int      `json:"duration_ms"`
	Title        string   `json:"title"`
	ChannelID    string   `json:"channel_id"`
	ChannelName  string   `json:"channel_name"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Artists      []string `json:"artists"`
}

// YTMusicUser represents a YouTube Music user with OAuth tokens
type YTMusicUser struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	TokenExpiration int64  `json:"token_expiration"`
}

// GoogleUser represents user info from Google OAuth2
type GoogleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
