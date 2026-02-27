package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/registry"
	"github.com/graperegistry/grape/internal/storage"
	"github.com/graperegistry/grape/internal/storage/local"
)

type APIHandler struct {
	storage    storage.Storage
	dataPath   string
	local      *local.Storage
	proxy      *registry.Proxy
	cfg        *config.Config
	version    string
	startTime  time.Time
	applyFn    func(*config.Config)
}

func NewAPIHandler(localStorage *local.Storage, dataPath string, proxy *registry.Proxy, cfg *config.Config, version string, applyFn func(*config.Config)) *APIHandler {
	return &APIHandler{
		storage:   localStorage,
		dataPath:  dataPath,
		local:     localStorage,
		proxy:     proxy,
		cfg:       cfg,
		version:   version,
		startTime: time.Now(),
		applyFn:   applyFn,
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
	for _, pkg := range packages {
		info := PackageInfo{
			Name:        pkg.Name,
			Description: pkg.Description,
			Version:     pkg.Version,
			Private:     pkg.Private,
			UpdatedAt:   pkg.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		result = append(result, info)
	}

	c.JSON(http.StatusOK, gin.H{"packages": result})
}

// GetStats 获取统计信息
// GET /-/api/stats
func (h *APIHandler) GetStats(c *gin.Context) {
	stats, err := h.storage.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	// 转换为 MB
	storageSizeMB := stats.TotalSize / (1024 * 1024)

	result := gin.H{
		"totalPackages": stats.TotalPackages,
		"storageSize":   storageSizeMB,
		"upstreams":     h.proxy.Upstreams(),
	}

	c.JSON(http.StatusOK, result)
}

// GetUpstreams 获取上游配置
// GET /-/api/upstreams
func (h *APIHandler) GetUpstreams(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"upstreams": h.proxy.Upstreams(),
	})
}

// SearchPackages 搜索包
// GET /-/api/search?q=keyword
func (h *APIHandler) SearchPackages(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	packages, err := h.storage.ListPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list packages"})
		return
	}

	result := make([]PackageInfo, 0)
	for _, pkg := range packages {
		if strings.Contains(strings.ToLower(pkg.Name), query) ||
			strings.Contains(strings.ToLower(pkg.Description), query) {
			info := PackageInfo{
				Name:        pkg.Name,
				Description: pkg.Description,
				Version:     pkg.Version,
				Private:     pkg.Private,
			}
			result = append(result, info)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"packages": result,
		"total":    len(result),
	})
}

// GetSystemInfo 获取系统信息
// GET /-/api/admin/system
func (h *APIHandler) GetSystemInfo(c *gin.Context) {
	uptime := time.Since(h.startTime)
	// 格式化 uptime
	uptimeStr := formatDuration(uptime)

	c.JSON(http.StatusOK, gin.H{
		"version":      h.version,
		"startTime":    h.startTime.UTC().Format(time.RFC3339),
		"uptime":       uptimeStr,
		"storagePath":  h.cfg.Storage.Path,
		"databasePath": h.cfg.Database.DSN,
		"host":         fmt.Sprintf("%s:%d", h.cfg.Server.Host, h.cfg.Server.Port),
		"upstreams":    h.proxy.Upstreams(),
	})
}

// GetConfig 获取可编辑配置
// GET /-/api/admin/config
func (h *APIHandler) GetConfig(c *gin.Context) {
	// JWT 密钥脱敏显示
	jwtSecretMasked := "***"
	if len(h.cfg.Auth.JWTSecret) > 0 {
		jwtSecretMasked = "***"
	}

	upstreams := make([]gin.H, 0, len(h.cfg.Registry.Upstreams))
	for _, u := range h.cfg.Registry.Upstreams {
		upstreams = append(upstreams, gin.H{
			"name":    u.Name,
			"url":     u.URL,
			"scope":   u.Scope,
			"timeout": int(u.Timeout.Seconds()),
			"enabled": u.Enabled,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"registry": gin.H{
			"upstream":  h.cfg.Registry.Upstream,
			"upstreams": upstreams,
		},
		"auth": gin.H{
			"jwtSecret":         jwtSecretMasked,
			"jwtExpiry":         int(h.cfg.Auth.JWTExpiry.Hours()),
			"allowRegistration": h.cfg.Auth.AllowRegistration,
		},
		"log": gin.H{
			"level": h.cfg.Log.Level,
		},
	})
}

// UpdateConfigRequest PUT /-/api/admin/config 请求体
type UpdateConfigRequest struct {
	Registry *struct {
		Upstream  string           `json:"upstream"`
		Upstreams []UpstreamCfgReq `json:"upstreams"`
	} `json:"registry"`
	Auth *struct {
		JWTSecret         string `json:"jwtSecret"`
		JWTExpiry         int    `json:"jwtExpiry"` // 小时
		AllowRegistration *bool  `json:"allowRegistration"`
	} `json:"auth"`
	Log *struct {
		Level string `json:"level"`
	} `json:"log"`
}

// UpstreamCfgReq 上游配置请求
type UpstreamCfgReq struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Scope   string `json:"scope"`
	Timeout int    `json:"timeout"` // 秒
	Enabled bool   `json:"enabled"`
}

// UpdateConfig 保存配置并热加载
// PUT /-/api/admin/config
func (h *APIHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 更新 registry 配置
	if req.Registry != nil {
		if req.Registry.Upstream != "" {
			h.cfg.Registry.Upstream = req.Registry.Upstream
		}
		if len(req.Registry.Upstreams) > 0 {
			upstreams := make([]config.UpstreamConfig, 0, len(req.Registry.Upstreams))
			for _, u := range req.Registry.Upstreams {
				timeout := time.Duration(u.Timeout) * time.Second
				if timeout == 0 {
					timeout = 30 * time.Second
				}
				upstreams = append(upstreams, config.UpstreamConfig{
					Name:    u.Name,
					URL:     u.URL,
					Scope:   u.Scope,
					Timeout: timeout,
					Enabled: u.Enabled,
				})
			}
			h.cfg.Registry.Upstreams = upstreams
		}
	}

	// 更新 auth 配置
	if req.Auth != nil {
		// JWT 密钥：只有不是 *** 时才更新
		if req.Auth.JWTSecret != "" && req.Auth.JWTSecret != "***" {
			h.cfg.Auth.JWTSecret = req.Auth.JWTSecret
		}
		if req.Auth.JWTExpiry > 0 {
			h.cfg.Auth.JWTExpiry = time.Duration(req.Auth.JWTExpiry) * time.Hour
		}
		if req.Auth.AllowRegistration != nil {
			h.cfg.Auth.AllowRegistration = *req.Auth.AllowRegistration
		}
	}

	// 更新 log 配置
	if req.Log != nil && req.Log.Level != "" {
		h.cfg.Log.Level = req.Log.Level
	}

	// 持久化到配置文件
	if err := config.Save(h.cfg); err != nil {
		logger.Warnf("Failed to save config to file: %v", err)
		// 不返回错误，继续热更新
	}

	// 热更新运行中的组件
	if h.applyFn != nil {
		h.applyFn(h.cfg)
	}

	// 记录审计日志
	currentUser := auth.GetCurrentUser(c)
	username := ""
	if currentUser != nil {
		username = currentUser.Username
	}
	db.RecordAudit("config_update", username, c.ClientIP(), "系统配置已更新")

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "配置已保存，即时生效"})
}

// formatDuration 格式化运行时长
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
