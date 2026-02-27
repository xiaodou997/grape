package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	globalViper    *viper.Viper
	globalCfgPath  string
)

// GetConfigPath 返回当前配置文件路径
func GetConfigPath() string {
	return globalCfgPath
}

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

	// 保存全局 viper 实例
	globalViper = v
	if configPath != "" {
		globalCfgPath = configPath
	} else {
		globalCfgPath = v.ConfigFileUsed()
	}

	return cfg, nil
}

// Save 将配置写回配置文件（热加载支持）
func Save(cfg *Config) error {
	if globalViper == nil {
		return fmt.Errorf("config not initialized")
	}

	// 更新 viper 中可编辑的字段
	globalViper.Set("registry.upstream", cfg.Registry.Upstream)
	globalViper.Set("registry.upstreams", upstreamsToSlice(cfg.Registry.Upstreams))
	globalViper.Set("auth.jwt_secret", cfg.Auth.JWTSecret)
	globalViper.Set("auth.jwt_expiry", cfg.Auth.JWTExpiry.String())
	globalViper.Set("auth.allow_registration", cfg.Auth.AllowRegistration)
	globalViper.Set("log.level", cfg.Log.Level)

	if globalCfgPath == "" {
		return fmt.Errorf("no config file path available")
	}

	return globalViper.WriteConfigAs(globalCfgPath)
}

// upstreamsToSlice 将 UpstreamConfig slice 转为 viper 可序列化的格式
func upstreamsToSlice(upstreams []UpstreamConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(upstreams))
	for _, u := range upstreams {
		result = append(result, map[string]interface{}{
			"name":    u.Name,
			"url":     u.URL,
			"scope":   u.Scope,
			"timeout": u.Timeout.String(),
			"enabled": u.Enabled,
		})
	}
	return result
}
