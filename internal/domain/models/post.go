package models

type Post struct {
	PostId  int
	UserId  int
	Header  string
	Content string
	Themes  []string
}
