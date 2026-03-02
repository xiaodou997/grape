// Package webhook 实现事件 Webhook 通知
// 支持：异步队列投递、并发控制、HMAC-SHA256 签名、3 次自动重试
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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

type deliveryTask struct {
	hook db.Webhook
	data []byte
}

// Dispatcher Webhook 分发器
type Dispatcher struct {
	client    *http.Client
	taskQueue chan deliveryTask
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewDispatcher 创建并启动 Webhook 分发器
func NewDispatcher() *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Dispatcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		taskQueue: make(chan deliveryTask, 100),
		ctx:       ctx,
		cancel:    cancel,
	}

	for i := 0; i < 3; i++ {
		go d.worker(i)
	}

	return d
}

// Stop 停止分发器
func (d *Dispatcher) Stop() {
	d.cancel()
	close(d.taskQueue)
	d.wg.Wait()
}

// Dispatch 发送事件
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

	go func() {
		var hooks []db.Webhook
		if err := db.DB.Where("enabled = ?", true).Find(&hooks).Error; err != nil {
			logger.Errorf("Webhook: failed to query webhooks: %v", err)
			return
		}

		for _, hook := range hooks {
			if hook.Events != "" && !containsEvent(hook.Events, string(eventType)) {
				continue
			}

			d.enqueue(hook, data)
		}
	}()
}

// Test 发送测试消息
func (d *Dispatcher) Test(hook db.Webhook) {
	event := Event{
		Event:     "webhook:test",
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"message": "This is a test event from Grape Registry. 🍇",
			"test":    true,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	logger.Infof("Webhook: Enqueuing test task for %s", hook.Name)
	d.enqueue(hook, data)
}

func (d *Dispatcher) enqueue(hook db.Webhook, data []byte) {
	select {
	case d.taskQueue <- deliveryTask{hook: hook, data: data}:
		d.wg.Add(1)
	default:
		logger.Errorf("Webhook: task queue full, dropping event for %s", hook.URL)
	}
}

func (d *Dispatcher) worker(id int) {
	logger.Infof("Webhook: Worker %d started", id)
	for task := range d.taskQueue {
		d.deliver(task.hook, task.data)
		d.wg.Done()
	}
}

func (d *Dispatcher) deliver(hook db.Webhook, data []byte) {
	const maxRetries = 3
	const retryDelay = 5 * time.Second

	finalData := d.adaptPayload(hook.URL, hook.Secret, data)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Infof("Webhook: Sending to %s (attempt %d/%d)", hook.Name, attempt, maxRetries)
		err := d.send(hook.URL, finalData)
		if err == nil {
			logger.Infof("Webhook: Successfully delivered to %s", hook.Name)
			db.DB.Model(&hook).Update("last_delivery_at", time.Now())
			return
		}
		
		logger.Warnf("Webhook: Delivery failed to %s: %v", hook.Name, err)
		
		if attempt < maxRetries {
			select {
			case <-time.After(retryDelay):
				continue
			case <-d.ctx.Done():
				return
			}
		}
	}
}

func (d *Dispatcher) adaptPayload(url, secret string, data []byte) []byte {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return data
	}

	msgText := fmt.Sprintf("🍇 [Grape Registry]\nEvent: %s\nTime: %s", 
		event.Event, event.Timestamp.Local().Format("2006-01-02 15:04:05"))
	
	// 增强类型转换安全性
	if p, ok := event.Payload.(map[string]interface{}); ok {
		if pkg, exists := p["package"]; exists { msgText += fmt.Sprintf("\nPackage: %v", pkg) }
		if user, exists := p["publisher"]; exists { msgText += fmt.Sprintf("\nBy: %v", user) }
		if msg, exists := p["message"]; exists { msgText += fmt.Sprintf("\nDetail: %v", msg) }
	}

	// 飞书 (Feishu/Lark)
	if strings.Contains(url, "feishu.cn") || strings.Contains(url, "larksuite.com") {
		content := map[string]interface{}{"text": msgText}
		feishuMsg := map[string]interface{}{
			"msg_type": "text",
			"content":  content,
		}
		if secret != "" {
			ts := time.Now().Unix()
			feishuMsg["timestamp"] = fmt.Sprintf("%d", ts)
			feishuMsg["sign"] = computeFeishuSign(secret, ts)
		}
		adapted, _ := json.Marshal(feishuMsg)
		return adapted
	}

	// 钉钉 (DingTalk)
	if strings.Contains(url, "dingtalk.com") {
		dingMsg := map[string]interface{}{
			"msgtype": "text",
			"text":    map[string]interface{}{"content": msgText},
		}
		adapted, _ := json.Marshal(dingMsg)
		return adapted
	}

	return data
}

func (d *Dispatcher) send(url string, data []byte) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 记录业务回执（非常重要：飞书会在这里返回错误码）
	respStr := string(body)
	if strings.Contains(respStr, "\"code\"") || strings.Contains(respStr, "\"errcode\"") {
		logger.Infof("Webhook: Response from server: %s", respStr)
		// 如果飞书返回 code != 0，应视为失败
		if strings.Contains(respStr, "\"code\":0") || strings.Contains(respStr, "\"errcode\":0") {
			return nil
		}
		if !strings.Contains(respStr, ":0") {
			return fmt.Errorf("business error: %s", respStr)
		}
	}

	return nil
}

func computeFeishuSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write([]byte(""))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func containsEvent(events, target string) bool {
	for _, e := range strings.Split(events, ",") {
		if strings.TrimSpace(e) == target {
			return true
		}
	}
	return false
}
