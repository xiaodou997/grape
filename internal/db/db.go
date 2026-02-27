package db

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	Type string // sqlite | postgres
	DSN  string // 数据库连接字符串
}

// Init 初始化数据库连接
func Init(cfg *Config) error {
	var err error
	var gormDB *gorm.DB

	switch cfg.Type {
	case "sqlite":
		// 确保 SQLite 文件目录存在
		dir := filepath.Dir(cfg.DSN)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}

		gormDB, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case "postgres":
		// TODO: PostgreSQL 支持
		return fmt.Errorf("postgres not implemented yet")
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = gormDB
	return nil
}

// Migrate 自动迁移数据库表结构
func Migrate(models ...interface{}) error {
	return DB.AutoMigrate(models...)
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
