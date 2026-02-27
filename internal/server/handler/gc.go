package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/storage/local"
	"gorm.io/gorm"
)

// GCHandler Garbage Collection Handler
type GCHandler struct {
	storage  *local.Storage
	dataPath string
}

// NewGCHandler creates a new GC handler
func NewGCHandler(storage *local.Storage, dataPath string) *GCHandler {
	return &GCHandler{
		storage:  storage,
		dataPath: dataPath,
	}
}

// GCStats GC statistics
type GCStats struct {
	TotalPackages      int64 `json:"totalPackages"`
	TotalVersions      int64 `json:"totalVersions"`
	TotalSize          int64 `json:"totalSize"`
	OrphanedFiles      int64 `json:"orphanedFiles"`
	OrphanedSize       int64 `json:"orphanedSize"`
	DeprecatedPackages int64 `json:"deprecatedPackages"`
	OldPackages        int64 `json:"oldPackages"`      // Not accessed for > 180 days
	OldPackagesSize    int64 `json:"oldPackagesSize"`  // Size of old packages
}

// GCCandidate a package candidate for GC
type GCCandidate struct {
	PackageName    string    `json:"packageName"`
	Version        string    `json:"version"`
	LastAccessed   string    `json:"lastAccessed"`
	AccessCount    int       `json:"accessCount"`
	Size           int64     `json:"size"`
	IsDeprecated   bool      `json:"isDeprecated"`
	Reason         string    `json:"reason"`
}

// GetGCStats returns GC statistics
// GET /-/api/admin/gc/stats
func (h *GCHandler) GetGCStats(c *gin.Context) {
	stats := GCStats{}

	// Total packages
	db.DB.Model(&db.Package{}).Count(&stats.TotalPackages)

	// Total versions
	db.DB.Model(&db.PackageVersion{}).Count(&stats.TotalVersions)

	// Total size
	packagesPath := filepath.Join(h.dataPath, "packages")
	filepath.Walk(packagesPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			stats.TotalSize += info.Size()
		}
		return nil
	})

	// Deprecated packages
	db.DB.Model(&db.PackageDeprecation{}).Count(&stats.DeprecatedPackages)

	// Old packages (not accessed for > 180 days)
	cutoff := time.Now().AddDate(0, 0, -180)
	db.DB.Model(&db.PackageGCMetadata{}).
		Where("last_accessed_at < ? OR last_accessed_at IS NULL", cutoff).
		Count(&stats.OldPackages)

	// Calculate old packages size
	var oldPackages []db.PackageGCMetadata
	db.DB.Where("last_accessed_at < ? OR last_accessed_at IS NULL", cutoff).Find(&oldPackages)

	for _, pkg := range oldPackages {
		// Get package size
		packagePath := filepath.Join(packagesPath, strings.ReplaceAll(pkg.PackageName, "@", ""), "metadata.json")
		if info, err := os.Stat(packagePath); err == nil {
			stats.OldPackagesSize += info.Size()
		}
		// Also count tarballs
		tarballPath := filepath.Join(packagesPath, strings.ReplaceAll(pkg.PackageName, "@", ""), "tgz")
		filepath.Walk(tarballPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				stats.OldPackagesSize += info.Size()
			}
			return nil
		})
	}

	// Orphaned files count (approximate)
	stats.OrphanedFiles = 0
	stats.OrphanedSize = 0

	c.JSON(http.StatusOK, stats)
}

