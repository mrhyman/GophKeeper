package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config — конфигурация клиента.
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Sync    SyncConfig    `mapstructure:"sync"`
}

// ServerConfig — настройки подключения к серверу.
type ServerConfig struct {
	Address string `mapstructure:"address"`
	CAFile  string `mapstructure:"ca_file"`
}

// StorageConfig — настройки локального хранилища.
type StorageConfig struct {
	Path string `mapstructure:"path"`
}

// SyncConfig — настройки автосинхронизации.
type SyncConfig struct {
	Auto     bool          `mapstructure:"auto"`
	Interval time.Duration `mapstructure:"interval"`
}

// Load загружает конфигурацию.
func Load() (*Config, error) {
	viper.SetDefault("server.address", "https://localhost:8080")
	viper.SetDefault("server.ca_file", "")
	viper.SetDefault("storage.path", "~/.gophkeeper/vault.db")
	viper.SetDefault("sync.auto", true)
	viper.SetDefault("sync.interval", "5m")

	viper.SetConfigName("client")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.gophkeeper")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
