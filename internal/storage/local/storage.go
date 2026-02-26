package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/graperegistry/grape/internal/registry"
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

func (s *Storage) packageDir(packageName string) string {
	return filepath.Join(s.packagesDir(), packageName)
}

func (s *Storage) metadataPath(packageName string) string {
	return filepath.Join(s.packageDir(packageName), "metadata.json")
}

func (s *Storage) tarballsDir(packageName string) string {
	return filepath.Join(s.packageDir(packageName), "tarballs")
}

func (s *Storage) tarballPath(packageName, filename string) string {
	return filepath.Join(s.tarballsDir(packageName), filename)
}

func (s *Storage) HasPackage(packageName string) bool {
	_, err := os.Stat(s.metadataPath(packageName))
	return err == nil
}

func (s *Storage) GetMetadata(packageName string) ([]byte, error) {
	path := s.metadataPath(packageName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, registry.ErrPackageNotFound
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	return data, nil
}

func (s *Storage) SaveMetadata(packageName string, data []byte) error {
	dir := s.packageDir(packageName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	path := s.metadataPath(packageName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	return nil
}

func (s *Storage) HasTarball(packageName, filename string) bool {
	_, err := os.Stat(s.tarballPath(packageName, filename))
	return err == nil
}

func (s *Storage) GetTarball(packageName, filename string) ([]byte, error) {
	path := s.tarballPath(packageName, filename)
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
	dir := s.tarballsDir(packageName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tarballs directory: %w", err)
	}

	path := s.tarballPath(packageName, filename)
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

func (s *Storage) ListPackages() ([]string, error) {
	dir := s.packagesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() {
			packages = append(packages, entry.Name())
		}
	}
	return packages, nil
}

func (s *Storage) DeletePackage(packageName string) error {
	dir := s.packageDir(packageName)
	return os.RemoveAll(dir)
}

func (s *Storage) DeleteTarball(packageName, filename string) error {
	path := s.tarballPath(packageName, filename)
	return os.Remove(path)
}