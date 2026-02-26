package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	cfg := Default()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/grape")
	}

	v.SetEnvPrefix("GRAPE")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Storage.Path != "" {
		if err := os.MkdirAll(cfg.Storage.Path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
		packagesDir := filepath.Join(cfg.Storage.Path, "packages")
		if err := os.MkdirAll(packagesDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create packages directory: %w", err)
		}
	}

	return cfg, nil
}
