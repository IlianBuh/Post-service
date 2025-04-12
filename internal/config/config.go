package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {

	value := ""
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	if d.Duration, err = time.ParseDuration(value); err != nil {
		return err
	}

	return nil
}

type Config struct {
	Env     string  `json:"env"`
	Storage Storage `json:"storage"`
	GRPC    GRPCObj `json:"grpc"`
}

type GRPCObj struct {
	Host    string   `json:"host"`
	Port    int      `json:"port"`
	Timeout Duration `json:"timeout"`
}

type Storage struct {
	DBName   string `json:"dbname"`
	User     string `json:"name"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Timeout  int    `json:"connection-timeout,omitempty"`
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
