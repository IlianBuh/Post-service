package grpcobj

import (
	"github.com/IlianBuh/Post-service/internal/config/grpcobj/duration"
)

// GRPCObj object representation of json data
type GRPCObj struct {
	Host    string            `json:"host"`
	Port    int               `json:"port"`
	Timeout duration.Duration `json:"timeout"`
}
