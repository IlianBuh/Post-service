package userProvider

import (
	"github.com/IlianBuh/Post-service/internal/config/duration"
)

type UserProvider struct {
	Host    string            `json:"host"`
	Port    int               `json:"port"`
	Timeout duration.Duration `json:"timeout"`
}
