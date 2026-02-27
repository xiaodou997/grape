package storage

import "time"

// PackageInfo 包信息
type PackageInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Private     bool      `json:"private"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// StorageStats 存储统计
type StorageStats struct {
	TotalPackages int64
	TotalSize     int64
	CachedCount   int64
	PrivateCount  int64
}

// Storage 存储接口
type Storage interface {
	// 包管理
	HasPackage(name string) bool
	GetMetadata(name string) ([]byte, error)
	SaveMetadata(name string, data []byte) error
	DeletePackage(name string) error

	// tarball 管理
	HasTarball(name, filename string) bool
	GetTarball(name, filename string) ([]byte, error)
	SaveTarball(name, filename string, data []byte) error
	DeleteTarball(name, filename string) error

	// 查询
	ListPackages() ([]PackageInfo, error)
	GetStats() (*StorageStats, error)
}
