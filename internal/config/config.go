package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/IlianBuh/Post-service/internal/config/grpcobj"
	"github.com/IlianBuh/Post-service/internal/config/storage"
	"os"
)

type Config struct {
	Env     string          `json:"env"`
	Storage storage.Storage `json:"storage"`
	GRPC    grpcobj.GRPCObj `json:"grpc"`
}

const (
	defaultConfigPath = "./config/config.json"
)

// New creates new object of applications' configuration
func New() *Config {
	path := fetchConfigPath()

	cfg := MustLoad(path)

	return cfg
}

// MustLoad is wrapper of load function to panic if error occurred
func MustLoad(path string) *Config {
	cfg, err := Load(path)

	if err != nil {
		panic("failed to load config file: " + err.Error())
	}

	return cfg
}

// Load loads config from json file by path. Return error if occurred
func Load(path string) (*Config, error) {
	cfg := new(Config)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	jsonContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	err = json.Unmarshal(jsonContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return cfg, nil
}

// fetchConfigPath fetches config path from either flag 'config' or environment variable.
// If both are empty default value will be returned
// flag > env > default
func fetchConfigPath() string {
	res := ""

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res != "" {
		return res
	}

	res = os.Getenv("CONFIG_PATH")
	if res != "" {
		return res
	}

	return defaultConfigPath
}
