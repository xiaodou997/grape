package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/storage/local"
)

type PublishHandler struct {
	storage *local.Storage
}

func NewPublishHandler(storage *local.Storage) *PublishHandler {
	return &PublishHandler{storage: storage}
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

	logger.Infof("Publishing package: %s by user: %s", packageName, user.Username)

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

	// 构建并保存元数据
	metadata := h.buildMetadata(&req, packageName, user.Username)
	metadataJSON, _ := json.Marshal(metadata)
	
	if err := h.storage.SaveMetadata(packageName, metadataJSON); err != nil {
		logger.Errorf("Failed to save metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save metadata"})
		return
	}

	logger.Infof("Package published: %s", packageName)

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

	if err := h.storage.DeletePackage(packageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PublishHandler) buildMetadata(req *PublishRequest, packageName, publisher string) map[string]interface{} {
	timeMap := make(map[string]string)
	for version := range req.Versions {
		timeMap[version] = time.Now().UTC().Format(time.RFC3339)
	}

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