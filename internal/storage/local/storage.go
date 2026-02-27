package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/registry"
	"github.com/graperegistry/grape/internal/storage"
)

var (
	ErrInvalidPath      = errors.New("invalid path: path traversal attempt detected")
	ErrInvalidPackageName = errors.New("invalid package name")
	ErrInvalidFilename    = errors.New("invalid filename")
)

type Storage struct {
	basePath string
}

func New(basePath string) *Storage {
	return &Storage{basePath: basePath}
}

func (s *Storage) packagesDir() string {
	return filepath.Join(s.basePath, "packages")
}

// validatePath 验证路径是否在 basePath 范围内，防止路径穿越攻击
func (s *Storage) validatePath(resolved string) error {
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}
	
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	
	// 确保解析后的路径在 basePath 范围内
	if !strings.HasPrefix(absResolved, absBase+string(filepath.Separator)) && absResolved != absBase {
		return ErrInvalidPath
	}
	return nil
}

// validatePackageName 验证包名是否合法
func validatePackageName(name string) error {
	if name == "" {
		return ErrInvalidPackageName
	}
	
	// 检查路径穿越字符
	if strings.Contains(name, "..") {
		return ErrInvalidPackageName
	}
	if strings.Contains(name, "\x00") {
		return ErrInvalidPackageName
	}
	
	// 检查绝对路径
	if filepath.IsAbs(name) {
		return ErrInvalidPackageName
	}
	
	// 检查是否以 / 开头
	if strings.HasPrefix(name, "/") {
		return ErrInvalidPackageName
	}
	
	return nil
}

// validateFilename 验证文件名是否合法
func validateFilename(name string) error {
	if name == "" {
		return ErrInvalidFilename
	}
	
	// 检查路径穿越字符
	if strings.Contains(name, "..") {
		return ErrInvalidFilename
	}
	if strings.Contains(name, "\x00") {
		return ErrInvalidFilename
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return ErrInvalidFilename
	}
	
	return nil
}

