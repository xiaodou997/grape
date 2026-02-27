package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

// loginLimiter 登录限流器
type loginLimiter struct {
	attempts map[string]*attemptInfo
	mu       sync.RWMutex
	done     chan struct{} // 用于优雅关闭
	stopped  bool          // 防止重复关闭
}

type attemptInfo struct {
	count     int
	firstSeen time.Time
}

var (
	limiter     *loginLimiter
	limiterOnce sync.Once
)

func getLoginLimiter() *loginLimiter {
	limiterOnce.Do(func() {
		limiter = &loginLimiter{
			attempts: make(map[string]*attemptInfo),
			done:     make(chan struct{}),
		}
		// 定期清理过期记录
		go limiter.runCleanup()
	})
	return limiter
}

// runCleanup 定期清理过期记录，支持优雅关闭
func (l *loginLimiter) runCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.done:
			return
		}
	}
}

// Stop 停止清理 goroutine
func (l *loginLimiter) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.stopped {
		close(l.done)
		l.stopped = true
	}
}

func (l *loginLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	now := time.Now()
	for ip, info := range l.attempts {
		if now.Sub(info.firstSeen) > time.Minute {
			delete(l.attempts, ip)
		}
	}
}

func (l *loginLimiter) checkLimit(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	info, exists := l.attempts[ip]
	
	if !exists || now.Sub(info.firstSeen) > time.Minute {
		l.attempts[ip] = &attemptInfo{
			count:     1,
			firstSeen: now,
		}
		return true
	}

	if info.count >= 10 {
		return false
	}

	info.count++
	return true
}

func (l *loginLimiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// StopLoginLimiter 停止登录限流器（服务关闭时调用）
func StopLoginLimiter() {
	if limiter != nil {
		limiter.Stop()
	}
}

type AuthHandler struct {
	userStore        auth.UserStore
	jwtService       *auth.JWTService
	allowRegistration bool
}

func NewAuthHandler(userStore auth.UserStore, jwtService *auth.JWTService, allowRegistration bool) *AuthHandler {
	return &AuthHandler{
		userStore:         userStore,
		jwtService:        jwtService,
		allowRegistration: allowRegistration,
	}
}

// SetAllowRegistration 动态更新自助注册开关
func (h *AuthHandler) SetAllowRegistration(allow bool) {
	h.allowRegistration = allow
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
	// 检查登录限流
	clientIP := c.ClientIP()
	if !getLoginLimiter().checkLimit(clientIP) {
		logger.Warnf("Login rate limited for IP: %s", clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many login attempts, please try again later",
		})
		return
	}

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
		// 用户不存在，检查是否允许自助注册
		if !h.allowRegistration {
			logger.Warnf("Registration attempted for non-existent user: %s (registration disabled)", username)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "registration is disabled, contact your administrator",
			})
			return
		}
		
		// 允许自助注册，创建新用户
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
			return
		}
		
		// 验证密码强度
		if err := auth.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// 登录成功，重置限流计数
	getLoginLimiter().reset(clientIP)

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
	db.RecordAudit("login", username, c.ClientIP(), "登录成功")

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

	// 验证密码强度
	if err := auth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色
	if req.Role == "" {
		req.Role = "developer"
	}
	if err := auth.ValidateRole(req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	adminUser := auth.GetCurrentUser(c)
	adminName := ""
	if adminUser != nil {
		adminName = adminUser.Username
	}
	db.RecordAudit("user_create", adminName, c.ClientIP(), "创建用户: "+req.Name)

	c.JSON(http.StatusCreated, gin.H{
		"ok":       true,
		"username": user.Username,
		"role":     user.Role,
	})
}

// UpdateUser 更新用户信息 (管理员)
// PUT /-/api/admin/users/:username
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	username := c.Param("username")

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"` // 留空则不修改密码
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 如果要更新密码，验证密码强度
	if req.Password != "" {
		if err := auth.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// 如果要更新角色，验证角色合法性
	if req.Role != "" {
		if err := auth.ValidateRole(req.Role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	existingUser, err := h.userStore.Get(username)
	if err != nil {
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// 只更新非空字段
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Role != "" {
		existingUser.Role = req.Role
	}
	if req.Password != "" {
		existingUser.Password = req.Password
	}

	if err := h.userStore.Update(existingUser); err != nil {
		logger.Errorf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	logger.Infof("Admin updated user: %s", username)
	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"username": username,
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
	delAdminUser := auth.GetCurrentUser(c)
	delAdminName := ""
	if delAdminUser != nil {
		delAdminName = delAdminUser.Username
	}
	db.RecordAudit("user_delete", delAdminName, c.ClientIP(), "删除用户: "+username)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func extractUsernameFromPath(pathUsername string) string {
	// 路径格式: org.couchdb.user:username
	if strings.HasPrefix(pathUsername, "org.couchdb.user:") {
		return strings.TrimPrefix(pathUsername, "org.couchdb.user:")
	}
	return pathUsername
}
