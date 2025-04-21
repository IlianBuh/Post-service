package models

import (
	"time"
)

type Post struct {
	PostId    int
	UserId    int
	CreatedAt time.Time
	Header    string
	Content   string
	Themes    []string
}
