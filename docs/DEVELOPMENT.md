# Grape 开发文档

本文档介绍如何参与 Grape 的开发工作。

## 目录

- [项目架构](#项目架构)
- [开发环境搭建](#开发环境搭建)
- [代码结构](#代码结构)
- [核心模块说明](#核心模块说明)
- [开发工作流](#开发工作流)
- [测试指南](#测试指南)
- [代码规范](#代码规范)
- [贡献指南](#贡献指南)

---

## 项目架构

### 技术栈

| 组件 | 技术 | 版本 | 说明 |
|------|------|------|------|
| **后端语言** | Go | 1.21+ | 高性能、并发、单一二进制 |
| **Web 框架** | Gin | v1.11.0 | 高性能 HTTP 框架 |
| **配置管理** | Viper | v1.21.0 | 多格式配置支持 |
| **日志** | Zap | v1.27.1 | 高性能结构化日志 |
| **认证** | JWT | v5.3.1 | 无状态认证 |
| **数据库** | GORM + SQLite | v1.31.1 | ORM 框架 |
| **指标** | Prometheus | v1.23.2 | 监控指标 |
| **前端框架** | Vue 3 | v3.5.x | 响应式前端 |
| **UI 组件** | Element Plus | v2.13.x | Vue 3 组件库 |
| **构建工具** | Vite | v7.x | 快速构建 |

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│                    npm/yarn/pnpm/bun                     │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                     HTTP Server (Gin)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Auth      │  │  Registry   │  │    API      │     │
│  │ Middleware  │  │   Handler   │  │   Handler   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────┬───────────────┬─────────────────┬─────────────┘
          │               │                 │
          ▼               ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Auth Module    │ │ Registry Core   │ │  Storage Layer  │
│  - JWT          │ │  - Multi-upstream│ │  - Local FS     │
│  - User Store   │ │  - Proxy cache   │ │  - Metadata     │
│  - Password     │ │  - Tarball       │ │  - Tarball      │
└─────────────────┘ └─────────────────┘ └─────────────────┘
          │               │                 │
          ▼               ▼                 ▼
┌─────────────────────────────────────────────────────────┐
│                     Database (SQLite)                     │
│  - Users                                                  │
│  - Packages                                               │
│  - PackageVersions                                        │
│  - Webhooks                                               │
└─────────────────────────────────────────────────────────┘
```

---

## 开发环境搭建

### 1. 安装依赖

```bash
# Go 1.21+
go version

# Node.js 18+ (前端开发)
node --version
npm --version
```

### 2. 克隆项目

```bash
git clone https://github.com/graperegistry/grape.git
cd grape
```

### 3. 安装 Go 依赖

```bash
go mod download
go mod tidy
```

### 4. 安装前端依赖

```bash
cd web
npm install
cd ..
```

### 5. 运行开发环境

```bash
# 方式一：分别启动前后端
# 终端 1 - 后端
make run

# 终端 2 - 前端
cd web
npm run dev

# 方式二：使用 Makefile 一键启动
make dev
```

### 6. 验证安装

```bash
# 访问后端
curl http://localhost:4873/-/health

# 访问前端
open http://localhost:3000
```

---

## 代码结构

```
grape/
├── cmd/grape/                  # 程序入口
│   └── main.go                 # main 函数，初始化组件
│
├── internal/                   # 内部包（不对外暴露）
│   ├── auth/                   # 认证模块
│   │   ├── user.go             # 用户模型和操作
│   │   ├── db_user.go          # 数据库用户存储实现
│   │   ├── jwt.go              # JWT 生成和验证
│   │   ├── jwt_test.go         # JWT 测试
│   │   ├── user_test.go        # 用户测试
│   │   └── middleware.go       # 认证中间件
│   │
│   ├── config/                 # 配置管理
│   │   ├── config.go           # 配置结构定义
│   │   └── loader.go           # 配置加载
│   │
│   ├── db/                     # 数据库
│   │   ├── db.go               # 数据库初始化和连接
│   │   └── models.go           # GORM 模型定义
│   │
│   ├── logger/                 # 日志系统
│   │   └── logger.go           # Zap 日志封装
│   │
│   ├── metrics/                # Prometheus 指标
│   │   └── metrics.go          # 指标定义
│   │
│   ├── registry/               # npm registry 核心
│   │   ├── proxy.go            # 多上游代理
│   │   └── errors.go           # registry 错误定义
│   │
│   ├── server/                 # HTTP 服务器
│   │   ├── server.go           # 服务器主逻辑
│   │   └── handler/            # 请求处理器
│   │       ├── api.go          # 管理 API handler
│   │       ├── auth.go         # 认证 API handler
│   │       ├── publish.go      # 发布 API handler
│   │       ├── registry.go     # registry API handler
│   │       └── webhook.go      # webhook API handler
│   │
│   ├── storage/                # 存储抽象层
│   │   ├── storage.go          # 存储接口定义
│   │   └── local/              # 本地存储实现
│   │       └── storage.go      # 本地文件系统存储
│   │
│   ├── webhook/                # Webhook 事件通知
│   │   └── webhook.go          # Webhook 分发器
│   │
│   └── web/                    # 前端嵌入
│       └── embed.go            # embed.FS 定义
│
├── pkg/                        # 公共包（可对外暴露）
│   └── apierr/                 # 统一错误码
│       └── errors.go           # 错误码定义
│
├── web/                        # 前端源码
│   ├── src/
│   │   ├── api/                # API 调用
│   │   ├── components/         # Vue 组件
│   │   ├── router/             # 路由配置
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── views/              # 页面视图
│   │   └── App.vue             # 根组件
│   └── package.json
│
├── configs/                    # 配置文件示例
│   ├── config.yaml             # 默认配置
│   └── README.md               # 配置文档
│
├── docs/                       # 文档
│   ├── API.md                  # API 文档
│   ├── DEPLOYMENT.md           # 部署文档
│   ├── DEVELOPMENT.md          # 开发文档
│   ├── USAGE.md                # 使用指南
│   └── WEBHOOKS.md             # Webhook 文档
│
├── data/                       # 数据目录（运行时生成）
│   ├── grape.db                # SQLite 数据库
│   └── packages/               # 包存储目录
│
├── bin/                        # 编译产物
│   └── grape                   # 可执行文件
│
├── Makefile                    # 构建脚本
├── go.mod                      # Go 模块定义
├── go.sum                      # Go 依赖校验
└── README.md                   # 项目说明
```

---

## 核心模块说明

### 1. 认证模块 (internal/auth)

#### 主要功能

- JWT Token 生成和验证
- 用户密码哈希和验证
- 用户 CRUD 操作
- 认证中间件

#### 关键代码

```go
// internal/auth/jwt.go
type JWTService struct {
    secret string
    expiry time.Duration
}

func (s *JWTService) GenerateToken(user *User) (string, error) {
    claims := jwt.MapClaims{
        "username": user.Username,
        "role":     user.Role,
        "exp":      time.Now().Add(s.expiry).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.secret))
}

// internal/auth/middleware.go
func AuthMiddleware(jwtService *JWTService, userStore UserStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        // 解析和验证 JWT token
        // 将用户信息存入 context
        c.Next()
    }
}
```

### 2. Registry 模块 (internal/registry)

#### 主要功能

- 多上游代理
- Scope 路由
- 包元数据获取
- Tarball 下载

#### 关键代码

```go
// internal/registry/proxy.go
type Proxy struct {
    upstreams []*Upstream
    defaultUp *Upstream
    scopeMap  map[string]*Upstream
}

func (p *Proxy) selectUpstream(packageName string) *Upstream {
    // 提取 scope
    if strings.HasPrefix(packageName, "@") {
        idx := strings.Index(packageName, "/")
        if idx > 0 {
            scope := packageName[:idx]
            if up, ok := p.scopeMap[scope]; ok {
                return up
            }
        }
    }
    return p.defaultUp
}

func (p *Proxy) GetMetadata(packageName string) ([]byte, error) {
    up := p.selectUpstream(packageName)
    // 从上游获取元数据
}
```

### 3. 存储模块 (internal/storage)

#### 接口定义

```go
// internal/storage/storage.go
type Storage interface {
    HasPackage(name string) bool
    GetMetadata(name string) ([]byte, error)
    SaveMetadata(name string, data []byte) error
    DeletePackage(name string) error

    HasTarball(name, filename string) bool
    GetTarball(name, filename string) ([]byte, error)
    SaveTarball(name, filename string, data []byte) error
    DeleteTarball(name, filename string) error

    ListPackages() ([]PackageInfo, error)
    GetStats() (*StorageStats, error)
}
```

#### 本地存储实现

```go
// internal/storage/local/storage.go
type Storage struct {
    basePath string
}

func (s *Storage) SaveTarball(name, filename string, data []byte) error {
    path := filepath.Join(s.basePath, "packages", name, "tarballs", filename)
    return os.WriteFile(path, data, 0644)
}
```

### 4. HTTP 处理器 (internal/server/handler)

#### Registry Handler

```go
// internal/server/handler/registry.go
type RegistryHandler struct {
    proxy   *registry.Proxy
    storage *local.Storage
}

func (h *RegistryHandler) GetPackage(c *gin.Context) {
    packageName := c.Param("package")
    // 1. 检查本地缓存
    // 2. 从上游获取
    // 3. 缓存并返回
}
```

#### Publish Handler

```go
// internal/server/handler/publish.go
type PublishHandler struct {
    storage *local.Storage
    locks   sync.Map  // 包名 -> *sync.Mutex
}

func (h *PublishHandler) Publish(c *gin.Context) {
    // 1. 验证认证
    // 2. 获取包锁（防止并发冲突）
    // 3. 检查版本是否存在
    // 4. 保存 tarball
    // 5. 更新元数据
}
```

---

## 开发工作流

### 1. 创建分支

```bash
# 从 main 分支创建特性分支
git checkout -b feature/your-feature-name

# 或修复 bug
git checkout -b fix/bug-fix-name
```

### 2. 开发和测试

```bash
# 编写代码
# ...

# 运行测试
make test

# 代码格式化
make fmt

# 运行 linter
make lint
```

### 3. 提交代码

```bash
# 添加更改
git add .

# 提交（遵循 Conventional Commits）
git commit -m "feat: add new feature"
# 或
git commit -m "fix: resolve bug"
```

### 4. 推送和 PR

```bash
# 推送到远程
git push origin feature/your-feature-name

# 在 GitHub 上创建 Pull Request
```

---

## 测试指南

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示详细输出
go test ./... -v

# 运行覆盖率
go test ./... -cover

# 运行特定包的测试
go test ./internal/auth -v

# 运行特定测试函数
go test ./internal/auth -run TestJWT
```

### 编写测试

```go
// internal/auth/jwt_test.go
package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestJWTService_GenerateToken(t *testing.T) {
    jwtService := NewJWTService("test-secret", time.Hour)
    user := &User{
        Username: "testuser",
        Role:     "developer",
    }

    token, err := jwtService.GenerateToken(user)

    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestJWTService_ValidateToken(t *testing.T) {
    jwtService := NewJWTService("test-secret", time.Hour)
    
    // 生成有效 token
    token, _ := jwtService.GenerateToken(&User{Username: "test"})
    claims, err := jwtService.ValidateToken(token)
    
    assert.NoError(t, err)
    assert.Equal(t, "test", claims.Username)
}
```

### 测试覆盖率目标

| 模块 | 目标覆盖率 |
|------|-----------|
| auth | 80%+ |
| registry | 70%+ |
| storage | 80%+ |
| server/handler | 60%+ |

---

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://golang.org/doc/effective_go.html) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

#### 命名规范

```go
// 包名：小写，简短
package auth

// 类型：大驼峰
type User struct {}

// 接口：-er 后缀
type Storage interface {}

// 函数：大驼峰导出，小驼峰私有
func NewUser() *User {}
func (u *User) validate() error {}

// 常量：全大写
const MaxSize = 1024
```

#### 错误处理

```go
// 错误检查
if err != nil {
    return nil, fmt.Errorf("failed to xxx: %w", err)
}

// 自定义错误
var ErrUserNotFound = &APIError{Code: 4042, Message: "user not found"}
```

#### 注释规范

```go
// User 用户模型
type User struct {
    // Username 用户名
    Username string `json:"username"`
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
    // 实现...
}
```

### 前端代码规范

#### Vue 组件结构

```vue
<template>
  <div class="component-name">
    <!-- 模板内容 -->
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// Props
const props = defineProps<{
  title: string
}>()

// Emits
const emit = defineEmits<{
  (e: 'update', value: string): void
}>()

// 状态
const loading = ref(false)

// 方法
const handleClick = () => {
  // ...
}
</script>

<style scoped>
.component-name {
  /* 样式 */
}
</style>
```

#### TypeScript 类型定义

```typescript
// src/types/index.ts
export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'developer'
}

export interface Package {
  name: string
  version: string
  description: string
  private: boolean
}
```

---

## 贡献指南

### 提交 PR

1. **Fork 仓库**
2. **创建分支**: `git checkout -b feature/amazing-feature`
3. **提交更改**: `git commit -m 'feat: add amazing feature'`
4. **推送分支**: `git push origin feature/amazing-feature`
5. **创建 Pull Request**

### Commit 规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

#### Type 类型

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响代码运行） |
| `refactor` | 重构（既不是新功能也不是 bug 修复） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建/工具/配置相关 |

#### 示例

```
feat(auth): add LDAP authentication support

- Add LDAP client integration
- Add LDAP configuration options
- Add user sync mechanism

Closes #123
```

```
fix(registry): resolve tarball download issue for scoped packages

- Fix URL encoding for scoped package names
- Add test cases for scoped packages
```

### 代码审查清单

在提交 PR 前，请确保：

- [ ] 代码已通过 `go fmt` 格式化
- [ ] 代码已通过 `go vet` 检查
- [ ] 所有测试通过
- [ ] 测试覆盖率未下降
- [ ] 添加了必要的单元测试
- [ ] 更新了相关文档
- [ ] Commit 信息符合规范

---

## 调试技巧

### 后端调试

```bash
# 使用 delve 调试
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/grape

# 设置断点
(dlv) break main.go:50
(dlv) continue
```

### 前端调试

```bash
# 启动开发服务器
cd web
npm run dev

# 在浏览器中打开开发者工具
# http://localhost:3000
```

### 日志调试

```yaml
# 配置文件中设置 debug 级别
log:
  level: "debug"
```

```bash
# 查看调试日志
tail -f /var/log/grape.log | grep DEBUG
```

---

## 性能分析

### CPU 分析

```bash
# 启动 pprof
go tool pprof http://localhost:4873/debug/pprof/profile

# 查看分析结果
(pprof) top10
(pprof) web
```

### 内存分析

```bash
# 获取堆内存分析
go tool pprof http://localhost:4873/debug/pprof/heap

# 查看分析结果
(pprof) top10
```

---

## 相关文档

- [API 文档](API.md) - API 参考
- [配置指南](../configs/README.md) - 配置详解
- [部署指南](DEPLOYMENT.md) - 部署说明
- [README](../README.md) - 项目介绍
