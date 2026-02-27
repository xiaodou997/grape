# 🍇 Grape

> **轻盈如风的企业级私有 npm 仓库**  
> One binary, zero debt. 一个二进制，零负担。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![npm compatible](https://img.shields.io/badge/npm-compatible-brightgreen)](https://www.npmjs.com)

Grape 是一个用 Go 语言编写的**轻量级、高性能**私有 npm 仓库，完美兼容 npm/yarn/pnpm/bun 客户端。相比 Verdaccio，它提供**更低的资源占用**、**更强大的权限控制**和**更现代化的 Web 界面**。

## ✨ 特性

| 特性 | 说明 |
|------|------|
| 🚀 **单一二进制** | 无需 Node.js，下载即用，部署极简 |
| 📦 **npm 兼容** | 完整支持 npm/yarn/pnpm/bun，零学习成本 |
| 🔀 **多上游路由** | 按 scope 路由到不同上游，支持私有仓库 |
| 🔐 **用户认证** | JWT 认证，SQLite 持久化，支持发布私有包 |
| 💾 **数据持久化** | SQLite 存储，重启不丢失 |
| 🗄️ **智能缓存** | 自动缓存公共包，加速团队开发 |
| 🌐 **现代 Web UI** | Vue 3 + Element Plus 管理界面 |
| 🪶 **轻量级** | 内存占用 < 10MB，远低于 Verdaccio |
| 🔔 **Webhook 通知** | 支持包发布/删除事件通知 |
| 📊 **Prometheus 指标** | 完整的监控指标支持 |

## 🚀 快速开始

### 方式一：下载预编译二进制

```bash
# macOS (Intel)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-darwin-amd64 -o grape
chmod +x grape

# macOS (Apple Silicon)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-darwin-arm64 -o grape
chmod +x grape

# Linux (amd64)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-linux-amd64 -o grape
chmod +x grape

# Linux (arm64)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-linux-arm64 -o grape
chmod +x grape

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/graperegistry/grape/releases/latest/download/grape-windows-amd64.exe" -OutFile "grape.exe"

# 运行
./grape

# 访问
open http://localhost:4873
```

### 方式二：从源码构建

```bash
# 克隆仓库
git clone https://github.com/graperegistry/grape.git
cd grape

# 构建 (包含前端)
make build

# 运行
./bin/grape

# 或使用配置文件
./bin/grape -c ./configs/config.yaml
```

### 方式三：Docker 部署

```bash
# 运行容器
docker run -d \
  --name grape \
  -p 4873:4873 \
  -v grape-data:/data \
  graperegistry/grape:latest

# 查看日志
docker logs -f grape
```

## 📖 使用方法

### 1. 配置 npm

```bash
# 全局设置（所有包都使用 Grape）
npm set registry http://localhost:4873

# 或仅设置特定 scope（推荐）
npm set @mycompany:registry http://localhost:4873

# 查看当前配置
npm config list
```

### 2. 安装包

```bash
# 安装公共包（自动从上游缓存）
npm install lodash
npm install express

# 安装 scoped 包
npm install @babel/core
npm install @mycompany/utils
```

### 3. 发布私有包

```bash
# 登录（默认账户：admin / admin）
npm login --registry http://localhost:4873

# 发布包
npm publish --registry http://localhost:4873

# 发布 beta 版本
npm publish --tag beta --registry http://localhost:4873
```

### 4. 删除包

```bash
# 删除特定版本
npm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# 删除整个包（谨慎操作）
npm unpublish @mycompany/my-package --force --registry http://localhost:4873
```

### 5. 恢复默认源

```bash
npm set registry https://registry.npmjs.org
```

## 🌊 多上游配置

Grape 支持配置多个上游仓库，按包的 scope 自动路由：

```yaml
# config.yaml
registry:
  upstreams:
    # 默认上游（公共包）
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""              # 空字符串表示默认
      timeout: 30s
      enabled: true

    # 淘宝镜像（可选加速）
    - name: "npmmirror"
      url: "https://registry.npmmirror.com"
      scope: ""
      timeout: 15s
      enabled: false

    # 公司私有包
    - name: "company-private"
      url: "https://npm.company.com"
      scope: "@company"      # 所有 @company/* 包
      timeout: 30s
      enabled: true

    # 内部工具包
    - name: "internal-tools"
      url: "https://npm-internal.company.com"
      scope: "@internal"
      timeout: 30s
      enabled: true
```

| 包名 | 路由到 |
|------|--------|
| `lodash` | npmjs (默认) |
| `@babel/core` | npmjs (默认) |
| `@company/utils` | company-private |
| `@internal/cli` | internal-tools |

## 🔧 配置说明

### 完整配置文件示例

```yaml
# server 配置
server:
  host: "0.0.0.0"
  port: 4873
  read_timeout: 30s
  write_timeout: 30s

# registry 配置
registry:
  upstreams:
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""
      timeout: 30s
      enabled: true

# storage 配置
storage:
  type: "local"
  path: "./data"

# log 配置
log:
  level: "info"  # debug, info, warn, error

# auth 配置
auth:
  jwt_secret: "your-secret-key-change-in-production"  # ⚠️ 生产环境必须修改
  jwt_expiry: 24h
  allow_registration: false  # 是否允许自助注册

# database 配置
database:
  type: "sqlite"
  dsn: "./data/grape.db"
```

### 配置项说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `server.host` | string | 0.0.0.0 | 监听地址 |
| `server.port` | int | 4873 | 监听端口 |
| `server.read_timeout` | duration | 30s | 请求读取超时 |
| `server.write_timeout` | duration | 30s | 响应写入超时 |
| `registry.upstreams` | []Upstream | - | 多上游配置 |
| `storage.type` | string | local | 存储类型 |
| `storage.path` | string | ./data | 数据存储路径 |
| `log.level` | string | info | 日志级别 |
| `auth.jwt_secret` | string | - | JWT 签名密钥 |
| `auth.jwt_expiry` | duration | 24h | Token 有效期 |
| `database.type` | string | sqlite | 数据库类型 |
| `database.dsn` | string | ./data/grape.db | 数据库连接字符串 |

## 🌐 Web UI

访问 Web 管理界面：http://localhost:4873

### 功能

- 📦 **包浏览** - 查看已缓存的包列表和详情
- 👤 **用户管理** - 创建/删除用户，分配角色
- 📊 **系统监控** - 查看统计信息和服务状态
- 🔔 **Webhook 配置** - 配置事件通知

### 默认账户

首次启动时自动创建管理员账户：

- **用户名**: `admin`
- **密码**: `admin`

⚠️ **生产环境请立即修改默认密码！**

## 🔌 API 端点

### npm Registry API

| 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/:package` | GET | 获取包元数据 | 可选 |
| `/:package/-/:filename` | GET | 下载 tarball | 可选 |
| `/:package` | PUT | 发布包 | 必须 |
| `/:package` | DELETE | 删除包 | 必须 |
| `/-/user/:username` | PUT | 用户登录/注册 | 否 |

### 管理 API

| 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/-/health` | GET | 健康检查 | 否 |
| `/-/metrics` | GET | Prometheus 指标 | 否 |
| `/-/api/packages` | GET | 包列表 | 可选 |
| `/-/api/stats` | GET | 统计信息 | 可选 |
| `/-/api/search?q=` | GET | 搜索包 | 可选 |
| `/-/api/upstreams` | GET | 上游配置 | 可选 |
| `/-/api/user` | GET | 当前用户信息 | 必须 |

### 管理员 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/-/api/admin/users` | GET | 用户列表 |
| `/-/api/admin/users` | POST | 创建用户 |
| `/-/api/admin/users/:username` | PUT | 更新用户 |
| `/-/api/admin/users/:username` | DELETE | 删除用户 |
| `/-/api/admin/webhooks` | GET/POST/PUT/DELETE | Webhook 管理 |

## 📊 监控指标

Grape 提供 Prometheus 格式的监控指标，访问 `http://localhost:4873/-/metrics`：

- `grape_http_requests_total` - HTTP 请求总数
- `grape_http_request_duration_seconds` - HTTP 请求耗时
- `grape_package_downloads_total` - 包下载次数
- `grape_package_publish_total` - 包发布次数
- `grape_proxy_requests_total` - 上游代理请求次数
- `grape_stored_packages_total` - 已缓存包数量
- `grape_registered_users_total` - 注册用户数

## 🗂️ 项目结构

```
grape/
├── cmd/grape/              # 程序入口
├── internal/
│   ├── auth/               # 用户认证 (JWT + SQLite)
│   ├── config/             # 配置管理 (Viper)
│   ├── db/                 # 数据库模型 (GORM)
│   ├── logger/             # 日志系统 (Zap)
│   ├── metrics/            # Prometheus 指标
│   ├── registry/           # npm registry 核心 (多上游代理)
│   ├── server/             # HTTP 服务 (Gin)
│   ├── storage/            # 存储抽象层
│   ├── webhook/            # Webhook 事件通知
│   └── web/                # 前端嵌入
├── pkg/apierr/             # 统一错误码
├── web/                    # 前端源码 (Vue 3 + Element Plus)
├── configs/                # 配置示例
├── docs/                   # 文档
├── data/                   # 数据目录
└── Makefile
```

## 🛠️ 开发

### 环境要求

- Go 1.21+
- Node.js 18+ (仅前端开发需要)

### 开发命令

```bash
# 构建后端（不含前端）
make build-only

# 构建前端
make build-frontend

# 完整构建（前后端）
make build

# 运行开发环境
make run

# 运行测试
make test

# 代码格式化
make fmt

# 清理构建产物
make clean
```

### 前端开发

```bash
cd web
npm install
npm run dev  # http://localhost:3000
```

## 🧪 测试项目

我们提供了测试项目目录，用于在不影响全局配置的情况下测试 Grape 功能。

### 快速测试

```bash
# 1. 启动 Grape
./bin/grape

# 2. 进入测试项目
cd test-projects/vue3-demo

# 3. 安装依赖（自动通过 Grape 代理）
npm install

# 4. 运行项目
npm run dev  # http://localhost:5173
```

### 测试项目说明

| 项目 | 说明 |
|------|------|
| [vue3-demo](./test-projects/vue3-demo/) | Vue 3 + Vite + TypeScript 测试项目 |

**配置方式：** 每个测试项目都有独立的 `.npmrc` 文件，仅在当前项目生效，不会影响全局或其他项目。

详见：[测试项目使用指南](./test-projects/README.md)

## 📚 文档

- [**使用指南**](docs/USAGE.md) - 包管理器配置和使用详解
- [**API 文档**](docs/API.md) - 完整的 API 参考
- [**部署指南**](docs/DEPLOYMENT.md) - 生产环境部署指南
- [**开发文档**](docs/DEVELOPMENT.md) - 开发者指南
- [**Webhook 文档**](docs/WEBHOOKS.md) - Webhook 配置和使用

## 🔒 安全建议

1. **修改 JWT 密钥**: 设置复杂的 `auth.jwt_secret`
2. **修改默认密码**: 首次启动后立即修改 admin 密码
3. **使用 HTTPS**: 生产环境建议配置反向代理 (nginx/caddy)
4. **限制网络访问**: 仅允许可信网络访问服务端口
5. **禁用自助注册**: 设置 `auth.allow_registration: false`

## 🆚 与 Verdaccio 对比

| 维度 | Grape | Verdaccio |
|------|-------|-----------|
| 技术栈 | Go | Node.js |
| 内存占用 | < 10MB | ~ 50MB |
| 部署方式 | 单一二进制 | npm install |
| 多上游路由 | ✅ 按 scope | ❌ 单一上游 |
| 数据持久化 | ✅ SQLite | ❌ 文件系统 |
| 权限模型 | JWT + 数据库 | 配置文件 ACL |
| Web UI | Vue 3 + Element Plus | 内置简单 UI |
| Prometheus 指标 | ✅ 内置 | ❌ 需插件 |

## 🗺️ 路线图

### v0.2.0 (计划中)

- [ ] RBAC 权限模型
- [ ] PostgreSQL 支持
- [ ] 操作审计日志
- [ ] 包作用域权限

### v0.3.0 (计划中)

- [ ] Redis 缓存
- [ ] S3/MinIO 存储
- [ ] 垃圾回收机制
- [ ] 性能优化

### v1.0.0 (计划中)

- [ ] LDAP/OIDC 集成
- [ ] 高可用集群
- [ ] Docker 镜像
- [ ] Helm Chart

## 🤝 贡献

欢迎贡献代码、文档或建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📄 许可证

[Apache License 2.0](LICENSE)

## 🙏 致谢

- [npm](https://www.npmjs.com) - JavaScript 包管理器
- [Verdaccio](https://verdaccio.org) - 灵感来源
- [Gin](https://gin-gonic.com) - Go Web 框架
- [Vue 3](https://vuejs.org) - 前端框架
- [Element Plus](https://element-plus.org) - UI 组件库

---

<p align="center">
  Made with ❤️ by the Grape Team<br>
  🍇 轻盈如风，功能如山
</p>
