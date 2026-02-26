package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Registry RegistryConfig `mapstructure:"registry"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Log      LogConfig      `mapstructure:"log"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type RegistryConfig struct {
	Upstream string `mapstructure:"upstream"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type AuthConfig struct {
	JWTSecret string        `mapstructure:"jwt_secret"`
	JWTExpiry time.Duration `mapstructure:"jwt_expiry"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         4873,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Registry: RegistryConfig{
			Upstream: "https://registry.npmjs.org",
		},
		Storage: StorageConfig{
			Type: "local",
			Path: "./data",
		},
		Log: LogConfig{
			Level: "info",
		},
		Auth: AuthConfig{
			JWTSecret: "grape-secret-key-change-in-production",
			JWTExpiry: 24 * time.Hour,
		},
	}
}