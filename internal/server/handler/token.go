package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

// TokenHandler Token 管理 Handler
type TokenHandler struct{}

// NewTokenHandler 创建 TokenHandler
func NewTokenHandler() *TokenHandler {
	return &TokenHandler{}
}

// CreateTokenRequest 创建 token 请求
type CreateTokenRequest struct {
	Name     string `json:"name"`              // token 名称，如 "github-ci"
	Readonly bool   `json:"readonly"`          // 是否只读
	Days     int    `json:"days,omitempty"`    // 有效期天数，0 表示永不过期
}

// TokenResponse token 响应
type TokenResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Readonly  bool       `json:"readonly"`
	Token     string     `json:"token,omitempty"` // 仅创建时返回
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	LastUsed  *time.Time `json:"lastUsed,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// CreateToken 创建新 token
// POST /-/npm/v1/tokens
func (h *TokenHandler) CreateToken(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		req.Name = "token"
	}

	// 生成随机 token (32 字节 = 64 字符 hex)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// 计算哈希存储
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// 计算过期时间
	var expiresAt *time.Time
	if req.Days > 0 {
		exp := time.Now().AddDate(0, 0, req.Days)
		expiresAt = &exp
	}

	// 创建 token 记录
	tokenRecord := &db.Token{
		UserID:    user.ID,
		Name:      req.Name,
		TokenHash: tokenHash,
		Readonly:  req.Readonly,
		ExpiresAt: expiresAt,
	}

	if err := db.DB.Create(tokenRecord).Error; err != nil {
		logger.Errorf("Failed to save token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

	// 记录审计日志
	db.DB.Create(&db.AuditLog{
		Action:   "token_create",
		Username: user.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Created token '%s' (readonly: %v)", req.Name, req.Readonly),
	})

	logger.Infof("User %s created token '%s'", user.Username, req.Name)

	c.JSON(http.StatusCreated, TokenResponse{
		ID:        tokenRecord.ID,
		Name:      tokenRecord.Name,
		Readonly:  tokenRecord.Readonly,
		Token:     token, // 只在创建时返回一次
		ExpiresAt: tokenRecord.ExpiresAt,
		CreatedAt: tokenRecord.CreatedAt,
	})
}

// ListTokens 列出当前用户的所有 token
// GET /-/npm/v1/tokens
func (h *TokenHandler) ListTokens(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var tokens []db.Token
	if err := db.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		logger.Errorf("Failed to list tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}

	responses := make([]TokenResponse, len(tokens))
	for i, t := range tokens {
		responses[i] = TokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			Readonly:  t.Readonly,
			ExpiresAt: t.ExpiresAt,
			LastUsed:  t.LastUsed,
			CreatedAt: t.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"objects": responses,
	})
}

// DeleteToken 撤销 token
// DELETE /-/npm/v1/tokens/token/:id
func (h *TokenHandler) DeleteToken(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	tokenID := c.Param("id")
	if tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token id required"})
		return
	}

	// 查找 token，确保属于当前用户
	var token db.Token
	if err := db.DB.Where("id = ? AND user_id = ?", tokenID, user.ID).First(&token).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	if err := db.DB.Delete(&token).Error; err != nil {
		logger.Errorf("Failed to delete token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete token"})
		return
	}

	// 记录审计日志
	db.DB.Create(&db.AuditLog{
		Action:   "token_delete",
		Username: user.Username,
		IP:       c.ClientIP(),
		Detail:   fmt.Sprintf("Revoked token '%s' (id: %d)", token.Name, token.ID),
	})

	logger.Infof("User %s revoked token '%s'", user.Username, token.Name)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