// AnalyzeGC analyzes packages for GC candidates
// GET /-/api/admin/gc/analyze
func (h *GCHandler) AnalyzeGC(c *gin.Context) {
	// Get query parameters
	daysStr := c.DefaultQuery("days", "180")
	minVersionsStr := c.DefaultQuery("minVersions", "5")
	includeDeprecated := c.Query("includeDeprecated") == "true"

	cutoff, err := strconv.Atoi(daysStr)
	if err != nil {
		cutoff = 180
	}
	cutoffTime := time.Now().AddDate(0, 0, -cutoff)

	minVersionsInt, err := strconv.Atoi(minVersionsStr)
	if err != nil {
		minVersionsInt = 5
	}

	var candidates []GCCandidate

	// Find old packages
	var oldPackages []db.PackageGCMetadata
	query := db.DB.Where("last_accessed_at < ? OR last_accessed_at IS NULL", cutoffTime)
	if !includeDeprecated {
		query = query.Where("package_name NOT IN (SELECT DISTINCT package_name FROM package_deprecations)")
	}
	query.Find(&oldPackages)

	for _, pkg := range oldPackages {
		// Check version count
		var versionCount int64
		db.DB.Model(&db.PackageVersion{}).
			Where("package_name = ?", pkg.PackageName).
			Count(&versionCount)

		if versionCount <= int64(minVersionsInt) {
			continue // Keep packages with few versions
		}

		// Check if deprecated
		var isDeprecated bool
		var deprecation db.PackageDeprecation
		result := db.DB.Where("package_name = ?", pkg.PackageName).First(&deprecation)
		isDeprecated = result.Error == nil

		// Calculate size
		var size int64
		packageNameSafe := strings.ReplaceAll(pkg.PackageName, "@", "")
		packagePath := filepath.Join(h.dataPath, "packages", packageNameSafe)
		filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				size += info.Size()
			}
			return nil
		})

		lastAccessed := "never"
		if pkg.LastAccessedAt != nil {
			lastAccessed = pkg.LastAccessedAt.Format(time.RFC3339)
		}

		candidates = append(candidates, GCCandidate{
			PackageName:  pkg.PackageName,
			LastAccessed: lastAccessed,
			AccessCount:  pkg.AccessCount,
			Size:         size,
			IsDeprecated: isDeprecated,
			Reason:       fmt.Sprintf("Not accessed for %d+ days", cutoff),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"candidates": candidates,
		"policy": gin.H{
			"maxInactiveDays":   cutoff,
			"minVersionsToKeep": minVersionsInt,
			"includeDeprecated": includeDeprecated,
		},
	})
}

