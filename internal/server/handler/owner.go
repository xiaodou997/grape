package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

// OwnerHandler 包 Owner 管理 Handler
type OwnerHandler struct{}

// NewOwnerHandler 创建 OwnerHandler
func NewOwnerHandler() *OwnerHandler {
	return &OwnerHandler{}
}

// OwnerInfo owner 信息响应
type OwnerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// ListOwners 列出包的 owners
// GET /-/package/:name/collaborators
func (h *OwnerHandler) ListOwners(c *gin.Context) {
	packageName := c.Param("name")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package name required"})
		return
	}

	// 查询包的 owners
	var owners []db.PackageOwner
	if err := db.DB.Where("package_name = ?", packageName).Preload("User").Find(&owners).Error; err != nil {
		logger.Errorf("Failed to list owners: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list owners"})
		return
	}

	// 构建响应（npm 格式）
	result := make(map[string]interface{})
	for _, owner := range owners {
		if owner.User != nil {
			result[owner.User.Username] = map[string]interface{}{
				"name":  owner.User.Username,
				"email": owner.User.Email,
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

// AddOwnerRequest 添加 owner 请求
type AddOwnerRequest struct {
	Name string `json:"name"` // 用户名
}

// AddOwner 添加包 owner
// PUT /-/package/:name/collaborators/:username
func (h *OwnerHandler) AddOwner(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	packageName := c.Param("name")
	username := c.Param("username")

	if packageName == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package name and username required"})
		return
	}

	// 检查权限：只有 admin 或包的 owner 可以添加 owner
	if currentUser.Role != "admin" {
		var existingOwner db.PackageOwner
		result := db.DB.Where("package_name = ? AND user_id = ?", packageName, currentUser.ID).First(&existingOwner)
		if result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only package owners can add collaborators"})
			return
		}
	}

	// 查找要添加的用户
	var targetUser db.User
	if err := db.DB.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// 添加 owner
	owner := &db.PackageOwner{
		PackageName: packageName,
		UserID:      targetUser.ID,
		CanPublish:  true,
	}

	if err := db.DB.Create(owner).Error; err != nil {
		// 检查是否已存在
		if err.Error() != "" {
			logger.Warnf("Owner already exists or error: %v", err)
		}
		// 幂等操作，已存在也返回成功
	}

	// 记录审计日志
	db.DB.Create(&db.AuditLog{
		Action:   "package_owner_add",
		Username: currentUser.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Added %s as owner of %s", username, packageName),
	})

	logger.Infof("User %s added %s as owner of %s", currentUser.Username, username, packageName)

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"name": username,
		"email": targetUser.Email,
	})
}

// RemoveOwner 移除包 owner
// DELETE /-/package/:name/collaborators/:username
func (h *OwnerHandler) RemoveOwner(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	packageName := c.Param("name")
	username := c.Param("username")

	if packageName == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package name and username required"})
		return
	}

	// 检查权限：只有 admin 或包的 owner 可以移除 owner
	if currentUser.Role != "admin" {
		var existingOwner db.PackageOwner
		result := db.DB.Where("package_name = ? AND user_id = ?", packageName, currentUser.ID).First(&existingOwner)
		if result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only package owners can remove collaborators"})
			return
		}
	}

	// 查找要移除的用户
	var targetUser db.User
	if err := db.DB.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// 不能移除自己（除非是 admin）
	if currentUser.ID == targetUser.ID && currentUser.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove yourself as owner"})
		return
	}

	// 移除 owner
	result := db.DB.Where("package_name = ? AND user_id = ?", packageName, targetUser.ID).Delete(&db.PackageOwner{})
	if result.Error != nil {
		logger.Errorf("Failed to remove owner: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove owner"})
		return
	}

	// 记录审计日志
	db.DB.Create(&db.AuditLog{
		Action:   "package_owner_remove",
		Username: currentUser.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Removed %s as owner of %s", username, packageName),
	})

	logger.Infof("User %s removed %s as owner of %s", currentUser.Username, username, packageName)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListPackageOwnersAdmin 管理员列出包 owners（管理后台用）
// GET /-/api/admin/packages/:name/owners
func (h *OwnerHandler) ListPackageOwnersAdmin(c *gin.Context) {
	packageName := c.Param("name")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package name required"})
		return
	}

	var owners []db.PackageOwner
	if err := db.DB.Where("package_name = ?", packageName).Preload("User").Find(&owners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list owners"})
		return
	}

	result := make([]map[string]interface{}, 0)
	for _, owner := range owners {
		if owner.User != nil {
			result = append(result, map[string]interface{}{
				"id":       owner.ID,
				"username": owner.User.Username,
				"email":    owner.User.Email,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"owners": result})
}

// SetPackageOwnerAdmin 管理员设置包 owner
// POST /-/api/admin/packages/:name/owners
func (h *OwnerHandler) SetPackageOwnerAdmin(c *gin.Context) {
	packageName := c.Param("name")
	var req AddOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	// 查找用户
	var targetUser db.User
	if err := db.DB.Where("username = ?", req.Name).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// 添加 owner
	owner := &db.PackageOwner{
		PackageName: packageName,
		UserID:      targetUser.ID,
		CanPublish:  true,
	}

	if err := db.DB.Create(owner).Error; err != nil {
		logger.Warnf("Owner already exists or error: %v", err)
	}

	currentUser := auth.GetCurrentUser(c)
	db.DB.Create(&db.AuditLog{
		Action:   "package_owner_add",
		Username: currentUser.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Admin added %s as owner of %s", req.Name, packageName),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RemovePackageOwnerAdmin 管理员移除包 owner
// DELETE /-/api/admin/packages/:name/owners/:username
func (h *OwnerHandler) RemovePackageOwnerAdmin(c *gin.Context) {
	packageName := c.Param("name")
	username := c.Param("username")

	var targetUser db.User
	if err := db.DB.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	db.DB.Where("package_name = ? AND user_id = ?", packageName, targetUser.ID).Delete(&db.PackageOwner{})

	currentUser := auth.GetCurrentUser(c)
	db.DB.Create(&db.AuditLog{
		Action:   "package_owner_remove",
		Username: currentUser.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Admin removed %s as owner of %s", username, packageName),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
