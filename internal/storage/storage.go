package storage

import (
	"errors"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrNotCreator = errors.New("user is not creator")
)
