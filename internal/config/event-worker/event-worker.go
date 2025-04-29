package worker

import (
	"github.com/IlianBuh/Post-service/internal/config/duration"
)

type Config struct {
	PageSize int               `json:"page-size"`
	Interval duration.Duration `json:"interval"`
}
