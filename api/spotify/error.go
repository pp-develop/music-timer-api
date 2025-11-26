package spotify

import (
	"errors"
	"strings"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// IsAuthError checks if the error is an authentication error (401)
// Returns true if it's a 401 error indicating token expiration
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a Spotify API error with status code
	var spotifyErr spotify.Error
	if errors.As(err, &spotifyErr) {
		return spotifyErr.Status == 401
	}

	return false
}

// WrapSpotifyError wraps a Spotify API error into appropriate model error
// Returns specific model errors for known error types, otherwise returns the original error
// If returnFallback is true and error is not a known type, returns the fallback error instead
func WrapSpotifyError(err error, fallbackErr ...error) error {
	if err == nil {
		return nil
	}

	// Check if it's a Spotify API error with status code
	var spotifyErr spotify.Error
	if errors.As(err, &spotifyErr) {
		// 401 Unauthorized - token expired
		if spotifyErr.Status == 401 {
			return model.ErrAccessTokenExpired
		}

		// 429 Too Many Requests - rate limit
		if spotifyErr.Status == 429 {
			return model.ErrSpotifyRateLimit
		}
	}

	// Check error message for specific patterns (fallback for cases where status code isn't available)
	errMsg := err.Error()

	// 通常、エラーの種類はステータスコードで判定するのが望ましいが、
	// 現在使用しているフレームワークの制約により、エラーメッセージの文字列を判定する方法も採用している。
	if strings.Contains(errMsg, "token expired") {
		return model.ErrAccessTokenExpired
	}
	if strings.Contains(errMsg, "rate limit") {
		return model.ErrSpotifyRateLimit
	}
	if strings.Contains(errMsg, "quota") {
		return model.ErrPlaylistQuotaExceeded
	}

	// If fallback error is provided, return it for unknown errors
	if len(fallbackErr) > 0 && fallbackErr[0] != nil {
		return fallbackErr[0]
	}

	return err
}
