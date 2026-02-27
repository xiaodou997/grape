# Grape API 文档

本文档详细介绍 Grape 提供的所有 API 端点和使用方法。

## 目录

- [API 概览](#api-概览)
- [认证机制](#认证机制)
- [npm Registry API](#npm-registry-api)
- [管理 API](#管理-api)
- [管理员 API](#管理员-api)
- [Webhook API](#webhook-api)
- [错误响应](#错误响应)

---

## API 概览

### 基础 URL

```
http://localhost:4873
```

### API 分类

| 分类 | 说明 | 认证要求 |
|------|------|----------|
| **npm Registry API** | 兼容 npm 协议的 API | 部分需要 |
| **管理 API** | 包查询、统计等 | 可选 |
| **管理员 API** | 用户管理、Webhook 配置 | 必须管理员 |

---

## 认证机制

### JWT Token 认证

需要认证的 API 使用 Bearer Token 方式：

```http
Authorization: Bearer <jwt_token>
```

### 获取 Token

通过 npm login 命令获取：

```bash
npm login --registry http://localhost:4873
```

响应：

```json
{
  "ok": true,
  "id": "org.couchdb.user:admin",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Token 会保存在 `~/.npmrc` 文件中：

```ini
//localhost:4873/:_authToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token 有效期

默认 24 小时，可通过配置调整：

```yaml
auth:
  jwt_expiry: 8h  # 8 小时
```

---

## npm Registry API

### GET /:package

获取包元数据。

**请求：**

```http
GET /:package
Accept: application/json
```

**响应 200 OK：**

```json
{
  "_id": "@grape/cli",
  "name": "@grape/cli",
  "dist-tags": {
    "latest": "1.2.3",
    "beta": "2.0.0-beta.1"
  },
  "versions": {
    "1.2.3": {
      "name": "@grape/cli",
      "version": "1.2.3",
      "dist": {
        "shasum": "abc123...",
        "tarball": "http://localhost:4873/@grape/cli/-/cli-1.2.3.tgz"
      }
    }
  },
  "time": {
    "created": "2024-01-01T00:00:00Z",
    "1.2.3": "2024-01-02T00:00:00Z"
  },
  "maintainers": [
    {
      "name": "admin",
      "email": "admin@grape.local"
    }
  ],
  "description": "Grape CLI tool",
  "readme": "# @grape/cli\n...",
  "license": "MIT"
}
```

**响应 404 Not Found：**

```json
{
  "code": 4041,
  "message": "package not found"
}
```

**示例：**

```bash
# 获取包信息
curl http://localhost:4873/lodash

# 获取 scoped 包信息
curl http://localhost:4873/@babel/core
```

---

### GET /:package/-/:filename

下载包的 tarball 文件。

**请求：**

```http
GET /:package/-/:filename
```

**响应 200 OK：**

```
Content-Type: application/octet-stream
Content-Length: 12345

<tarball binary data>
```

**响应 404 Not Found：**

```json
{
  "code": 4040,
  "message": "resource not found"
}
```

**示例：**

```bash
# 下载 tarball
curl -O http://localhost:4873/lodash/-/lodash-4.17.21.tgz

# 使用 npm 安装（自动调用此 API）
npm install lodash --registry http://localhost:4873
```

---

### PUT /:package

发布新包或新版本。

**请求：**

```http
PUT /:package
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体：**

```json
{
  "_id": "@grape/cli",
  "name": "@grape/cli",
  "description": "Grape CLI tool",
  "dist-tags": {
    "latest": "1.2.3"
  },
  "versions": {
    "1.2.3": {
      "name": "@grape/cli",
      "version": "1.2.3",
      "dist": {
        "shasum": "abc123...",
        "tarball": "@grape-cli-1.2.3.tgz"
      }
    }
  },
  "_attachments": {
    "@grape-cli-1.2.3.tgz": {
      "content_type": "application/octet-stream",
      "data": "<base64 encoded tarball>"
    }
  },
  "readme": "# @grape/cli\n..."
}
```

**响应 201 Created：**

```json
{
  "ok": true,
  "rev": "1-@grape/cli",
  "success": true
}
```

**响应 401 Unauthorized：**

```json
{
  "code": 4010,
  "message": "authentication required"
}
```

**响应 409 Conflict：**

```json
{
  "code": 4091,
  "message": "version already exists"
}
```

**示例：**

```bash
# 使用 npm 发布
npm publish --registry http://localhost:4873
```

---

### DELETE /:package

删除包或特定版本。

**请求：**

```http
DELETE /:package
Authorization: Bearer <token>
```

**或删除特定版本：**

```http
DELETE /:package/-/:filename
Authorization: Bearer <token>
```

**响应 200 OK：**

```json
{
  "ok": true
}
```

**响应 403 Forbidden：**

```json
{
  "code": 4030,
  "message": "insufficient permissions"
}
```

**示例：**

```bash
# 删除特定版本
npm unpublish @grape/cli@1.2.3 --registry http://localhost:4873

# 删除整个包（谨慎操作）
npm unpublish @grape/cli --force --registry http://localhost:4873
```

---

### PUT /-/user/org.couchdb.user::username

用户登录或注册。

**请求：**

```http
PUT /-/user/org.couchdb.user:admin
Content-Type: application/json
```

**请求体：**

```json
{
  "name": "admin",
  "password": "admin123",
  "email": "admin@example.com",
  "type": "user"
}
```

**响应 200 OK (登录成功)：**

```json
{
  "ok": true,
  "id": "org.couchdb.user:admin",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应 201 Created (注册成功)：**

```json
{
  "ok": true,
  "id": "org.couchdb.user:newuser",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应 401 Unauthorized：**

```json
{
  "error": "invalid credentials"
}
```

**响应 403 Forbidden：**

```json
{
  "error": "registration is disabled, contact your administrator"
}
```

**示例：**

```bash
# 使用 npm 登录
npm login --registry http://localhost:4873

# 或使用 curl
curl -X PUT http://localhost:4873/-/user/org.couchdb.user:admin \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"admin123"}'
```

---

## 管理 API

### GET /-/health

健康检查。

**请求：**

```http
GET /-/health
```

**响应 200 OK：**

```json
{
  "status": "ok",
  "time": "2024-01-01T12:00:00Z"
}
```

**示例：**

```bash
curl http://localhost:4873/-/health
```

---

### GET /-/metrics

Prometheus 格式的性能指标。

**请求：**

```http
GET /-/metrics
```

**响应 200 OK：**

```
# HELP grape_http_requests_total Total number of HTTP requests
# TYPE grape_http_requests_total counter
grape_http_requests_total{method="GET",path="/:package",status="200"} 1234
grape_http_requests_total{method="GET",path="/-/health",status="200"} 567

# HELP grape_http_request_duration_seconds HTTP request duration in seconds
# TYPE grape_http_request_duration_seconds histogram
grape_http_request_duration_seconds_bucket{method="GET",path="/:package",le="0.1"} 1000
...

# HELP grape_package_downloads_total Total number of package tarball downloads
# TYPE grape_package_downloads_total counter
grape_package_downloads_total{package="lodash"} 500
grape_package_downloads_total{package="express"} 300

# HELP grape_stored_packages_total Total number of packages stored locally
# TYPE grape_stored_packages_total gauge
grape_stored_packages_total 42

# HELP grape_registered_users_total Total number of registered users
# TYPE grape_registered_users_total gauge
grape_registered_users_total 5
```

**示例：**

```bash
curl http://localhost:4873/-/metrics
```

---

### GET /-/api/packages

获取所有已缓存的包列表。

**请求：**

```http
GET /-/api/packages
Authorization: Bearer <token>  # 可选
```

**响应 200 OK：**

```json
{
  "packages": [
    {
      "name": "lodash",
      "description": "Lodash modular utilities",
      "version": "4.17.21",
      "private": false,
      "updatedAt": "2024-01-01 12:00:00"
    },
    {
      "name": "@grape/cli",
      "description": "Grape CLI tool",
      "version": "1.2.3",
      "private": true,
      "updatedAt": "2024-01-02 15:30:00"
    }
  ]
}
```

**示例：**

```bash
curl http://localhost:4873/-/api/packages
```

---

### GET /-/api/stats

获取统计信息。

**请求：**

```http
GET /-/api/stats
Authorization: Bearer <token>  # 可选
```

**响应 200 OK：**

```json
{
  "totalPackages": 42,
  "storageSize": 128,
  "upstreams": [
    {
      "name": "npmjs",
      "url": "https://registry.npmjs.org",
      "scope": "",
      "enabled": true
    },
    {
      "name": "company-private",
      "url": "https://npm.company.com",
      "scope": "@company",
      "enabled": true
    }
  ]
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `totalPackages` | int | 已缓存的包总数 |
| `storageSize` | int | 存储占用（MB） |
| `upstreams` | []Upstream | 上游配置列表 |

**示例：**

```bash
curl http://localhost:4873/-/api/stats
```

---

### GET /-/api/search

搜索包。

**请求：**

```http
GET /-/api/search?q=lodash
Authorization: Bearer <token>  # 可选
```

**响应 200 OK：**

```json
{
  "packages": [
    {
      "name": "lodash",
      "description": "Lodash modular utilities",
      "version": "4.17.21",
      "private": false
    },
    {
      "name": "lodash-es",
      "description": "Lodash ES module utilities",
      "version": "4.17.21",
      "private": false
    }
  ],
  "total": 2
}
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `q` | string | 是 | 搜索关键词（支持包名和描述） |

**示例：**

```bash
# 搜索包名
curl "http://localhost:4873/-/api/search?q=lodash"

# 搜索描述
curl "http://localhost:4873/-/api/search?q=utility"
```

---

### GET /-/api/upstreams

获取上游配置。

**请求：**

```http
GET /-/api/upstreams
Authorization: Bearer <token>  # 可选
```

**响应 200 OK：**

```json
{
  "upstreams": [
    {
      "name": "npmjs",
      "url": "https://registry.npmjs.org",
      "scope": "",
      "enabled": true
    },
    {
      "name": "company-private",
      "url": "https://npm.company.com",
      "scope": "@company",
      "enabled": true
    }
  ]
}
```

**示例：**

```bash
curl http://localhost:4873/-/api/upstreams
```

---

### GET /-/api/user

获取当前用户信息。

**请求：**

```http
GET /-/api/user
Authorization: Bearer <token>
```

**响应 200 OK：**

```json
{
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "createdAt": "2024-01-01T00:00:00Z",
  "lastLogin": "2024-01-02T12:00:00Z"
}
```

**响应 401 Unauthorized：**

```json
{
  "error": "not authenticated"
}
```

**示例：**

```bash
curl http://localhost:4873/-/api/user \
  -H "Authorization: Bearer <token>"
```

---

### DELETE /-/api/session

用户登出。

**请求：**

```http
DELETE /-/api/session
Authorization: Bearer <token>
```

**响应 200 OK：**

```json
{
  "ok": true
}
```

**示例：**

```bash
curl -X DELETE http://localhost:4873/-/api/session \
  -H "Authorization: Bearer <token>"
```

---

## 管理员 API

以下 API 需要管理员权限（`role: admin`）。

### GET /-/api/admin/users

获取用户列表。

**请求：**

```http
GET /-/api/admin/users
Authorization: Bearer <admin_token>
```

**响应 200 OK：**

```json
{
  "users": [
    {
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "createdAt": "2024-01-01T00:00:00Z",
      "lastLogin": "2024-01-02T12:00:00Z"
    },
    {
      "username": "developer",
      "email": "dev@example.com",
      "role": "developer",
      "createdAt": "2024-01-01T10:00:00Z",
      "lastLogin": "2024-01-02T09:00:00Z"
    }
  ]
}
```

**示例：**

```bash
curl http://localhost:4873/-/api/admin/users \
  -H "Authorization: Bearer <admin_token>"
```

---

### POST /-/api/admin/users

创建新用户。

**请求：**

```http
POST /-/api/admin/users
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体：**

```json
{
  "name": "newuser",
  "password": "SecurePass123!",
  "email": "newuser@example.com",
  "role": "developer"
}
```

**响应 201 Created：**

```json
{
  "ok": true,
  "username": "newuser",
  "role": "developer"
}
```

**响应 400 Bad Request：**

```json
{
  "error": "password must be at least 8 characters"
}
```

**响应 409 Conflict：**

```json
{
  "error": "user already exists"
}
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 用户名 |
| `password` | string | 是 | 密码（至少 8 字符） |
| `email` | string | 否 | 邮箱 |
| `role` | string | 否 | 角色：`admin` 或 `developer`，默认 `developer` |

**示例：**

```bash
curl -X POST http://localhost:4873/-/api/admin/users \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"newuser","password":"SecurePass123!","email":"newuser@example.com"}'
```

---

### PUT /-/api/admin/users/:username

更新用户信息。

**请求：**

```http
PUT /-/api/admin/users/admin
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体：**

```json
{
  "email": "newemail@example.com",
  "password": "NewSecurePass123!",  // 可选，留空则不修改
  "role": "admin"                   // 可选
}
```

**响应 200 OK：**

```json
{
  "ok": true,
  "username": "admin"
}
```

**示例：**

```bash
curl -X PUT http://localhost:4873/-/api/admin/users/developer \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"email":"newdev@example.com"}'
```

---

### DELETE /-/api/admin/users/:username

删除用户。

**请求：**

```http
DELETE /-/api/admin/users/developer
Authorization: Bearer <admin_token>
```

**响应 200 OK：**

```json
{
  "ok": true
}
```

**响应 403 Forbidden：**

```json
{
  "error": "cannot delete admin user"
}
```

**示例：**

```bash
curl -X DELETE http://localhost:4873/-/api/admin/users/developer \
  -H "Authorization: Bearer <admin_token>"
```

---

## Webhook API

### GET /-/api/admin/webhooks

获取 Webhook 列表。

**请求：**

```http
GET /-/api/admin/webhooks
Authorization: Bearer <admin_token>
```

**响应 200 OK：**

```json
{
  "webhooks": [
    {
      "id": 1,
      "name": "Slack Notification",
      "url": "https://hooks.slack.com/services/xxx",
      "events": "package:published,package:unpublished",
      "enabled": true,
      "createdAt": "2024-01-01T00:00:00Z",
      "lastDeliveryAt": "2024-01-02T12:00:00Z"
    }
  ]
}
```

---

### POST /-/api/admin/webhooks

创建 Webhook。

**请求：**

```http
POST /-/api/admin/webhooks
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体：**

```json
{
  "name": "Slack Notification",
  "url": "https://hooks.slack.com/services/xxx",
  "secret": "hmac-secret-key",
  "events": "package:published,package:unpublished",
  "enabled": true
}
```

**响应 201 Created：**

```json
{
  "ok": true,
  "id": 1
}
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | Webhook 名称 |
| `url` | string | 是 | 接收端点 URL |
| `secret` | string | 否 | HMAC 签名密钥 |
| `events` | string | 否 | 逗号分隔的事件类型，空表示所有事件 |
| `enabled` | bool | 否 | 是否启用，默认 `true` |

**事件类型：**

| 事件 | 说明 |
|------|------|
| `package:published` | 包发布 |
| `package:unpublished` | 包删除 |
| `user:created` | 用户创建 |
| `user:deleted` | 用户删除 |

---

### PUT /-/api/admin/webhooks/:id

更新 Webhook。

**请求：**

```http
PUT /-/api/admin/webhooks/1
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体：**

```json
{
  "name": "Updated Name",
  "url": "https://new-url.com/webhook",
  "enabled": false
}
```

---

### DELETE /-/api/admin/webhooks/:id

删除 Webhook。

**请求：**

```http
DELETE /-/api/admin/webhooks/1
Authorization: Bearer <admin_token>
```

---

### POST /-/api/admin/webhooks/:id/test

测试 Webhook。

**请求：**

```http
POST /-/api/admin/webhooks/1/test
Authorization: Bearer <admin_token>
```

**响应 200 OK：**

```json
{
  "ok": true,
  "message": "Test event sent"
}
```

---

## 错误响应

### 错误格式

```json
{
  "code": 4041,
  "message": "package not found",
  "reason": "lodash"
}
```

### 错误码列表

| 错误码 | HTTP 状态 | 说明 |
|--------|----------|------|
| `4000` | 400 | 错误请求 |
| `4001` | 400 | 无效请求体 |
| `4010` | 401 | 需要认证 |
| `4011` | 401 | 凭证无效 |
| `4030` | 403 | 权限不足 |
| `4040` | 404 | 资源不存在 |
| `4041` | 404 | 包不存在 |
| `4042` | 404 | 用户不存在 |
| `4090` | 409 | 资源冲突 |
| `4091` | 409 | 版本已存在 |
| `4092` | 409 | 用户已存在 |
| `4290` | 429 | 请求过于频繁 |
| `5000` | 500 | 服务器内部错误 |

---

## 使用示例

### 使用 curl 调用 API

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X PUT http://localhost:4873/-/user/org.couchdb.user:admin \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"admin123"}' \
  | jq -r '.token')

# 2. 获取当前用户信息
curl http://localhost:4873/-/api/user \
  -H "Authorization: Bearer $TOKEN"

# 3. 获取包列表
curl http://localhost:4873/-/api/packages \
  -H "Authorization: Bearer $TOKEN"

# 4. 获取统计信息
curl http://localhost:4873/-/api/stats
```

### 使用 npm 命令

```bash
# 登录
npm login --registry http://localhost:4873

# 安装包
npm install lodash --registry http://localhost:4873

# 发布包
npm publish --registry http://localhost:4873

# 删除包
npm unpublish my-package@1.0.0 --registry http://localhost:4873
```

---

## 相关文档

- [使用指南](USAGE.md) - 包管理器配置
- [Webhook 文档](WEBHOOKS.md) - Webhook 详细使用
- [配置指南](../configs/README.md) - 配置文件说明
