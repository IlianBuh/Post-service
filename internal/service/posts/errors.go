package posts

import (
	"errors"
)

var (
	ErrInternal       = errors.New("internal error")
	ErrNotFound       = errors.New("not found")
	ErrNotCreator     = errors.New("user is not creator")
	ErrUserNotFound   = errors.New("user does not exist")
	ErrBlockedChannel = errors.New("channel is blocked to write")
)
