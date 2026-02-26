package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/logger"
)

type AuthHandler struct {
	userStore  auth.UserStore
	jwtService *auth.JWTService
}

func NewAuthHandler(userStore auth.UserStore, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userStore:  userStore,
		jwtService: jwtService,
	}
}

// LoginRequest npm login 请求格式
type LoginRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Type     string `json:"type"`
}

// LoginResponse npm login 响应格式
type LoginResponse struct {
	OK    bool   `json:"ok"`
	ID    string `json:"id"`
	Rev   string `json:"rev,omitempty"`
	Token string `json:"token,omitempty"`
}

// Login 处理 npm login 请求
// PUT /-/user/org.couchdb.user:{username}
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Debugf("Failed to parse login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 获取用户名 (可能在 name 或 username 字段)
	username := req.Name
	if username == "" {
		username = req.Username
	}
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	// 从 URL 提取用户名
	pathUsername := extractUsernameFromPath(c.Param("username"))
	if pathUsername != "" && pathUsername != username {
		username = pathUsername
	}

	// 检查是登录还是注册
	existingUser, err := h.userStore.Get(username)
	if err == auth.ErrUserNotFound {
		// 用户不存在，创建新用户
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
			return
		}
		
		newUser := &auth.User{
			Username: username,
			Password: req.Password,
			Email:    req.Email,
			Role:     "developer",
		}
		
		if err := h.userStore.Create(newUser); err != nil {
			logger.Errorf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		
		existingUser = newUser
		logger.Infof("New user registered: %s", username)
	} else if err != nil {
		logger.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// 验证密码
	if req.Password != "" {
		validatedUser, err := h.userStore.Validate(username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		existingUser = validatedUser
	}

	// 更新最后登录时间
	now := time.Now()
	existingUser.LastLogin = &now
	_ = h.userStore.Update(existingUser)

	// 生成 JWT token
	token, err := h.jwtService.GenerateToken(existingUser)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	logger.Infof("User logged in: %s", username)

	c.JSON(http.StatusOK, LoginResponse{
		OK:    true,
		ID:    fmt.Sprintf("org.couchdb.user:%s", username),
		Token: token,
	})
}

// GetCurrentUser 获取当前用户信息
// GET /-/api/user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"createdAt":  user.CreatedAt,
		"lastLogin":  user.LastLogin,
	})
}

// Logout 用户登出
// DELETE /-/api/session
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 是无状态的，登出只需客户端删除 token
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListUsers 列出所有用户 (管理员)
// GET /-/api/admin/users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users := h.userStore.List()
	
	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		result = append(result, gin.H{
			"username":  u.Username,
			"email":     u.Email,
			"role":      u.Role,
			"createdAt": u.CreatedAt,
			"lastLogin": u.LastLogin,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

// CreateUser 创建用户 (管理员)
// POST /-/api/admin/users
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Role == "" {
		req.Role = "developer"
	}

	user := &auth.User{
		Username: req.Name,
		Password: req.Password,
		Email:    req.Email,
		Role:     req.Role,
	}

	if err := h.userStore.Create(user); err != nil {
		if err == auth.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	logger.Infof("Admin created user: %s", req.Name)

	c.JSON(http.StatusCreated, gin.H{
		"ok":       true,
		"username": user.Username,
		"role":     user.Role,
	})
}

// DeleteUser 删除用户 (管理员)
// DELETE /-/api/admin/users/:username
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	
	if username == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
		return
	}

	if err := h.userStore.Delete(username); err != nil {
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	logger.Infof("Admin deleted user: %s", username)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func extractUsernameFromPath(pathUsername string) string {
	// 路径格式: org.couchdb.user:username
	if strings.HasPrefix(pathUsername, "org.couchdb.user:") {
		return strings.TrimPrefix(pathUsername, "org.couchdb.user:")
	}
	return pathUsername
}
