package model

import (
	"errors"
)

var (
	ErrFailedGetSession      = errors.New("session: Failed to get session")
	ErrNotFoundPlaylist      = errors.New("playlist: Not Found")
	ErrTimeoutCreatePlaylist = errors.New("create playlist: Time out")
	ErrAccessTokenExpired    = errors.New("token expired")
	ErrInvalidState          = errors.New("state invalid")
)
