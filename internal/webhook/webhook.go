// Package webhook 实现事件 Webhook 通知
// 支持：异步队列投递、多平台自适应、自动重试
package webhook

import (
	"bytes"
	"context"
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

type EventType string

const (
	EventPackagePublished   EventType = "package:published"
	EventPackageUnpublished EventType = "package:unpublished"
	EventUserCreated        EventType = "user:created"
	EventUserDeleted        EventType = "user:deleted"
)

type Event struct {
	Event     EventType   `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type deliveryTask struct {
	hook db.Webhook
	data []byte
}

type Dispatcher struct {
	client    *http.Client
	taskQueue chan deliveryTask
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewDispatcher() *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Dispatcher{
		client:    &http.Client{Timeout: 15 * time.Second},
		taskQueue: make(chan deliveryTask, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
	for i := 0; i < 3; i++ {
		go d.worker(i)
	}
	return d
}

func (d *Dispatcher) Stop() {
	d.cancel()
	close(d.taskQueue)
	d.wg.Wait()
}

func (d *Dispatcher) Dispatch(eventType EventType, payload interface{}) {
	event := Event{
		Event:     eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	go func() {
		var hooks []db.Webhook
		if err := db.DB.Where("enabled = ?", true).Find(&hooks).Error; err != nil {
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

func (d *Dispatcher) Test(hook db.Webhook) {
	event := Event{
		Event:     "webhook:test",
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"message": "This is a test event from Grape Registry. 🍇",
			"test":    true,
		},
	}
	data, _ := json.Marshal(event)
	logger.Infof("Webhook: Enqueuing test task for %s", hook.Name)
	d.enqueue(hook, data)
}

func (d *Dispatcher) enqueue(hook db.Webhook, data []byte) {
	select {
	case d.taskQueue <- deliveryTask{hook: hook, data: data}:
		d.wg.Add(1)
	default:
		logger.Errorf("Webhook: queue full for %s", hook.URL)
	}
}

func (d *Dispatcher) worker(id int) {
	for task := range d.taskQueue {
		d.deliver(task.hook, task.data)
		d.wg.Done()
	}
}

func (d *Dispatcher) deliver(hook db.Webhook, data []byte) {
	const maxRetries = 3
	const retryDelay = 5 * time.Second

	finalData := d.adaptToPlatform(hook, data)

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
			time.Sleep(retryDelay)
		}
	}
}

func (d *Dispatcher) adaptToPlatform(hook db.Webhook, data []byte) []byte {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return data
	}

	url := strings.ToLower(hook.URL)

	if strings.Contains(url, "feishu.cn") || strings.Contains(url, "larksuite.com") {
		adapted, err := buildFeishuCard(event, hook.Secret)
		if err == nil {
			return adapted
		}
	}

	if strings.Contains(url, "dingtalk.com") {
		msg := map[string]interface{}{
			"msgtype": "text",
			"text":    map[string]interface{}{"content": fmt.Sprintf("🍇 [Grape] %s\nTime: %s", event.Event, time.Now().Format("15:04:05"))},
		}
		adapted, _ := json.Marshal(msg)
		return adapted
	}

	return data
}

func (d *Dispatcher) send(url string, data []byte) error {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Grape-Webhook/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if strings.Contains(url, "feishu") || strings.Contains(url, "dingtalk") {
		logger.Infof("Webhook: Response from server: %s", string(body))
	}
	return nil
}

func containsEvent(events, target string) bool {
	for _, e := range strings.Split(events, ",") {
		if strings.TrimSpace(e) == target {
			return true
		}
	}
	return false
}
