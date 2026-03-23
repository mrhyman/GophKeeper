// internal/server/config/config.go
package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Address string    `mapstructure:"address"`
	TLS     TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	// Значения по умолчанию
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.tls.cert", "")
	viper.SetDefault("server.tls.key", "")
	viper.SetDefault("database.dsn", "")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("jwt.secret", "")
	viper.SetDefault("jwt.access_ttl", "15m")
	viper.SetDefault("jwt.refresh_ttl", "720h")
	viper.SetDefault("log.level", "info")

	// Конфиг-файл
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Явный биндинг env-переменных к ключам конфига
	_ = viper.BindEnv("server.address", "GOPHKEEPER_SERVER_ADDRESS")
	_ = viper.BindEnv("server.tls.cert", "GOPHKEEPER_SERVER_TLS_CERT")
	_ = viper.BindEnv("server.tls.key", "GOPHKEEPER_SERVER_TLS_KEY")
	_ = viper.BindEnv("database.dsn", "GOPHKEEPER_DATABASE_DSN")
	_ = viper.BindEnv("database.max_open_conns", "GOPHKEEPER_DATABASE_MAX_OPEN_CONNS")
	_ = viper.BindEnv("database.max_idle_conns", "GOPHKEEPER_DATABASE_MAX_IDLE_CONNS")
	_ = viper.BindEnv("jwt.secret", "GOPHKEEPER_JWT_SECRET")
	_ = viper.BindEnv("jwt.access_ttl", "GOPHKEEPER_JWT_ACCESS_TTL")
	_ = viper.BindEnv("jwt.refresh_ttl", "GOPHKEEPER_JWT_REFRESH_TTL")
	_ = viper.BindEnv("log.level", "GOPHKEEPER_LOG_LEVEL")

	// Флаги командной строки
	pflag.StringP("address", "a", "", "server listen address")
	pflag.StringP("dsn", "d", "", "database DSN")
	pflag.Parse()

	_ = viper.BindPFlag("server.address", pflag.Lookup("address"))
	_ = viper.BindPFlag("database.dsn", pflag.Lookup("dsn"))

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
