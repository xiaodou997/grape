package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/storage/local"
)

type APIHandler struct {
	storage   *local.Storage
	dataPath  string
}

func NewAPIHandler(storage *local.Storage, dataPath string) *APIHandler {
	return &APIHandler{
		storage:  storage,
		dataPath: dataPath,
	}
}

// PackageInfo 包信息
type PackageInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Private     bool   `json:"private"`
	UpdatedAt   string `json:"updatedAt"`
}

// ListPackages 列出所有包
// GET /-/api/packages
func (h *APIHandler) ListPackages(c *gin.Context) {
	packages, err := h.storage.ListPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list packages"})
		return
	}

	result := make([]PackageInfo, 0, len(packages))
	for _, pkgName := range packages {
		info := PackageInfo{
			Name:    pkgName,
			Private: true, // 本地存储的都是私有包或缓存的包
		}

		// 尝试获取更多元数据
		if data, err := h.storage.GetMetadata(pkgName); err == nil {
			var meta map[string]interface{}
			if json.Unmarshal(data, &meta) == nil {
				if desc, ok := meta["description"].(string); ok {
					info.Description = desc
				}
				if distTags, ok := meta["dist-tags"].(map[string]interface{}); ok {
					if latest, ok := distTags["latest"].(string); ok {
						info.Version = latest
					}
				}
			}
		}

		result = append(result, info)
	}

	c.JSON(http.StatusOK, gin.H{"packages": result})
}

// GetStats 获取统计信息
// GET /-/api/stats
func (h *APIHandler) GetStats(c *gin.Context) {
	packages, _ := h.storage.ListPackages()
	
	stats := gin.H{
		"totalPackages": len(packages),
		"storageSize":   h.calculateStorageSize(),
	}

	c.JSON(http.StatusOK, stats)
}

func (h *APIHandler) calculateStorageSize() int64 {
	var size int64
	packagesDir := filepath.Join(h.dataPath, "packages")
	
	filepath.Walk(packagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size / (1024 * 1024) // 转换为 MB
}

// SearchPackages 搜索包
// GET /-/api/search?q=keyword
func (h *APIHandler) SearchPackages(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	packages, _ := h.storage.ListPackages()
	result := make([]PackageInfo, 0)

	for _, pkgName := range packages {
		if strings.Contains(strings.ToLower(pkgName), query) {
			info := PackageInfo{Name: pkgName}
			
			if data, err := h.storage.GetMetadata(pkgName); err == nil {
				var meta map[string]interface{}
				if json.Unmarshal(data, &meta) == nil {
					if desc, ok := meta["description"].(string); ok {
						info.Description = desc
					}
				}
			}
			
			result = append(result, info)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"packages": result,
		"total":    len(result),
	})
}
