package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/graperegistry/grape/internal/logger"
)

// BackupCommand 备份命令
type BackupCommand struct {
	Output string // 输出文件路径
	List   bool   // 列出备份内容
	Input  string // 用于 list 命令的输入文件
}

// Run 执行备份命令
func (c *BackupCommand) Run() error {
	// 列出备份内容
	if c.List {
		return c.listBackup()
	}

	// 创建备份
	return c.createBackup()
}

func (c *BackupCommand) createBackup() error {
	// 确定输出路径
	outputPath := c.Output
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		outputPath = fmt.Sprintf("grape-backup-%s.tar.gz", timestamp)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 要备份的目录和文件
	dataDir := "data"
	configFile := "configs/config.yaml"

	// 添加数据目录
	if _, err := os.Stat(dataDir); err == nil {
		if err := c.addDirectoryToTar(tarWriter, dataDir, "data"); err != nil {
			return fmt.Errorf("failed to backup data directory: %w", err)
		}
		logger.Infof("✅ Backed up: %s/", dataDir)
	}

	// 添加配置文件
	if _, err := os.Stat(configFile); err == nil {
		if err := c.addFileToTar(tarWriter, configFile, "config.yaml"); err != nil {
			logger.Warnf("Failed to backup config file: %v", err)
		} else {
			logger.Infof("✅ Backed up: %s", configFile)
		}
	}

	// 添加备份元数据
	metadata := fmt.Sprintf("grape-backup\n%s\n", time.Now().Format(time.RFC3339))
	if err := c.addBytesToTar(tarWriter, "BACKUP-META", []byte(metadata)); err != nil {
		logger.Warnf("Failed to write backup metadata: %v", err)
	}

	// 确保 tar 和 gzip 完全写入
	tarWriter.Close()
	gzWriter.Close()

	logger.Infof("🎉 Backup created: %s", outputPath)
	return nil
}

func (c *BackupCommand) addDirectoryToTar(tw *tar.Writer, srcPath, destPath string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过临时文件
		if strings.HasSuffix(path, ".tmp") {
			return nil
		}

		// 创建 tar header
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		tarPath := filepath.Join(destPath, relPath)

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = tarPath

		// 写入 header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 如果是普通文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			return err
		}

		return nil
	})
}

func (c *BackupCommand) addFileToTar(tw *tar.Writer, srcPath, destPath string) error {
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = destPath

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func (c *BackupCommand) addBytesToTar(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name:     name,
		Mode:     0644,
		Size:     int64(len(data)),
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write(data)
	return err
}

func (c *BackupCommand) listBackup() error {
	if c.Input == "" {
		return fmt.Errorf("input file required for list command (--input)")
	}

	file, err := os.Open(c.Input)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	fmt.Printf("Backup contents: %s\n\n", c.Input)
	fmt.Printf("%-12s %-40s %s\n", "SIZE", "PATH", "MODIFIED")
	fmt.Println(strings.Repeat("-", 80))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		size := "DIR"
		if header.Typeflag == tar.TypeReg {
			size = formatSize(header.Size)
		}

		fmt.Printf("%-12s %-40s %s\n", size, header.Name, header.ModTime.Format("2006-01-02 15:04"))
	}

	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RestoreCommand 恢复命令
type RestoreCommand struct {
	Input string // 输入备份文件路径
	Force bool   // 强制覆盖
}

// Run 执行恢复命令
func (c *RestoreCommand) Run() error {
	if c.Input == "" {
		return fmt.Errorf("input file required (--input)")
	}

	// 检查输入文件
	file, err := os.Open(c.Input)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// 解压
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// 检查数据目录是否存在
	dataDir := "data"
	if _, err := os.Stat(dataDir); err == nil && !c.Force {
		return fmt.Errorf("data directory already exists, use --force to overwrite")
	}

	logger.Infof("📦 Restoring from: %s", c.Input)

	// 解压文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// 跳过特殊文件
		if header.Name == "BACKUP-META" || header.Name == "config.yaml" {
			continue
		}

		// 只恢复 data 目录
		if !strings.HasPrefix(header.Name, "data/") {
			continue
		}

		targetPath := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}

		case tar.TypeReg:
			// 确保父目录存在
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// 创建文件
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			outFile.Close()
		}
	}

	logger.Info("✅ Restore completed")
	logger.Info("⚠️  Please restart Grape server to apply changes")

	return nil
}
