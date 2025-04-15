package storage

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotCreator     = errors.New("user is not creator")
	ErrBlockedChannel = errors.New("channel is blocked to write")
	ErrClose          = errors.New("failed to close database")
)
