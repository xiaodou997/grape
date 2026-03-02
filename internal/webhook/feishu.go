package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// buildFeishuCard 构造飞书互动卡片载荷
func buildFeishuCard(event Event, secret string) ([]byte, error) {
	title := "🍇 Grape Registry 通知"
	template := "blue"

	switch event.Event {
	case EventPackagePublished:
		title = "📦 新包发布通知"
		template = "green"
	case EventPackageUnpublished:
		title = "🗑️ 包已删除通知"
		template = "red"
	case "webhook:test":
		title = "🧪 Webhook 测试连接"
		template = "purple"
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("**事件类型:** %s \n", event.Event))
	md.WriteString(fmt.Sprintf("**触发时间:** %s \n", event.Timestamp.Local().Format("2006-01-02 15:04:05")))

	// 解析 Payload 详情
	if p, ok := event.Payload.(map[string]interface{}); ok {
		if pkg, exists := p["package"]; exists {
			md.WriteString(fmt.Sprintf("**包名称:** %v \n", pkg))
		}
		if v, exists := p["versions"]; exists {
			if vMap, ok := v.(map[string]interface{}); ok {
				if latest, exists := vMap["latest"]; exists {
					md.WriteString(fmt.Sprintf("**最新版本:** v%v \n", latest))
				}
			}
		}
		if user, exists := p["publisher"]; exists {
			md.WriteString(fmt.Sprintf("**发布者:** %v \n", user))
		}
		if msg, exists := p["message"]; exists {
			md.WriteString(fmt.Sprintf("**详情描述:** %v \n", msg))
		}
	}

	card := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title":    map[string]string{"tag": "plain_text", "content": title},
				"template": template,
			},
			"elements": []interface{}{
				map[string]interface{}{
					"tag": "div",
					"text": map[string]string{
						"tag":     "lark_md",
						"content": md.String(),
					},
				},
				map[string]interface{}{"tag": "hr"},
				map[string]interface{}{
					"tag": "note",
					"elements": []interface{}{
						map[string]string{"tag": "plain_text", "content": "来自 Grape Registry 自动化流水线"},
					},
				},
			},
		},
	}

	// 飞书专用签名逻辑
	if secret != "" {
		ts := time.Now().Unix()
		card["timestamp"] = fmt.Sprintf("%d", ts)
		card["sign"] = computeFeishuSign(secret, ts)
	}

	return json.Marshal(card)
}

// computeFeishuSign 飞书签名算法
func computeFeishuSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write([]byte(""))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
