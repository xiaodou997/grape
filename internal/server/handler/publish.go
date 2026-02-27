package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/storage/local"
	"github.com/graperegistry/grape/internal/webhook"
)

type PublishHandler struct {
	storage    *local.Storage
	locks      sync.Map // package name -> *sync.Mutex
	dispatcher *webhook.Dispatcher
}

func NewPublishHandler(storage *local.Storage, dispatcher *webhook.Dispatcher) *PublishHandler {
	return &PublishHandler{storage: storage, dispatcher: dispatcher}
}

// getPackageLock 获取包级别的互斥锁
func (h *PublishHandler) getPackageLock(name string) *sync.Mutex {
	mu, _ := h.locks.LoadOrStore(name, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// releasePackageLock 释放并删除包锁
func (h *PublishHandler) releasePackageLock(name string) {
	h.locks.Delete(name)
}

// PublishRequest npm publish 请求格式
type PublishRequest struct {
	ID            string                            `json:"_id"`
	Name          string                            `json:"name"`
	Description   string                            `json:"description"`
	DistTags      map[string]string                 `json:"dist-tags"`
	Versions      map[string]map[string]interface{} `json:"versions"`
	Access        string                            `json:"access"`
	Attachments   map[string]Attachment             `json:"_attachments"`
	Readme        string                            `json:"readme"`
	Rev           string                            `json:"_rev,omitempty"`
}

// Attachment tarball 附件
type Attachment struct {
	ContentType string `json:"content_type"`
	Data        string `json:"data"`
	Length      int    `json:"length"`
}

// Publish 处理 npm publish 请求
// PUT /:package
func (h *PublishHandler) Publish(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// 检查是否使用只读 Token
	tokenInfo := auth.GetTokenInfo(c)
	if tokenInfo != nil && tokenInfo.Readonly {
		c.JSON(http.StatusForbidden, gin.H{"error": "read-only token cannot publish"})
		return
	}

	packageName := decodePackageName(c.Param("package"))
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package name required"})
		return
	}

	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Failed to parse publish request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name != "" && req.Name != packageName {
		packageName = req.Name
	}

	// 获取包级别锁，防止并发发布冲突
	lock := h.getPackageLock(packageName)
	lock.Lock()
	defer func() {
		lock.Unlock()
		h.releasePackageLock(packageName) // 发布完成后释放锁
	}()

	logger.Infof("Publishing package: %s by user: %s", packageName, user.Username)

	// 检查包所有权（访问控制）
	isNewPackage := !h.storage.HasPackage(packageName)
	if !isNewPackage {
		// 已有包：检查用户是否有权限发布
		if !h.canUserPublishPackage(packageName, user) {
			logger.Warnf("User %s is not allowed to publish package %s", user.Username, packageName)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "you are not an owner of this package",
			})
			return
		}
	}

	// 读取现有元数据并合并
	var existingMeta map[string]interface{}
	if h.storage.HasPackage(packageName) {
		existingData, err := h.storage.GetMetadata(packageName)
		if err == nil {
			if err := json.Unmarshal(existingData, &existingMeta); err != nil {
				logger.Warnf("Failed to parse existing metadata: %v", err)
			}
		}
	}

	// 检查版本是否已存在
	if existingMeta != nil {
		if versions, ok := existingMeta["versions"].(map[string]interface{}); ok {
			for newVersion := range req.Versions {
				if _, exists := versions[newVersion]; exists {
					logger.Warnf("Version already exists: %s@%s", packageName, newVersion)
					c.JSON(http.StatusConflict, gin.H{
						"error": fmt.Sprintf("version %s already exists", newVersion),
					})
					return
				}
			}
		}
	}

	// 处理 tarball 文件
	for filename, attachment := range req.Attachments {
		tarballData, err := base64.StdEncoding.DecodeString(attachment.Data)
		if err != nil {
			logger.Errorf("Failed to decode tarball: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tarball data"})
			return
		}

		if err := h.storage.SaveTarball(packageName, filename, tarballData); err != nil {
			logger.Errorf("Failed to save tarball: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save tarball"})
			return
		}

		logger.Infof("Saved tarball: %s (%d bytes)", filename, len(tarballData))
	}

	// 构建并合并元数据
	metadata := h.mergeMetadata(existingMeta, &req, packageName, user.Username)
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		logger.Errorf("Failed to marshal metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process package metadata"})
		return
	}
	
	if err := h.storage.SaveMetadata(packageName, metadataJSON); err != nil {
		logger.Errorf("Failed to save metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save metadata"})
		return
	}

	// 如果是新包，自动将发布者设为 owner
	if isNewPackage {
		if err := h.addPackageOwner(packageName, user); err != nil {
			logger.Warnf("Failed to add package owner: %v", err)
		} else {
			logger.Infof("Added %s as owner of %s", user.Username, packageName)
		}
	}

	logger.Infof("Package published: %s", packageName)
	db.RecordAudit("package_publish", user.Username, c.ClientIP(), "发布包: "+packageName)

	h.dispatcher.Dispatch(webhook.EventPackagePublished, gin.H{
		"package":   packageName,
		"publisher": user.Username,
		"versions":  req.DistTags,
	})

	c.JSON(http.StatusCreated, gin.H{
		"ok":      true,
		"rev":     fmt.Sprintf("%d-%s", len(req.Versions), packageName),
		"success": true,
	})
}

