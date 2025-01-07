package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_FromEnvironment(t *testing.T) {
	t.Setenv("CONFIG_PATH", "environment")
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_SSL_MODE", "disable")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpassword")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("FOREX_KEY", "forexapikey")
	t.Setenv("YOOMONEY_TOKEN", "yoomoneytoken")
	t.Setenv("YOOMONEY_CLIENT_ID", "yoomoneyclientid")
	t.Setenv("YOOMONEY_RECEIVER", "12345")

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "localhost", config.Postgres.Host)
	assert.Equal(t, 5432, config.Postgres.Port)
	assert.Equal(t, "disable", config.Postgres.SSLMode)
	assert.Equal(t, "testdb", config.Postgres.DB)
	assert.Equal(t, "testuser", config.Postgres.User)
	assert.Equal(t, "testpassword", config.Postgres.Password)
	assert.Equal(t, "redis://localhost:6379", config.Redis.URL)
	assert.Equal(t, "forexapikey", config.Forex.Key)
	assert.Equal(t, "yoomoneytoken", config.Yoomoney.Token)
	assert.Equal(t, "yoomoneyclientid", config.Yoomoney.ClientID)
	assert.Equal(t, 12345, config.Yoomoney.Receiver)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "invalid_file_path.yaml")

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unable to process config")
}
