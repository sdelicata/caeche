package config

import (
	"github.com/BurntSushi/toml"
)

const (
	DEFAULT_PORT string = "8080"
	DEFAULT_READ_TIMEOUT int = 10
	DEFAULT_WRITE_TIMEOUT int = 10
	DEFAULT_BACKEND_SCHEME string = "http"
	DEFAULT_BACKEND_HOST string = ":80"
)

type Config struct {
	Port    string
	ReadTimeout int
	WriteTimeout int
	Backend BackendConfig
}

type BackendConfig struct {
	Host   string
	Scheme string
}

func NewConfigFromFile(filePath string) (Config, error) {

	config := NewConfigWithDefault()
	_, err := toml.DecodeFile(filePath, &config)

	return config, err
}

func NewConfigWithDefault() Config {
	return Config{
		Port: DEFAULT_PORT,
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		Backend: BackendConfig{
			Host:   DEFAULT_BACKEND_HOST,
			Scheme: DEFAULT_BACKEND_SCHEME,
		},
	}
}