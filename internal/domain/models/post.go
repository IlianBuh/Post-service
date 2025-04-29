package models

import (
	"time"
)

type Post struct {
	Id        int
	UserId    int
	CreatedAt time.Time
	Header    string
	Content   string
	Themes    []string
}