func (s *Storage) packageDir(packageName string) (string, error) {
	if err := validatePackageName(packageName); err != nil {
		return "", err
	}
	dir := filepath.Join(s.packagesDir(), packageName)
	if err := s.validatePath(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func (s *Storage) metadataPath(packageName string) (string, error) {
	dir, err := s.packageDir(packageName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "metadata.json"), nil
}

func (s *Storage) tarballsDir(packageName string) (string, error) {
	dir, err := s.packageDir(packageName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tarballs"), nil
}

func (s *Storage) tarballPath(packageName, filename string) (string, error) {
	if err := validateFilename(filename); err != nil {
		return "", err
	}
	dir, err := s.tarballsDir(packageName)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, filename)
	if err := s.validatePath(path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Storage) HasPackage(packageName string) bool {
	path, err := s.metadataPath(packageName)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func (s *Storage) GetMetadata(packageName string) ([]byte, error) {
	path, err := s.metadataPath(packageName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, registry.ErrPackageNotFound
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	
	// 验证 JSON 完整性
	if err := validateMetadataJSON(data); err != nil {
		logger.Warnf("Corrupted metadata for package %s: %v", packageName, err)
		// 如果数据损坏，删除并返回不存在，让上层重新获取
		os.Remove(path)
		return nil, registry.ErrPackageNotFound
	}
	
	return data, nil
}

// validateMetadataJSON 验证元数据 JSON 是否完整
func validateMetadataJSON(data []byte) error {
	var raw json.RawMessage
	return json.Unmarshal(data, &raw)
}

func (s *Storage) SaveMetadata(packageName string, data []byte) error {
	dir, err := s.packageDir(packageName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// 验证 JSON 完整性
	if err := validateMetadataJSON(data); err != nil {
		return fmt.Errorf("invalid metadata JSON: %w", err)
	}

	path, err := s.metadataPath(packageName)
	if err != nil {
		return err
	}
	
	// 原子写入：先写入临时文件，再重命名
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	
	// 重命名临时文件到目标文件（原子操作）
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // 清理临时文件
		return fmt.Errorf("failed to finalize metadata: %w", err)
	}
	
	return nil
}

func (s *Storage) HasTarball(packageName, filename string) bool {
	path, err := s.tarballPath(packageName, filename)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func (s *Storage) GetTarball(packageName, filename string) ([]byte, error) {
	path, err := s.tarballPath(packageName, filename)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, registry.ErrTarballNotFound
		}
		return nil, fmt.Errorf("failed to read tarball: %w", err)
	}
	return data, nil
}

func (s *Storage) SaveTarball(packageName, filename string, data []byte) error {
	dir, err := s.tarballsDir(packageName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tarballs directory: %w", err)
	}

	path, err := s.tarballPath(packageName, filename)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write tarball: %w", err)
	}
	return nil
}

func (s *Storage) SavePackage(packageName string, metadata []byte, tarballs map[string][]byte) error {
	if err := s.SaveMetadata(packageName, metadata); err != nil {
		return err
	}

	for filename, data := range tarballs {
		if err := s.SaveTarball(packageName, filename, data); err != nil {
			return err
		}
	}
	return nil
}

// ListPackages 列出所有包（支持 scoped 包）- 实现 Storage 接口
func (s *Storage) ListPackages() ([]storage.PackageInfo, error) {
	var packages []storage.PackageInfo
	root := s.packagesDir()

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return packages, nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "metadata.json" {
			return nil
		}

		dir := filepath.Dir(path)
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		packageName := filepath.ToSlash(rel)

		info := storage.PackageInfo{
			Name:    packageName,
			Private: true,
		}

		data, err := os.ReadFile(path)
		if err == nil {
			var meta map[string]interface{}
			if json.Unmarshal(data, &meta) == nil {
				if desc, ok := meta["description"].(string); ok {
					info.Description = desc
				}
				if distTags, ok := meta["dist-tag"].(map[string]interface{}); ok {
					if latest, ok := distTags["latest"].(string); ok {
						info.Version = latest
					}
				}
				if distTags, ok := meta["dist-tags"].(map[string]interface{}); ok {
					if latest, ok := distTags["latest"].(string); ok {
						info.Version = latest
					}
				}
				if _, ok := meta["_upstream"]; ok {
					info.Private = false
				}
			}
		}

		// 获取文件修改时间
		if fileInfo, err := d.Info(); err == nil {
			info.UpdatedAt = fileInfo.ModTime()
		}

		packages = append(packages, info)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return packages, nil
}

// GetStats 获取存储统计 - 实现 Storage 接口
func (s *Storage) GetStats() (*storage.StorageStats, error) {
	root := s.packagesDir()

	stats := &storage.StorageStats{}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return stats, nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Name() == "metadata.json" {
			stats.TotalPackages++
		}
		if !d.IsDir() {
			info, statErr := d.Info()
			if statErr == nil {
				stats.TotalSize += info.Size()
			}
		}
		return nil
	})

	return stats, err
}

// ListPackageNames 列出所有包名（内部使用）
func (s *Storage) ListPackageNames() ([]string, error) {
	var packages []string
	root := s.packagesDir()

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return packages, nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "metadata.json" {
			return nil
		}

		dir := filepath.Dir(path)
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		packageName := filepath.ToSlash(rel)
		packages = append(packages, packageName)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return packages, nil
}

// SearchPackages 搜索包
func (s *Storage) SearchPackages(query string) ([]storage.PackageInfo, error) {
	all, err := s.ListPackages()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []storage.PackageInfo
	for _, pkg := range all {
		if strings.Contains(strings.ToLower(pkg.Name), query) ||
			strings.Contains(strings.ToLower(pkg.Description), query) {
			results = append(results, pkg)
		}
	}
	return results, nil
}

func (s *Storage) DeletePackage(packageName string) error {
	dir, err := s.packageDir(packageName)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (s *Storage) DeleteTarball(packageName, filename string) error {
	path, err := s.tarballPath(packageName, filename)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// GetStorageStats 获取存储统计（旧接口，保持兼容）
func (s *Storage) GetStorageStats() (totalPackages int, totalSize int64, err error) {
	stats, err := s.GetStats()
	if err != nil {
		return 0, 0, err
	}
	return int(stats.TotalPackages), stats.TotalSize, nil
}
