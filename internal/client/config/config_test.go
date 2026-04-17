package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetViper() {
	viper.Reset()
}

func TestLoad_Defaults(t *testing.T) {
	resetViper()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://localhost:8080", cfg.Server.Address)
	assert.Equal(t, "", cfg.Server.CAFile)
	assert.Equal(t, "~/.gophkeeper/vault.db", cfg.Storage.Path)
	assert.True(t, cfg.Sync.Auto)
	assert.Equal(t, 5*time.Minute, cfg.Sync.Interval)
}

func TestLoad_FromConfigFile(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.yaml")

	configContent := `
server:
  address: "https://example.com:9090"
  ca_file: "/path/to/ca.crt"
storage:
  path: "/custom/path/vault.db"
sync:
  auto: false
  interval: "10m"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	viper.AddConfigPath(tmpDir)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://example.com:9090", cfg.Server.Address)
	assert.Equal(t, "/path/to/ca.crt", cfg.Server.CAFile)
	assert.Equal(t, "/custom/path/vault.db", cfg.Storage.Path)
	assert.False(t, cfg.Sync.Auto)
	assert.Equal(t, 10*time.Minute, cfg.Sync.Interval)
}

func TestLoad_FromEnv(t *testing.T) {
	resetViper()

	// AutomaticEnv с префиксом GOPHKEEPER заменяет точки на подчёркивания
	// server.address -> GOPHKEEPER_SERVER_ADDRESS
	err := os.Setenv("GOPHKEEPER_SERVER_ADDRESS", "https://env.example.com:7070")
	require.NoError(t, err)
	defer func() {
		err := os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
		require.NoError(t, err)
	}()

	// Нужно явно забиндить переменную, т.к. AutomaticEnv работает только при Get()
	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()
	_ = viper.BindEnv("server.address", "GOPHKEEPER_SERVER_ADDRESS")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://env.example.com:7070", cfg.Server.Address)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.yaml")

	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	viper.AddConfigPath(tmpDir)

	_, err = Load()

	assert.Error(t, err)
}

func TestLoad_PartialConfigFile(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.yaml")

	// Только часть настроек, остальные должны быть дефолтными
	configContent := `
server:
  address: "https://partial.example.com:8080"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	viper.AddConfigPath(tmpDir)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "https://partial.example.com:8080", cfg.Server.Address)
	// Дефолтные значения
	assert.Equal(t, "", cfg.Server.CAFile)
	assert.True(t, cfg.Sync.Auto)
}
