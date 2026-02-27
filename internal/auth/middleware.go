package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

type contextKey string

const (
	UserKey       contextKey = "user"
	TokenKey      contextKey = "token"
	IsTokenAuth   contextKey = "isTokenAuth"
)

// TokenValidator Token 验证接口
type TokenValidator interface {
	ValidateTokenByHash(tokenString string) (*User, *TokenInfo, error)
}

// TokenInfo Token 信息（简化版，避免循环依赖）
type TokenInfo struct {
	ID       uint
	Name     string
	Readonly bool
}

// ValidateTokenByHash 通过 token 哈希验证 token，返回关联用户
func ValidateTokenByHash(tokenString string) (*User, *TokenInfo, error) {
	// 计算哈希
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	var token db.Token
	if err := db.DB.Where("token_hash = ?", tokenHash).Preload("User").First(&token).Error; err != nil {
		return nil, nil, err
	}

	// 检查是否过期
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token expired")
	}

	// 更新最后使用时间
	now := time.Now()
	db.DB.Model(&token).Update("last_used", now)

	// 转换为 auth.User
	user := &User{
		ID:        token.User.ID,
		Username:  token.User.Username,
		Email:     token.User.Email,
		Role:      token.User.Role,
		CreatedAt: token.User.CreatedAt,
		LastLogin: token.User.LastLogin,
	}

	tokenInfo := &TokenInfo{
		ID:       token.ID,
		Name:     token.Name,
		Readonly: token.Readonly,
	}

	return user, tokenInfo, nil
}

// AuthMiddleware JWT 和 Token 认证中间件
// 优先检查 JWT，再检查持久化 Token
func AuthMiddleware(jwtService *JWTService, userStore UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// 不再支持 query 参数传递 token，避免 token 泄露到日志

		if authHeader == "" {
			c.Next()
			return
		}

		// 解析 Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// 1. 先尝试 JWT 验证
		claims, err := jwtService.ValidateToken(tokenString)
		if err == nil {
			// JWT 有效
			user, err := userStore.Get(claims.Username)
			if err == nil {
				c.Set(string(UserKey), user)
				c.Set(string(IsTokenAuth), false)
			}
			c.Next()
			return
		}

		// 2. JWT 无效，尝试 Token 验证
		user, tokenInfo, err := ValidateTokenByHash(tokenString)
		if err != nil {
			logger.Debugf("Token validation failed: %v", err)
			c.Next()
			return
		}

		// Token 有效
		c.Set(string(UserKey), user)
		c.Set(string(TokenKey), tokenInfo)
		c.Set(string(IsTokenAuth), true)
		c.Next()
	}
}

// GetTokenInfo 从 context 获取 token 信息
func GetTokenInfo(c *gin.Context) *TokenInfo {
	info, exists := c.Get(string(TokenKey))
	if !exists {
		return nil
	}
	t, ok := info.(*TokenInfo)
	if !ok {
		return nil
	}
	return t
}

// IsTokenAuth check if current auth is via persistent token
func GetIsTokenAuth(c *gin.Context) bool {
	isToken, exists := c.Get(string(IsTokenAuth))
	if !exists {
		return false
	}
	return isToken.(bool)
}

// RequireAuth 要求认证的中间件
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(string(UserKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		// 确保用户类型正确
		_, ok := user.(*User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user context",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(string(UserKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		u, ok := user.(*User)
		if !ok || u.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 从 context 获取当前用户
func GetCurrentUser(c *gin.Context) *User {
	user, exists := c.Get(string(UserKey))
	if !exists {
		return nil
	}
	
	u, ok := user.(*User)
	if !ok {
		return nil
	}
	return u
}

// IsAuthenticated 检查是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(string(UserKey))
	return exists
}
