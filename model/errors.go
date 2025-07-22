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
	ErrUnauthorized          = errors.New("Unauthorized")
)
