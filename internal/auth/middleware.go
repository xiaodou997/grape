package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/logger"
)

type contextKey string

const (
	UserKey contextKey = "user"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtService *JWTService, userStore UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 尝试从 query 获取 token
			authHeader = c.Query("authToken")
			if authHeader != "" {
				authHeader = "Bearer " + authHeader
			}
		}

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
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Debugf("Token validation failed: %v", err)
			c.Next()
			return
		}

		// 获取用户信息
		user, err := userStore.Get(claims.Username)
		if err != nil {
			c.Next()
			return
		}

		// 将用户信息存入 context
		c.Set(string(UserKey), user)
		c.Next()
	}
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