// Unpublish 处理 npm unpublish 请求
func (h *PublishHandler) Unpublish(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	packageName := decodePackageName(c.Param("package"))
	filename := c.Param("filename")

	// 权限检查：只有 admin 或包维护者可以删除
	if user.Role != "admin" {
		// 检查是否为包维护者
		if !h.isPackageMaintainer(packageName, user.Username) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions to unpublish this package",
			})
			return
		}
	}

	// 获取包级别锁
	lock := h.getPackageLock(packageName)
	lock.Lock()
	defer func() {
		lock.Unlock()
		h.releasePackageLock(packageName)
	}()

	if filename != "" {
		logger.Infof("Unpublishing tarball: %s/-/%s by user: %s", packageName, filename, user.Username)
		
		if err := h.storage.DeleteTarball(packageName, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tarball"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	logger.Infof("Unpublishing package: %s by user: %s", packageName, user.Username)
	db.RecordAudit("package_unpublish", user.Username, c.ClientIP(), "删除包: "+packageName)

	if err := h.storage.DeletePackage(packageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete package"})
		return
	}

	h.dispatcher.Dispatch(webhook.EventPackageUnpublished, gin.H{
		"package":  packageName,
		"operator": user.Username,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// isPackageMaintainer 检查用户是否为包的维护者
func (h *PublishHandler) isPackageMaintainer(packageName, username string) bool {
	if !h.storage.HasPackage(packageName) {
		return false
	}

	data, err := h.storage.GetMetadata(packageName)
	if err != nil {
		return false
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}

	maintainers, ok := meta["maintainers"].([]interface{})
	if !ok {
		return false
	}

	for _, m := range maintainers {
		if maintainer, ok := m.(map[string]interface{}); ok {
			if name, ok := maintainer["name"].(string); ok && name == username {
				return true
			}
		}
	}

	return false
}

// mergeMetadata 合并新旧元数据
func (h *PublishHandler) mergeMetadata(existing map[string]interface{}, req *PublishRequest, packageName, publisher string) map[string]interface{} {
	if existing == nil {
		return h.buildMetadata(req, packageName, publisher)
	}

	// 合并 versions
	if existingVersions, ok := existing["versions"].(map[string]interface{}); ok {
		for ver, info := range req.Versions {
			existingVersions[ver] = info
		}
	} else {
		existing["versions"] = req.Versions
	}

	// 合并 dist-tags（保留旧的，更新新的）
	if existingTags, ok := existing["dist-tags"].(map[string]interface{}); ok {
		for tag, ver := range req.DistTags {
			existingTags[tag] = ver
		}
	} else {
		existing["dist-tags"] = req.DistTags
	}

	// 合并 time
	if existingTime, ok := existing["time"].(map[string]interface{}); ok {
		for version := range req.Versions {
			existingTime[version] = time.Now().UTC().Format(time.RFC3339)
		}
		existingTime["modified"] = time.Now().UTC().Format(time.RFC3339)
	} else {
		timeMap := make(map[string]string)
		for version := range req.Versions {
			timeMap[version] = time.Now().UTC().Format(time.RFC3339)
		}
		timeMap["modified"] = time.Now().UTC().Format(time.RFC3339)
		existing["time"] = timeMap
	}

	// 更新描述（如果有新值）
	if req.Description != "" {
		existing["description"] = req.Description
	}

	// 更新 README（如果有新值）
	if req.Readme != "" {
		existing["readme"] = req.Readme
	}

	// 添加维护者（如果不存在）
	maintainers, _ := existing["maintainers"].([]interface{})
	maintainerExists := false
	for _, m := range maintainers {
		if m, ok := m.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok && name == publisher {
				maintainerExists = true
				break
			}
		}
	}
	if !maintainerExists {
		existing["maintainers"] = append(maintainers, map[string]string{"name": publisher})
	}

	return existing
}

func (h *PublishHandler) buildMetadata(req *PublishRequest, packageName, publisher string) map[string]interface{} {
	timeMap := make(map[string]string)
	for version := range req.Versions {
		timeMap[version] = time.Now().UTC().Format(time.RFC3339)
	}
	timeMap["created"] = time.Now().UTC().Format(time.RFC3339)
	timeMap["modified"] = time.Now().UTC().Format(time.RFC3339)

	return map[string]interface{}{
		"_id":          packageName,
		"name":         packageName,
		"description":  req.Description,
		"dist-tags":    req.DistTags,
		"versions":     req.Versions,
		"readme":       req.Readme,
		"time":         timeMap,
		"_attachments": map[string]interface{}{},
		"maintainers": []map[string]string{
			{"name": publisher},
		},
	}
}

func decodePackageName(name string) string {
	name = strings.ReplaceAll(name, "%2F", "/")
	name = strings.ReplaceAll(name, "%40", "@")
	return name
}

// canUserPublishPackage 检查用户是否有权限发布包
func (h *PublishHandler) canUserPublishPackage(packageName string, user *auth.User) bool {
	// admin 可以发布任何包
	if user.Role == "admin" {
		return true
	}

	// 检查是否有 owner 记录
	var owners []db.PackageOwner
	if err := db.DB.Where("package_name = ?", packageName).Preload("User").Find(&owners).Error; err != nil {
		logger.Warnf("Failed to check package owners: %v", err)
		// 如果查询失败，允许发布（向后兼容）
		return true
	}

	// 如果没有 owner 记录，允许任何 developer 发布（向后兼容）
	if len(owners) == 0 {
		return true
	}

	// 检查用户是否在 owner 列表中
	for _, owner := range owners {
		if owner.User != nil && owner.User.Username == user.Username && owner.CanPublish {
			return true
		}
	}

	return false
}

// addPackageOwner 添加包 owner
func (h *PublishHandler) addPackageOwner(packageName string, user *auth.User) error {
	owner := &db.PackageOwner{
		PackageName: packageName,
		UserID:      user.ID,
		CanPublish:  true,
	}

	if err := db.DB.Create(owner).Error; err != nil {
		// 忽略重复键错误
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") &&
			!strings.Contains(err.Error(), "Duplicate entry") {
			return err
		}
	}

	return nil
}