// RunGC runs garbage collection
// POST /-/api/admin/gc/run
func (h *GCHandler) RunGC(c *gin.Context) {
	var req struct {
		DryRun            bool `json:"dryRun"`
		MaxInactiveDays   int  `json:"maxInactiveDays"`
		MinVersionsToKeep int  `json:"minVersionsToKeep"`
		IncludeDeprecated bool `json:"includeDeprecated"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Default values
	if req.MaxInactiveDays == 0 {
		req.MaxInactiveDays = 180
	}
	if req.MinVersionsToKeep == 0 {
		req.MinVersionsToKeep = 5
	}

	cutoffTime := time.Now().AddDate(0, 0, -req.MaxInactiveDays)

	var deleted []string
	var deletedSize int64
	var errors []string

	// Find old packages
	var oldPackages []db.PackageGCMetadata
	query := db.DB.Where("last_accessed_at < ? OR last_accessed_at IS NULL", cutoffTime)
	if !req.IncludeDeprecated {
		query = query.Where("package_name NOT IN (SELECT DISTINCT package_name FROM package_deprecations)")
	}
	query.Find(&oldPackages)

	for _, pkg := range oldPackages {
		// Check version count
		var versionCount int64
		db.DB.Model(&db.PackageVersion{}).
			Where("package_name = ?", pkg.PackageName).
			Count(&versionCount)

		if versionCount <= int64(req.MinVersionsToKeep) {
			continue
		}

		// Calculate size
		packageNameSafe := strings.ReplaceAll(pkg.PackageName, "@", "")
		packagePath := filepath.Join(h.dataPath, "packages", packageNameSafe)

		var size int64
		filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				size += info.Size()
			}
			return nil
		})

		if req.DryRun {
			deleted = append(deleted, pkg.PackageName+" (dry run)")
			deletedSize += size
			continue
		}

		// Actually delete
		if err := os.RemoveAll(packagePath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", pkg.PackageName, err))
			continue
		}

		// Delete from database
		db.DB.Where("package_name = ?", pkg.PackageName).Delete(&db.PackageVersion{})
		db.DB.Where("package_name = ?", pkg.PackageName).Delete(&db.Package{})
		db.DB.Where("package_name = ?", pkg.PackageName).Delete(&db.PackageGCMetadata{})
		db.DB.Where("package_name = ?", pkg.PackageName).Delete(&db.PackageOwner{})

		deleted = append(deleted, pkg.PackageName)
		deletedSize += size
	}

	// Record audit log
	user := getCurrentUsernameFromContext(c)
	action := "gc_run"
	if req.DryRun {
		action = "gc_dry_run"
	}
	db.DB.Create(&db.AuditLog{
		Action:   action,
		Username: user,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Deleted %d packages (%d bytes)", len(deleted), deletedSize),
	})

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"dryRun":       req.DryRun,
		"deleted":      deleted,
		"deletedCount": len(deleted),
		"deletedSize":  deletedSize,
		"errors":       errors,
	})
}

// UpdateAccessTime updates package access time (called when package is accessed)
func UpdateAccessTime(packageName string) {
	now := time.Now()
	var meta db.PackageGCMetadata
	result := db.DB.Where("package_name = ?", packageName).First(&meta)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new record
		meta = db.PackageGCMetadata{
			PackageName:    packageName,
			LastAccessedAt: &now,
			AccessCount:    1,
		}
		db.DB.Create(&meta)
	} else if result.Error == nil {
		// Update existing
		db.DB.Model(&meta).Updates(map[string]interface{}{
			"last_accessed_at": now,
			"access_count":     meta.AccessCount + 1,
			"updated_at":       now,
		})
	}
}

// DeprecatePackage marks a package/version as deprecated
// POST /-/api/admin/packages/:name/deprecate
func (h *GCHandler) DeprecatePackage(c *gin.Context) {
	packageName := c.Param("name")

	var req struct {
		Version string `json:"version"` // Optional: if empty, deprecate entire package
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user := getCurrentUsernameFromContext(c)

	deprecation := db.PackageDeprecation{
		PackageName:   packageName,
		Version:       req.Version,
		Reason:        req.Reason,
		DeprecatedBy:  user,
		DeprecatedAt:  time.Now(),
	}

	if err := db.DB.Create(&deprecation).Error; err != nil {
		// Update if exists
		db.DB.Where("package_name = ? AND (version = ? OR (? = '' AND version IS NULL))",
			packageName, req.Version, req.Version).
			Assign(map[string]interface{}{
				"reason":        req.Reason,
				"deprecated_by": user,
				"deprecated_at": time.Now(),
			}).
			FirstOrCreate(&deprecation)
	}

	db.DB.Create(&db.AuditLog{
		Action:   "package_deprecate",
		Username: user,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Deprecated %s (version: %s): %s", packageName, req.Version, req.Reason),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Package deprecated"})
}

// UndeprecatePackage removes deprecation
// DELETE /-/api/admin/packages/:name/deprecate
func (h *GCHandler) UndeprecatePackage(c *gin.Context) {
	packageName := c.Param("name")
	version := c.Query("version")

	query := db.DB.Where("package_name = ?", packageName)
	if version != "" {
		query = query.Where("version = ?", version)
	}

	query.Delete(&db.PackageDeprecation{})

	user := getCurrentUsernameFromContext(c)
	db.DB.Create(&db.AuditLog{
		Action:   "package_undeprecate",
		Username: user,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Undeprecated %s (version: %s)", packageName, version),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// getCurrentUsernameFromContext gets username from gin context
func getCurrentUsernameFromContext(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(interface{ GetUsername() string }); ok {
			return u.GetUsername()
		}
	}
	return "unknown"
}
