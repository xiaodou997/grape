# Grape Webhook 使用指南

本文档介绍如何配置和使用 Grape 的 Webhook 功能。

## 目录

- [Webhook 概述](#webhook-概述)
- [事件类型](#事件类型)
- [配置 Webhook](#配置-webhook)
- [接收 Webhook](#接收-webhook)
- [签名验证](#签名验证)
- [重试机制](#重试机制)
- [使用示例](#使用示例)
- [故障排查](#故障排查)

---

## Webhook 概述

Grape Webhook 功能允许您在特定事件发生时接收 HTTP POST 通知。您可以配置 Webhook 端点来：

- 接收包发布/删除通知
- 与 CI/CD 系统集成
- 发送 Slack/钉钉/企业微信通知
- 触发自定义工作流

### 工作流程

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Grape     │      │   Webhook   │      │   Your      │
│   Event     │─────>│ Dispatcher  │─────>│   Server    │
│  Occurs     │      │             │      │   Endpoint  │
└─────────────┘      └─────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │   Retry     │
                     │   Logic     │
                     └─────────────┘
```

---

## 事件类型

Grape 支持以下事件类型：

| 事件 | 说明 | 触发时机 |
|------|------|----------|
| `package:published` | 包发布 | 当新包或新版本发布时 |
| `package:unpublished` | 包删除 | 当包或版本被删除时 |
| `user:created` | 用户创建 | 当新用户被创建时 |
| `user:deleted` | 用户删除 | 当用户被删除时 |

### 事件载荷格式

#### package:published

```json
{
  "event": "package:published",
  "timestamp": "2024-01-02T12:00:00Z",
  "payload": {
    "package": "@grape/cli",
    "publisher": "admin",
    "versions": {
      "latest": "1.2.3"
    }
  }
}
```

#### package:unpublished

```json
{
  "event": "package:unpublished",
  "timestamp": "2024-01-02T12:00:00Z",
  "payload": {
    "package": "@grape/cli",
    "operator": "admin"
  }
}
```

#### user:created

```json
{
  "event": "user:created",
  "timestamp": "2024-01-02T12:00:00Z",
  "payload": {
    "username": "newuser",
    "role": "developer"
  }
}
```

#### user:deleted

```json
{
  "event": "user:deleted",
  "timestamp": "2024-01-02T12:00:00Z",
  "payload": {
    "username": "olduser"
  }
}
```

---

## 配置 Webhook

### 方式一：通过 Web UI

1. 登录 Grape Web 界面
2. 进入「管理后台」>「Webhook 管理」
3. 点击「创建 Webhook」
4. 填写配置：
   - **名称**: Webhook 标识名称
   - **URL**: 接收通知的端点 URL
   - **Secret**: HMAC 签名密钥（可选，推荐）
   - **事件**: 订阅的事件类型（逗号分隔，留空表示所有事件）
   - **启用**: 是否启用此 Webhook
5. 点击「保存」

### 方式二：通过 API

```bash
# 创建 Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack Notification",
    "url": "https://hooks.slack.com/services/xxx",
    "secret": "my-secret-key",
    "events": "package:published,package:unpublished",
    "enabled": true
  }'
```

### 配置项说明

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | 是 | Webhook 名称（标识符） |
| `url` | string | 是 | 接收端点 URL（必须可公网访问） |
| `secret` | string | 否 | HMAC 签名密钥，用于验证请求来源 |
| `events` | string | 否 | 逗号分隔的事件类型，留空表示订阅所有 |
| `enabled` | bool | 否 | 是否启用，默认 `true` |

### 管理 Webhook

```bash
# 获取 Webhook 列表
curl http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>"

# 更新 Webhook
curl -X PUT http://localhost:4873/-/api/admin/webhooks/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "enabled": false
  }'

# 删除 Webhook
curl -X DELETE http://localhost:4873/-/api/admin/webhooks/1 \
  -H "Authorization: Bearer <admin_token>"

# 测试 Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks/1/test \
  -H "Authorization: Bearer <admin_token>"
```

---

## 接收 Webhook

### 服务端点要求

1. **HTTP 方法**: 必须支持 `POST`
2. **Content-Type**: 接收 `application/json`
3. **响应状态码**: 返回 `2xx` 表示成功
4. **超时时间**: 建议在 10 秒内响应

### 示例端点（Node.js/Express）

```javascript
const express = require('express');
const crypto = require('crypto');
const app = express();

app.use(express.json());

const WEBHOOK_SECRET = 'my-secret-key';

// 验证签名
function verifySignature(req, res, buf) {
  const signature = req.headers['x-grape-signature'];
  const hmac = crypto.createHmac('sha256', WEBHOOK_SECRET);
  const digest = 'sha256=' + hmac.update(buf).digest('hex');
  
  if (signature !== digest) {
    throw new Error('Invalid signature');
  }
}

app.post('/webhook', (req, res) => {
  try {
    // 验证签名
    verifySignature(req, res, req.rawBody);
    
    const { event, timestamp, payload } = req.body;
    
    console.log(`Received event: ${event} at ${timestamp}`);
    console.log('Payload:', payload);
    
    // 处理事件
    switch (event) {
      case 'package:published':
        handlePackagePublished(payload);
        break;
      case 'package:unpublished':
        handlePackageUnpublished(payload);
        break;
      // ... 其他事件
    }
    
    res.status(200).json({ ok: true });
  } catch (error) {
    console.error('Webhook error:', error);
    res.status(400).json({ error: error.message });
  }
});

function handlePackagePublished(payload) {
  console.log(`Package ${payload.package} published by ${payload.publisher}`);
  // 发送通知、触发 CI/CD 等
}

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});
```

### 示例端点（Go/Gin）

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
)

const webhookSecret = "my-secret-key"

type WebhookEvent struct {
    Event     string      `json:"event"`
    Timestamp string      `json:"timestamp"`
    Payload   interface{} `json:"payload"`
}

func main() {
    r := gin.Default()
    
    r.POST("/webhook", func(c *gin.Context) {
        body, _ := c.GetRawData()
        
        // 验证签名
        signature := c.GetHeader("X-Grape-Signature")
        if !verifySignature(body, signature) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
            return
        }
        
        var event WebhookEvent
        if err := c.ShouldBindJSON(&event); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
            return
        }
        
        fmt.Printf("Received event: %s at %s\n", event.Event, event.Timestamp)
        fmt.Printf("Payload: %+v\n", event.Payload)
        
        // 处理事件
        handleEvent(event)
        
        c.JSON(http.StatusOK, gin.H{"ok": true})
    })
    
    r.Run(":3000")
}

func verifySignature(body []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(webhookSecret))
    mac.Write(body)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte("sha256="+expectedMAC), []byte(signature))
}

func handleEvent(event WebhookEvent) {
    // 处理不同类型的事件
}
```

### 示例端点（Python/Flask）

```python
from flask import Flask, request, abort
import hmac
import hashlib
import json

app = Flask(__name__)
WEBHOOK_SECRET = 'my-secret-key'

def verify_signature(payload, signature):
    mac = hmac.new(
        WEBHOOK_SECRET.encode(),
        payload,
        hashlib.sha256
    )
    expected = 'sha256=' + mac.hexdigest()
    return hmac.compare_digest(expected, signature)

@app.route('/webhook', methods=['POST'])
def webhook():
    signature = request.headers.get('X-Grape-Signature')
    payload = request.get_data()
    
    if not verify_signature(payload, signature):
        abort(400)
    
    data = json.loads(payload)
    event = data['event']
    timestamp = data['timestamp']
    payload_data = data['payload']
    
    print(f"Received event: {event} at {timestamp}")
    print(f"Payload: {payload_data}")
    
    # 处理事件
    handle_event(event, payload_data)
    
    return json.dumps({'ok': True})

def handle_event(event, payload):
    if event == 'package:published':
        print(f"Package {payload['package']} published by {payload['publisher']}")
    # ... 其他事件处理

if __name__ == '__main__':
    app.run(port=3000)
```

---

## 签名验证

### 签名机制

Grape 使用 HMAC-SHA256 对请求体进行签名，确保请求来源可信。

**签名计算：**

```
signature = HMAC-SHA256(secret, request_body)
```

**HTTP 请求头：**

```http
X-Grape-Signature: sha256=<hex_encoded_signature>
```

### 验证步骤

1. 读取请求体（原始 JSON 数据）
2. 使用配置的 `secret` 计算 HMAC-SHA256
3. 将计算结果与 `X-Grape-Signature` 头比较
4. 如果不匹配，拒绝请求

### 验证代码示例

```python
# Python
import hmac
import hashlib

def verify_signature(payload_bytes, secret, signature_header):
    # 计算期望的签名
    mac = hmac.new(secret.encode(), payload_bytes, hashlib.sha256)
    expected = 'sha256=' + mac.hexdigest()
    
    # 比较签名
    return hmac.compare_digest(expected, signature_header)
```

```javascript
// Node.js
const crypto = require('crypto');

function verifySignature(payload, secret, signature) {
  const mac = crypto.createHmac('sha256', secret);
  mac.update(payload);
  const digest = 'sha256=' + mac.digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(digest),
    Buffer.from(signature)
  );
}
```

```go
// Go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verifySignature(body []byte, secret, signature string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

---

## 重试机制

### 重试策略

如果 Webhook 端点返回非 `2xx` 状态码或超时，Grape 会自动重试：

| 参数 | 值 | 说明 |
|------|-----|------|
| 最大重试次数 | 3 次 | 超过后放弃投递 |
| 重试间隔 | 5 秒 | 每次重试间隔 |
| 超时时间 | 10 秒 | 单次请求超时 |

### 重试流程

```
第一次投递 (失败)
    ↓ (等待 5 秒)
第二次投递 (失败)
    ↓ (等待 5 秒)
第三次投递 (成功/失败)
    ↓
记录最后投递时间
```

### 最佳实践

1. **幂等性**: 确保端点可以处理重复事件
2. **快速响应**: 尽快返回 `2xx` 状态码
3. **异步处理**: 将耗时操作放入后台队列
4. **记录事件 ID**: 防止重复处理

---

## 使用示例

### 1. Slack 通知

```bash
# 创建 Slack Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack Notifications",
    "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXX",
    "secret": "",
    "events": "package:published",
    "enabled": true
  }'
```

**Slack 端点处理：**

```javascript
// Slack 需要特殊的消息格式
app.post('/slack-webhook', (req, res) => {
  const { event, payload } = req.body;
  
  const message = {
    text: `📦 Package ${payload.package} was published by ${payload.publisher}`
  };
  
  // 转发到 Slack
  axios.post(process.env.SLACK_WEBHOOK_URL, message);
  
  res.status(200).send('OK');
});
```

### 2. 钉钉通知

```bash
# 创建钉钉 Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DingTalk Notifications",
    "url": "https://oapi.dingtalk.com/robot/send?access_token=XXX",
    "secret": "SECXXX",
    "events": "package:published,package:unpublished",
    "enabled": true
  }'
```

**钉钉端点处理：**

```javascript
const crypto = require('crypto');

app.post('/dingtalk-webhook', (req, res) => {
  const { event, timestamp, payload } = req.body;
  
  // 计算签名（钉钉要求）
  const secret = 'SECXXX';
  const stringToSign = `${timestamp}\n${secret}`;
  const sign = crypto
    .createHmac('sha256', secret)
    .update(stringToSign)
    .digest()
    .toString('base64');
  
  const message = {
    msgtype: 'text',
    text: {
      content: `📦 包 ${payload.package} ${event === 'package:published' ? '发布' : '删除'}`
    }
  };
  
  // 发送到钉钉
  axios.post(req.query.access_token, message, {
    headers: {
      'Content-Type': 'application/json'
    }
  });
  
  res.status(200).send('OK');
});
```

### 3. CI/CD 集成

```bash
# 创建 CI/CD Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jenkins CI",
    "url": "https://jenkins.example.com/grape-webhook",
    "secret": "ci-secret-key",
    "events": "package:published",
    "enabled": true
  }'
```

**Jenkins Pipeline 触发：**

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    triggers {
        pollSCM('')
    }
    
    stages {
        stage('Build') {
            steps {
                script {
                    // 检查 Grape 包发布事件
                    def event = currentBuild.rawBuild.getAction(hudson.model.CauseAction.class)
                    if (event?.causes?.find { it.shortDescription?.contains('Grape') }) {
                        echo "Building due to Grape package publish..."
                        // 执行构建
                    }
                }
            }
        }
    }
}
```

### 4. 企业微信通知

```bash
# 创建企业微信 Webhook
curl -X POST http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WeCom Notifications",
    "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=XXX",
    "events": "package:published",
    "enabled": true
  }'
```

---

## 故障排查

### 查看投递日志

```bash
# 启用 debug 日志
# config.yaml
log:
  level: "debug"

# 查看日志
tail -f /var/log/grape.log | grep Webhook
```

### 常见问题

#### 1. 签名验证失败

**问题：** 端点返回 `400 Bad Request`

**原因：** 签名不匹配

**解决：**
- 确认 `secret` 配置一致
- 确保使用原始请求体计算签名
- 检查签名头名称 `X-Grape-Signature`

#### 2. 连接超时

**问题：** Webhook 投递超时

**原因：** 端点响应慢或不可达

**解决：**
- 检查端点可用性
- 确保端点在 10 秒内响应
- 使用异步处理

#### 3. 端点不可达

**问题：** 所有重试都失败

**原因：** URL 错误或网络问题

**解决：**
- 确认 URL 可公网访问
- 检查防火墙设置
- 使用 ngrok 等工具测试本地端点

#### 4. 事件未触发

**问题：** 收不到 Webhook 通知

**原因：** Webhook 未启用或事件过滤

**解决：**
```bash
# 检查 Webhook 状态
curl http://localhost:4873/-/api/admin/webhooks \
  -H "Authorization: Bearer <admin_token>"

# 确保 enabled: true
# 确保 events 配置包含目标事件
```

### 测试工具

#### 使用 ngrok 测试本地端点

```bash
# 安装 ngrok
npm install -g ngrok

# 启动本地服务器
node webhook-server.js

# 暴露本地端口
ngrok http 3000

# 使用生成的 URL 配置 Webhook
# https://xxx.ngrok.io/webhook
```

#### 使用 Webhook.site 测试

访问 [webhook.site](https://webhook.site) 获取临时 URL，用于测试 Webhook 投递。

---

## 安全建议

1. **始终使用签名**: 配置 `secret` 验证请求来源
2. **使用 HTTPS**: 确保端点使用 HTTPS
3. **限制 IP**: 如果可能，限制只接受 Grape 服务器 IP
4. **速率限制**: 对 Webhook 端点实施速率限制
5. **监控异常**: 监控失败的投递和异常请求

---

## 相关文档

- [API 文档](API.md) - Webhook API 参考
- [配置指南](../configs/README.md) - 配置文件说明
- [部署指南](DEPLOYMENT.md) - 部署说明
