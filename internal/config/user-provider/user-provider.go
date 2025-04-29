package userProvider

import (
	"github.com/IlianBuh/Post-service/internal/config/duration"
)

type Config struct {
	Host    string            `json:"host"`
	Port    int               `json:"port"`
	Timeout duration.Duration `json:"timeout"`
}
