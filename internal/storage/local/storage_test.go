package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_HasPackage(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "grape-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := New(tmpDir)

	// 测试不存在的包
	if storage.HasPackage("nonexistent") {
		t.Fatal("Expected HasPackage to return false for nonexistent package")
	}

	// 创建测试包
	pkgDir := filepath.Join(tmpDir, "packages", "test-package")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "metadata.json"), []byte(`{"name":"test-package"}`), 0644)

	// 测试存在的包
	if !storage.HasPackage("test-package") {
		t.Fatal("Expected HasPackage to return true for existing package")
	}
}

func TestStorage_Metadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grape-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := New(tmpDir)

	// 测试保存元数据
	metadata := []byte(`{"name":"test-package","version":"1.0.0"}`)
	err = storage.SaveMetadata("test-package", metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// 测试读取元数据
	data, err := storage.GetMetadata("test-package")
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if string(data) != string(metadata) {
		t.Fatalf("Metadata mismatch: expected %s, got %s", string(metadata), string(data))
	}
}

func TestStorage_Tarball(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grape-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := New(tmpDir)

	// 测试保存 tarball
	tarball := []byte("tarball content")
	err = storage.SaveTarball("test-package", "test-package-1.0.0.tgz", tarball)
	if err != nil {
		t.Fatalf("Failed to save tarball: %v", err)
	}

	// 测试读取 tarball
	data, err := storage.GetTarball("test-package", "test-package-1.0.0.tgz")
	if err != nil {
		t.Fatalf("Failed to get tarball: %v", err)
	}

	if string(data) != string(tarball) {
		t.Fatal("Tarball content mismatch")
	}
}

func TestStorage_ListPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grape-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := New(tmpDir)

	// 创建多个包
	for _, name := range []string{"package-a", "@scope/package-b"} {
		pkgDir := filepath.Join(tmpDir, "packages", name)
		os.MkdirAll(pkgDir, 0755)
		os.WriteFile(filepath.Join(pkgDir, "metadata.json"), []byte(`{}`), 0644)
	}

	// 测试列表
	packages, err := storage.ListPackages()
	if err != nil {
		t.Fatalf("Failed to list packages: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("Expected 2 packages, got: %d", len(packages))
	}

	// 验证包含 scoped 包
	found := false
	for _, pkg := range packages {
		if pkg.Name == "@scope/package-b" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Scoped package not found in list")
	}
}

func TestStorage_DeletePackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grape-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := New(tmpDir)

	// 创建包
	storage.SaveMetadata("to-delete", []byte(`{}`))

	// 测试删除
	err = storage.DeletePackage("to-delete")
	if err != nil {
		t.Fatalf("Failed to delete package: %v", err)
	}

	// 验证已删除
	if storage.HasPackage("to-delete") {
		t.Fatal("Package should be deleted")
	}
}
