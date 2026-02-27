package handler

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

// BackupHandler 备份恢复 Handler
type BackupHandler struct {
	dataDir string
}

// NewBackupHandler 创建 BackupHandler
func NewBackupHandler(dataDir string) *BackupHandler {
	return &BackupHandler{dataDir: dataDir}
}

// BackupInfo 备份信息
type BackupInfo struct {
	TotalPackages int    `json:"totalPackages"`
	StorageSize   int64  `json:"storageSize"`
	DatabaseSize  int64  `json:"databaseSize"`
	DataDir       string `json:"dataDir"`
}

// GetBackupInfo 获取备份信息
// GET /-/api/admin/backup/info
func (h *BackupHandler) GetBackupInfo(c *gin.Context) {
	// 统计包数量
	var totalPackages int64
	db.DB.Model(&db.Package{}).Count(&totalPackages)

	// 计算存储大小
	var storageSize int64
	filepath.Walk(filepath.Join(h.dataDir, "packages"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			storageSize += info.Size()
		}
		return nil
	})

	// 数据库大小
	var dbSize int64
	dbPath := filepath.Join(h.dataDir, "grape.db")
	if info, err := os.Stat(dbPath); err == nil {
		dbSize = info.Size()
	}

	c.JSON(http.StatusOK, BackupInfo{
		TotalPackages: int(totalPackages),
		StorageSize:   storageSize,
		DatabaseSize:  dbSize,
		DataDir:       h.dataDir,
	})
}

// CreateBackup 创建备份并下载
// GET /-/api/admin/backup/download
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "grape-backup-*.tar.gz")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(tmpFile)
	tarWriter := tar.NewWriter(gzWriter)

	// 添加数据目录
	dataPath := filepath.Join(h.dataDir, "packages")
	if _, err := os.Stat(dataPath); err == nil {
		if err := h.addDirectoryToTar(tarWriter, dataPath, "data/packages"); err != nil {
			logger.Warnf("Failed to backup packages: %v", err)
		}
	}

	// 添加数据库
	dbPath := filepath.Join(h.dataDir, "grape.db")
	if _, err := os.Stat(dbPath); err == nil {
		if err := h.addFileToTar(tarWriter, dbPath, "data/grape.db"); err != nil {
			logger.Warnf("Failed to backup database: %v", err)
		}
	}

	// 添加元数据
	metadata := fmt.Sprintf("grape-backup\n%s\n", time.Now().Format(time.RFC3339))
	h.addBytesToTar(tarWriter, "BACKUP-META", []byte(metadata))

	// 关闭 writer
	tarWriter.Close()
	gzWriter.Close()
	tmpFile.Close()

	// 读取文件内容
	fileData, err := os.ReadFile(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read backup file"})
		return
	}

	// 设置下载头
	filename := fmt.Sprintf("grape-backup-%s.tar.gz", time.Now().Format("20060102-150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/gzip", fileData)

	// 记录审计日志
	user := getCurrentUsername(c)
	db.DB.Create(&db.AuditLog{
		Action:   "backup_create",
		Username: user,
		IP:       c.ClientIP(),
		Detail:   "Created backup via Web UI",
	})

	logger.Infof("Backup created by %s (%d bytes)", user, len(fileData))
}

// RestoreBackup 从上传的文件恢复
// POST /-/api/admin/backup/restore
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	// 解压
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup file (not gzip)"})
		return
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// 备份当前数据
	backupDir := h.dataDir + ".restore-backup-" + time.Now().Format("20060102-150405")
	if err := h.backupCurrentData(backupDir); err != nil {
		logger.Warnf("Failed to backup current data: %v", err)
	}

	// 解压文件
	restored := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup file (tar error)"})
			return
		}

		// 只恢复 data 目录下的内容
		if !strings.HasPrefix(header.Name, "data/") {
			continue
		}

		// 去掉 data/ 前缀
		targetPath := filepath.Join(h.dataDir, strings.TrimPrefix(header.Name, "data/"))

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(targetPath, os.FileMode(header.Mode))

		case tar.TypeReg:
			// 确保父目录存在
			os.MkdirAll(filepath.Dir(targetPath), 0755)

			// 写入文件
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				logger.Warnf("Failed to create file %s: %v", targetPath, err)
				continue
			}
			io.Copy(outFile, tarReader)
			outFile.Close()
			restored++
		}
	}

	// 记录审计日志
	user := getCurrentUsername(c)
	db.DB.Create(&db.AuditLog{
		Action:   "backup_restore",
		Username: user,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Restored backup via Web UI (%d files)", restored),
	})

	logger.Infof("Backup restored by %s (%d files)", user, restored)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("恢复成功，共恢复 %d 个文件。请重启服务以应用更改。", restored),
		"restart": true,
	})
}

// ListBackups 列出可用的备份（如果有的话）
// GET /-/api/admin/backup/list
func (h *BackupHandler) ListBackups(c *gin.Context) {
	backups := []map[string]interface{}{}

	// 查找自动备份目录
	entries, err := os.ReadDir(filepath.Dir(h.dataDir))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"backups": backups})
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "data.restore-backup-") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, map[string]interface{}{
				"name":      name,
				"createdAt": info.ModTime().Format(time.RFC3339),
				"auto":      true,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"backups": backups})
}

func (h *BackupHandler) addDirectoryToTar(tw *tar.Writer, srcPath, destPath string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".tmp") {
			return nil
		}

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

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			io.Copy(tw, file)
		}

		return nil
	})
}

func (h *BackupHandler) addFileToTar(tw *tar.Writer, srcPath, destPath string) error {
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

func (h *BackupHandler) addBytesToTar(tw *tar.Writer, name string, data []byte) {
	header := &tar.Header{
		Name:     name,
		Mode:     0644,
		Size:     int64(len(data)),
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	}
	tw.WriteHeader(header)
	tw.Write(data)
}

func (h *BackupHandler) backupCurrentData(backupDir string) error {
	return os.Rename(h.dataDir, backupDir)
}

func getCurrentUsername(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(interface{ GetUsername() string }); ok {
			return u.GetUsername()
		}
	}
	return "unknown"
}
