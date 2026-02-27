package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/webhook"
)

// WebhookHandler 管理 Webhook 配置
type WebhookHandler struct {
	dispatcher *webhook.Dispatcher
}

func NewWebhookHandler(dispatcher *webhook.Dispatcher) *WebhookHandler {
	return &WebhookHandler{dispatcher: dispatcher}
}

// ListWebhooks GET /-/api/admin/webhooks
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	var hooks []db.Webhook
	if err := db.DB.Find(&hooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list webhooks"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"webhooks": hooks})
}

// CreateWebhook POST /-/api/admin/webhooks
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		URL     string `json:"url" binding:"required"`
		Secret  string `json:"secret"`
		Events  string `json:"events"` // 逗号分隔，空=全部
		Enabled *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	hook := &db.Webhook{
		Name:    req.Name,
		URL:     req.URL,
		Secret:  req.Secret,
		Events:  req.Events,
		Enabled: enabled,
	}

	if err := db.DB.Create(hook).Error; err != nil {
		logger.Errorf("Failed to create webhook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}

	logger.Infof("Webhook created: %s -> %s", hook.Name, hook.URL)
	c.JSON(http.StatusCreated, gin.H{"ok": true, "id": hook.ID})
}

// UpdateWebhook PUT /-/api/admin/webhooks/:id
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var hook db.Webhook
	if err := db.DB.First(&hook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	var req struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		Secret  string `json:"secret"`
		Events  string `json:"events"`
		Enabled *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Name != "" {
		hook.Name = req.Name
	}
	if req.URL != "" {
		hook.URL = req.URL
	}
	if req.Secret != "" {
		hook.Secret = req.Secret
	}
	if req.Events != "" {
		hook.Events = req.Events
	}
	if req.Enabled != nil {
		hook.Enabled = *req.Enabled
	}

	if err := db.DB.Save(&hook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteWebhook DELETE /-/api/admin/webhooks/:id
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result := db.DB.Delete(&db.Webhook{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete webhook"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// TestWebhook POST /-/api/admin/webhooks/:id/test
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var hook db.Webhook
	if err := db.DB.First(&hook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	// 异步发送测试事件
	h.dispatcher.Dispatch(webhook.EventType("webhook:test"), gin.H{
		"message": "This is a test event from Grape registry",
		"id":      hook.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "test event dispatched"})
}
