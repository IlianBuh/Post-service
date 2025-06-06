package storage

import (
	"errors"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrNotCreator = errors.New("user is not creator")
	ErrClose      = errors.New("failed to close database")
	ErrNoEvents   = errors.New("no new events")
)
