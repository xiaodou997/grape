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
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`          // Web UI 端口
	APIPort      int           `mapstructure:"api_port"`      // npm Registry API 端口
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	// 默认上游（向后兼容）
	Upstream string `mapstructure:"upstream"`
	// 多上游配置
	Upstreams []UpstreamConfig `mapstructure:"upstreams"`
}

// UpstreamConfig 上游配置
type UpstreamConfig struct {
	// 上游名称（用于标识）
	Name string `mapstructure:"name"`
	// 上游 URL
	URL string `mapstructure:"url"`
	// 匹配的 scope（为空表示默认上游）
	// 例如: "@company", "@internal" 等
	Scope string `mapstructure:"scope"`
	// 超时时间
	Timeout time.Duration `mapstructure:"timeout"`
	// 是否启用
	Enabled bool `mapstructure:"enabled"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type AuthConfig struct {
	JWTSecret         string        `mapstructure:"jwt_secret"`
	JWTExpiry         time.Duration `mapstructure:"jwt_expiry"`
	AllowRegistration bool          `mapstructure:"allow_registration"` // 是否允许自助注册，默认 false
}

type DatabaseConfig struct {
	Type string `mapstructure:"type"` // sqlite | postgres
	DSN  string `mapstructure:"dsn"`  // 数据库连接字符串
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
			Upstreams: []UpstreamConfig{
				{
					Name:    "npmjs",
					URL:     "https://registry.npmjs.org",
					Scope:   "", // 默认上游
					Timeout: 30 * time.Second,
					Enabled: true,
				},
			},
		},
		Storage: StorageConfig{
			Type: "local",
			Path: "./data",
		},
		Log: LogConfig{
			Level: "info",
		},
		Auth: AuthConfig{
			JWTSecret:         "grape-secret-key-change-in-production",
			JWTExpiry:         24 * time.Hour,
			AllowRegistration: false,
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			DSN:  "./data/grape.db",
		},
	}
}