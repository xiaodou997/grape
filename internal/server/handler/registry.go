package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/registry"
	"github.com/graperegistry/grape/internal/storage/local"
)

type RegistryHandler struct {
	proxy   *registry.Proxy
	storage *local.Storage
	baseURL string
}

func NewRegistryHandler(proxy *registry.Proxy, storage *local.Storage, baseURL string) *RegistryHandler {
	return &RegistryHandler{
		proxy:   proxy,
		storage: storage,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// GetPackage handles GET /:package
func (h *RegistryHandler) GetPackage(c *gin.Context) {
	packageName := c.Param("package")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package name"})
		return
	}

	logger.Debugf("Getting package: %s", packageName)

	// Check local storage first
	if h.storage.HasPackage(packageName) {
		data, err := h.storage.GetMetadata(packageName)
		if err != nil {
			logger.Errorf("Failed to read local metadata: %v", err)
		} else {
			rewritten, err := h.rewriteTarballURLs(data, packageName)
			if err != nil {
				logger.Errorf("Failed to rewrite URLs: %v", err)
				c.Data(http.StatusOK, "application/json", data)
				return
			}
			c.Data(http.StatusOK, "application/json", rewritten)
			return
		}
	}

	// Fetch from upstream
	data, err := h.proxy.GetMetadata(packageName)
	if err != nil {
		if err == registry.ErrPackageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
			return
		}
		logger.Errorf("Failed to fetch from upstream: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch from upstream"})
		return
	}

	// Cache the metadata
	if err := h.storage.SaveMetadata(packageName, data); err != nil {
		logger.Warnf("Failed to cache metadata: %v", err)
	}

	// Rewrite tarball URLs
	rewritten, err := h.rewriteTarballURLs(data, packageName)
	if err != nil {
		logger.Warnf("Failed to rewrite URLs: %v", err)
		c.Data(http.StatusOK, "application/json", data)
		return
	}

	c.Data(http.StatusOK, "application/json", rewritten)
}

// GetTarball handles GET /:package/-/:filename
func (h *RegistryHandler) GetTarball(c *gin.Context) {
	packageName := c.Param("package")
	filename := c.Param("filename")

	if packageName == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	logger.Debugf("Getting tarball: %s/-/%s", packageName, filename)

	// Check local storage first
	if h.storage.HasTarball(packageName, filename) {
		data, err := h.storage.GetTarball(packageName, filename)
		if err != nil {
			logger.Errorf("Failed to read local tarball: %v", err)
		} else {
			c.Data(http.StatusOK, "application/octet-stream", data)
			return
		}
	}

	// Fetch from upstream
	data, err := h.proxy.GetTarball(packageName, filename)
	if err != nil {
		if err == registry.ErrTarballNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "tarball not found"})
			return
		}
		logger.Errorf("Failed to fetch tarball from upstream: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch tarball from upstream"})
		return
	}

	// Cache the tarball
	if err := h.storage.SaveTarball(packageName, filename, data); err != nil {
		logger.Warnf("Failed to cache tarball: %v", err)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (h *RegistryHandler) rewriteTarballURLs(data []byte, packageName string) ([]byte, error) {
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	versions, ok := pkg["versions"].(map[string]interface{})
	if !ok {
		return data, nil
	}

	for version, v := range versions {
		versionData, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		dist, ok := versionData["dist"].(map[string]interface{})
		if !ok {
			continue
		}

		if tarball, ok := dist["tarball"].(string); ok {
			filename := path.Base(tarball)
			dist["tarball"] = fmt.Sprintf("%s/%s/-/%s", h.baseURL, packageName, filename)
			versionData["dist"] = dist
			versions[version] = versionData
		}
	}

	pkg["versions"] = versions
	return json.Marshal(pkg)
}
