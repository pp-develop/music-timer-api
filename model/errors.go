package model

import (
	"errors"
)

var (
	ErrFailedGetSession = errors.New("session: Failed to get userid")
	ErrNotFoundPlaylist = errors.New("playlist: Not Found")
)
