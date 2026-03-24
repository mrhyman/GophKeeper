package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetViperAndFlags() {
	viper.Reset()
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
}

func TestLoad_Defaults(t *testing.T) {
	resetViperAndFlags()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Server.Address)
	assert.Equal(t, "", cfg.Server.TLS.Cert)
	assert.Equal(t, "", cfg.Server.TLS.Key)
	assert.Equal(t, "", cfg.Database.DSN)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, "", cfg.JWT.Secret)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTTL)
	assert.Equal(t, 720*time.Hour, cfg.JWT.RefreshTTL)
	assert.Equal(t, "info", cfg.Log.Level)
}

func TestLoad_FromConfigFile(t *testing.T) {
	resetViperAndFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "server.yaml")

	configContent := `
server:
  address: ":9090"
  tls:
    cert: "/path/to/cert.pem"
    key: "/path/to/key.pem"
database:
  dsn: "postgres://user:pass@localhost/db"
  max_open_conns: 50
  max_idle_conns: 20
jwt:
  secret: "super-secret"
  access_ttl: "30m"
  refresh_ttl: "168h"
log:
  level: "debug"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	viper.AddConfigPath(tmpDir)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.Server.Address)
	assert.Equal(t, "/path/to/cert.pem", cfg.Server.TLS.Cert)
	assert.Equal(t, "/path/to/key.pem", cfg.Server.TLS.Key)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.Database.DSN)
	assert.Equal(t, 50, cfg.Database.MaxOpenConns)
	assert.Equal(t, 20, cfg.Database.MaxIdleConns)
	assert.Equal(t, "super-secret", cfg.JWT.Secret)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTTL)
	assert.Equal(t, "debug", cfg.Log.Level)
}

func TestLoad_FromEnv(t *testing.T) {
	resetViperAndFlags()

	require.NoError(t, os.Setenv("GOPHKEEPER_SERVER_ADDRESS", ":7070"))
	require.NoError(t, os.Setenv("GOPHKEEPER_DATABASE_DSN", "postgres://env@localhost/envdb"))
	require.NoError(t, os.Setenv("GOPHKEEPER_JWT_SECRET", "env-secret"))
	defer func() {
		require.NoError(t, os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS"))
		require.NoError(t, os.Unsetenv("GOPHKEEPER_DATABASE_DSN"))
		require.NoError(t, os.Unsetenv("GOPHKEEPER_JWT_SECRET"))
	}()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, ":7070", cfg.Server.Address)
	assert.Equal(t, "postgres://env@localhost/envdb", cfg.Database.DSN)
	assert.Equal(t, "env-secret", cfg.JWT.Secret)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	resetViperAndFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "server.yaml")

	err := os.WriteFile(configPath, []byte("invalid: yaml: [[["), 0644)
	require.NoError(t, err)

	viper.AddConfigPath(tmpDir)

	_, err = Load()

	assert.Error(t, err)
}
