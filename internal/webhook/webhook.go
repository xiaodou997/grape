// Package webhook 实现事件 Webhook 通知
// 支持：异步投递、HMAC-SHA256 签名、3 次自动重试
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
)

// EventType Webhook 事件类型
type EventType string

const (
	EventPackagePublished   EventType = "package:published"
	EventPackageUnpublished EventType = "package:unpublished"
	EventUserCreated        EventType = "user:created"
	EventUserDeleted        EventType = "user:deleted"
)

// Event Webhook 事件载荷
type Event struct {
	Event     EventType   `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Dispatcher Webhook 分发器
type Dispatcher struct {
	client *http.Client
}

// NewDispatcher 创建 Webhook 分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Dispatch 异步向所有匹配的 Webhook 投递事件
func (d *Dispatcher) Dispatch(eventType EventType, payload interface{}) {
	event := Event{
		Event:     eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Errorf("Webhook: failed to marshal event %s: %v", eventType, err)
		return
	}

	// 查询启用的 Webhook 配置
	var hooks []db.Webhook
	if err := db.DB.Where("enabled = ?", true).Find(&hooks).Error; err != nil {
		logger.Errorf("Webhook: failed to query webhooks: %v", err)
		return
	}

	for _, hook := range hooks {
		// 过滤事件类型（空 Events 表示接收所有事件）
		if hook.Events != "" && !containsEvent(hook.Events, string(eventType)) {
			continue
		}

		go d.deliver(hook, data)
	}
}

// deliver 向单个 Webhook 端点投递，最多重试 3 次
func (d *Dispatcher) deliver(hook db.Webhook, data []byte) {
	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := d.send(hook.URL, hook.Secret, data)
		if err == nil {
			logger.Debugf("Webhook delivered to %s (attempt %d)", hook.URL, attempt)
			// 更新最后投递时间
			now := time.Now()
			db.DB.Model(&hook).Update("last_delivery_at", now)
			return
		}
		logger.Warnf("Webhook delivery failed to %s (attempt %d/%d): %v", hook.URL, attempt, maxRetries, err)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}
	logger.Errorf("Webhook: all %d attempts failed for %s", maxRetries, hook.URL)
}

// send 发送一次 HTTP POST 请求
func (d *Dispatcher) send(url, secret string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Grape-Webhook/1.0")
	req.Header.Set("X-Grape-Event", "webhook")

	// HMAC-SHA256 签名
	if secret != "" {
		sig := computeHMAC(secret, data)
		req.Header.Set("X-Grape-Signature", "sha256="+sig)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

// computeHMAC 计算 HMAC-SHA256 签名
func computeHMAC(secret string, data []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// containsEvent 检查事件类型字符串（逗号分隔）是否包含目标事件
func containsEvent(events, target string) bool {
	start := 0
	for i := 0; i <= len(events); i++ {
		if i == len(events) || events[i] == ',' {
			if events[start:i] == target {
				return true
			}
			start = i + 1
		}
	}
	return false
}
