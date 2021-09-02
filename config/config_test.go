package config

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileOverridesDefaultValues(t *testing.T) {
	const expectedPort = "1234"
	const expectedReadTimeout = 30
	const expectedWriteTimeout = 30
	const expectedBackendHost = "domain.com:80"
	const expectedBackendScheme = "https"
	const expectedCacheDefaultTTL = 60

	configContent := Config{
		Port: expectedPort,
		ReadTimeout: expectedReadTimeout,
		WriteTimeout: expectedWriteTimeout,
		Backend: BackendConfig{
			Host:   expectedBackendHost,
			Scheme: expectedBackendScheme,
		},
		Cache: CacheConfig{
			DefaultTTL: expectedCacheDefaultTTL,
		},
	}

	configFile := createTempFileFromConfig(configContent)
	defer os.Remove(configFile)
	config, _ := NewConfigFromFile(configFile)

	assert.Equal(t, expectedPort, config.Port, "Wrong port")
	assert.Equal(t, expectedReadTimeout, config.ReadTimeout, "Wrong read timeout")
	assert.Equal(t, expectedWriteTimeout, config.WriteTimeout, "Wrong write timeout")
	assert.Equal(t, expectedBackendHost, config.Backend.Host, "Wrong backend host")
	assert.Equal(t, expectedBackendScheme, config.Backend.Scheme, "Wrong backend scheme")
	assert.Equal(t, expectedCacheDefaultTTL, config.Cache.DefaultTTL, "Wrong cache default TTL")
}

func createTempFileFromConfig(config Config) string {
	tempFile, err := ioutil.TempFile("./", "config_test_*.toml")
	if err != nil {
		panic("Cannot create temp file")
	}
	buffer := new(bytes.Buffer)
	if err := toml.NewEncoder(buffer).Encode(config); err != nil {
		panic("Cannot encode Config in TOML")
	}
	tempFile.WriteString(buffer.String())

	return tempFile.Name()
}