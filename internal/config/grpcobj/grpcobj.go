package grpcobj

import (
	"github.com/IlianBuh/Post-service/internal/config/duration"
)

// GRPCObj object representation of json data
type Config struct {
	Host    string            `json:"host"`
	Port    int               `json:"port"`
	Timeout duration.Duration `json:"timeout"`
}
