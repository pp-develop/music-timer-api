package model

import (
	"errors"
)

var (
	ErrFailedGetSession      = errors.New("session: Failed to get session")
	ErrNotFoundPlaylist      = errors.New("playlist: Not Found")
	ErrNotFoundTracks        = errors.New("tracks: Not Found")
	ErrTimeoutCreatePlaylist = errors.New("create playlist: Time out")
	ErrAccessTokenExpired    = errors.New("token expired")
	ErrInvalidState          = errors.New("invalid state")
	ErrFailedGetDB           = errors.New("Failed to get database instance")
	ErrInvalidRequest        = errors.New("Invalid request")
	ErrInvalidRefreshToken   = errors.New("Invalid or expired refresh token")

	// リソース不足エラー
	ErrNotEnoughTracks       = errors.New("Not enough tracks for specified duration")
	ErrNoFavoriteTracks      = errors.New("No favorite tracks found")

	// Spotify API制限エラー
	ErrSpotifyRateLimit      = errors.New("Spotify API rate limit exceeded")
	ErrPlaylistQuotaExceeded = errors.New("Spotify playlist quota exceeded")

	// 処理エラー
	ErrPlaylistCreationFailed = errors.New("Failed to create playlist on Spotify")
	ErrTrackAdditionFailed    = errors.New("Failed to add tracks to playlist")
)
