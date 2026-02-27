# Grape 项目概述

> 🍇 Grape - 轻盈如风的企业级私有 npm 仓库

本文档提供 Grape 项目的高层次概述，帮助新贡献者快速了解项目。

---

## 项目定位

**Grape** 是一个用 Go 语言编写的**轻量级、高性能**私有 npm 仓库，完美兼容 npm/yarn/pnpm/bun 客户端。

### 核心特性

- 🚀 **单一二进制** - 无需 Node.js，下载即用
- 📦 **npm 兼容** - 完整支持 npm/yarn/pnpm/bun
- 🔀 **多上游路由** - 按 scope 路由到不同上游
- 🔐 **用户认证** - JWT 认证，SQLite 持久化
- 🗄️ **智能缓存** - 自动缓存公共包
- 🌐 **现代 Web UI** - Vue 3 + Element Plus
- 🪶 **轻量级** - 内存占用 < 10MB

---

## 技术架构

### 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| **后端** | Go 1.21+ | 高性能、并发 |
| **Web 框架** | Gin | HTTP 服务器 |
| **配置** | Viper | 多格式配置 |
| **日志** | Zap | 结构化日志 |
| **认证** | JWT | 无状态认证 |
| **数据库** | GORM + SQLite | ORM + 持久化 |
| **指标** | Prometheus | 监控指标 |
| **前端** | Vue 3 + TS | 响应式 UI |
| **UI** | Element Plus | 组件库 |

### 目录结构

```
grape/
├── cmd/grape/              # 程序入口
├── internal/               # 内部包
│   ├── auth/               # 认证模块
│   ├── config/             # 配置管理
│   ├── db/                 # 数据库
│   ├── logger/             # 日志系统
│   ├── metrics/            # Prometheus 指标
│   ├── registry/           # npm registry 核心
│   ├── server/             # HTTP 服务器
│   ├── storage/            # 存储抽象
│   ├── webhook/            # Webhook 通知
│   └── web/                # 前端嵌入
├── pkg/                    # 公共包
│   └── apierr/             # 统一错误码
├── web/                    # 前端源码
├── configs/                # 配置示例
├── docs/                   # 文档
└── Makefile
```

---

## 核心模块

### 1. 认证模块 (auth)

- JWT Token 生成和验证
- 用户密码哈希（bcrypt）
- 用户 CRUD 操作
- 认证中间件

**关键文件：**
- `user.go` - 用户模型
- `db_user.go` - 数据库用户存储
- `jwt.go` - JWT 服务
- `middleware.go` - 认证中间件

### 2. Registry 模块 (registry)

- 多上游代理
- Scope 路由
- 包元数据获取
- Tarball 下载

**关键文件：**
- `proxy.go` - 多上游代理逻辑

### 3. 存储模块 (storage)

- 存储接口抽象
- 本地文件系统实现
- 包元数据和 tarball 管理

**关键文件：**
- `storage.go` - 存储接口定义
- `local/storage.go` - 本地存储实现

### 4. HTTP 处理器 (server/handler)

- Registry Handler - npm 协议处理
- Publish Handler - 包发布处理
- Auth Handler - 认证处理
- API Handler - 管理 API 处理

### 5. 前端 (web)

- Vue 3 + TypeScript
- Element Plus UI
- Pinia 状态管理
- Vue Router 路由

**关键目录：**
- `src/api/` - API 调用
- `src/stores/` - 状态管理
- `src/views/` - 页面视图

---

## 核心流程

### npm install 流程

```
npm install lodash
       ↓
检查本地缓存
       ↓
   ┌──────┐
   │ 有   │────→ 返回缓存数据
   └──────┘
       │ 无
       ↓
请求上游 (npmjs.org)
       ↓
缓存到本地
       ↓
返回数据
```

### npm publish 流程

```
npm publish
       ↓
验证 JWT Token
       ↓
获取包锁（并发保护）
       ↓
检查版本是否存在
       ↓
解码并保存 tarball
       ↓
更新元数据
       ↓
触发 Webhook 通知
       ↓
返回成功
```

### 用户认证流程

```
npm login
       ↓
验证用户名密码
       ↓
生成 JWT Token
       ↓
返回 Token
       ↓
客户端保存到 .npmrc
       ↓
后续请求携带 Token
```

---

## 开发快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ (前端开发)

### 搭建开发环境

```bash
# 克隆项目
git clone https://github.com/graperegistry/grape.git
cd grape

# 安装依赖
go mod download
cd web && npm install && cd ..

# 运行开发环境
make run  # 后端
cd web && npm run dev  # 前端
```

### 运行测试

```bash
# 所有测试
go test ./... -v

# 覆盖率
go test ./... -cover
```

### 构建

```bash
# 完整构建（前后端）
make build

# 仅后端
make build-only

# 仅前端
make build-frontend
```

---

## 版本历史

### v0.1.0 (当前版本)

**已完成功能：**
- ✅ npm 代理缓存
- ✅ 多上游支持（按 scope 路由）
- ✅ 用户认证（JWT + SQLite）
- ✅ 包发布/删除
- ✅ Web 管理界面
- ✅ 登录限流保护
- ✅ 包发布并发锁
- ✅ Webhook 通知
- ✅ Prometheus 指标

### 未来规划

**v0.2.0 (计划中)**
- RBAC 权限模型
- PostgreSQL 支持
- 操作审计日志

**v0.3.0 (计划中)**
- Redis 缓存
- S3/MinIO 存储
- 垃圾回收机制

**v1.0.0 (计划中)**
- LDAP/OIDC 集成
- 高可用集群
- Docker 镜像

---

## 文档导航

| 文档 | 说明 |
|------|------|
| [README](../README.md) | 项目介绍和快速开始 |
| [使用指南](USAGE.md) | 包管理器配置和使用 |
| [API 文档](API.md) | 完整 API 参考 |
| [部署指南](DEPLOYMENT.md) | 生产环境部署 |
| [开发文档](DEVELOPMENT.md) | 开发者指南 |
| [Webhook 文档](WEBHOOKS.md) | Webhook 使用 |
| [配置指南](../configs/README.md) | 配置文件详解 |

---

## 贡献

欢迎贡献代码、文档或建议！

1. Fork 仓库
2. 创建特性分支
3. 提交更改
4. 创建 Pull Request

详见 [贡献指南](DEVELOPMENT.md#贡献指南)。

---

<p align="center">
  Made with ❤️ by the Grape Team<br>
  🍇 轻盈如风，功能如山
</p>
