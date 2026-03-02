# 🍇 Grape

> **轻盈如风的企业级私有 npm 仓库**  
> One binary, zero debt. 一个二进制，零负担。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![npm compatible](https://img.shields.io/badge/npm-compatible-brightgreen)](https://www.npmjs.com)
[![Docker](https://img.shields.io/badge/Docker-available-blue?logo=docker)](https://ghcr.io/xiaodou997/grape)

**[English](README.md)**

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
| 🌐 **现代 Web UI** | 极具冲击力的 Vineyard 设计系统 (毛玻璃, Inter 字体) |
| 🌍 **全国际化支持** | 开箱即用的中英文双语界面切换 |
| 🪶 **超轻量级** | 内存占用 < 10MB，二进制约 20MB |
| 🔔 **Webhook 通知** | 支持包发布/删除事件通知 |
| 📊 **Prometheus 指标** | 完整的监控指标支持 |
| 🎫 **CI/CD Token** | 支持自动化发布的专用令牌 |
| 💾 **备份恢复** | Web UI 和命令行支持 |
| 🔒 **包级权限** | 细粒度的访问控制 |

## 🚀 快速开始

### 方式一：下载预编译二进制

```bash
# Linux (x86_64)
curl -sL https://github.com/xiaodou997/grape/releases/latest/download/grape-linux-amd64.tar.gz | tar xz
./grape-linux-amd64

# Linux (ARM64)
curl -sL https://github.com/xiaodou997/grape/releases/latest/download/grape-linux-arm64.tar.gz | tar xz
./grape-linux-arm64

# 访问 Web UI
open http://localhost:4873
```

### 方式二：Docker 部署

```bash
# 拉取并运行
docker pull ghcr.io/xiaodou997/grape:latest
docker run -d --name grape -p 4873:4873 -p 4874:4874 ghcr.io/xiaodou997/grape:latest

# 查看日志
docker logs -f grape
```

### 方式三：从源码构建

```bash
git clone https://github.com/xiaodou997/grape.git
cd grape
make build
./bin/grape
```

## 🏗️ 架构设计

Grape 采用**双端口架构**：

| 端口 | 用途 |
|------|------|
| **4873** | Web UI + 管理 API |
| **4874** | npm Registry API |

这种分离确保：
- Web UI 使用标准端口访问
- npm API 完全兼容 registry 协议
- 可独立扩展

## 📖 使用方法

### 配置 npm

```bash
# 全局设置
npm set registry http://localhost:4874

# 或仅设置特定 scope（推荐）
npm set @mycompany:registry http://localhost:4874
```

### 安装包

```bash
# 公共包（自动从上游缓存）
npm install lodash
npm install express

# Scoped 包
npm install @mycompany/utils
```

### 发布私有包

```bash
# 登录（默认：admin / admin）
npm login --registry http://localhost:4874

# 发布
npm publish --registry http://localhost:4874
```

## 🌐 Web UI 功能

- 📊 **智能仪表盘** - 集成系统指标、实时状态与最近活动
- 👤 **现代用户管理** - 专为团队协作打造的控制面板
- 🎫 **Token 管理** - 安全的 CI/CD 令牌，支持一键复制
- 💾 **备份恢复** - 视觉对齐的操作卡片，保障数据安全
- 🗑️ **垃圾回收** - VS Code 风格的清理预览与自动化建议

**默认账户：** `admin` / `admin`

> ⚠️ 生产环境请立即修改默认密码！

## 📦 版本发布

| Tag 格式 | 示例 | 构建产物 |
|---------|------|---------|
| `v*` | `v1.0.0` | Linux 二进制 + Docker 镜像 |
| `beta-v*` | `beta-v1.0.0` | 仅 Linux 二进制（预发布版） |

```bash
# 正式版
git tag v1.0.0 && git push origin v1.0.0

# 测试版
git tag beta-v1.0.0 && git push origin beta-v1.0.0
```

## 🔧 配置说明

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  web_port: 4873      # Web UI 端口
  api_port: 4874      # npm API 端口

registry:
  upstreams:
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""
      timeout: 30s
      enabled: true

auth:
  jwt_secret: "your-secret-key-change-in-production"
  jwt_expiry: 24h

storage:
  type: "local"
  path: "./data"

database:
  type: "sqlite"
  dsn: "./data/grape.db"
```

## 💡 为什么选择 Grape

| 特性 | 说明 |
|------|------|
| 🚀 **单一二进制** | Go 语言编写，无运行时依赖 |
| 🪶 **轻量级** | 内存占用 < 10MB，极小资源消耗 |
| 🔀 **多上游路由** | 按 scope 路由到不同的 registry |
| 💾 **数据持久化** | SQLite 数据库，重启不丢失 |
| 🔐 **现代认证** | JWT 认证，数据库存储用户 |
| 🎫 **CI/CD Token** | 专用令牌支持自动化发布 |
| 🌐 **现代 Web UI** | Vue 3 + Element Plus 管理界面 |
| 📊 **Prometheus 支持** | 内置监控指标 |
| 💾 **备份恢复** | Web UI 或命令行导出/导入数据 |
| 🔒 **包级权限** | 细粒度的包访问控制 |
| 🗑️ **垃圾回收** | 自动清理旧版本包 |

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
│   ├── registry/           # npm registry 核心
│   ├── server/             # HTTP 服务 (Gin)
│   ├── storage/            # 存储抽象层
│   ├── webhook/            # Webhook 事件通知
│   └── web/                # 前端嵌入
├── web/                    # 前端源码 (Vue 3)
├── configs/                # 配置示例
└── docs/                   # 文档
```

## 🛠️ 开发

```bash
# 开发模式（前后端热重载）
# 前端: http://localhost:3000 (Vite)
# 后端: http://localhost:4873 (Air)
make dev

# 构建后端（不含前端）
make build-only

# 构建前端
make build-frontend

# 完整构建
make build

# 运行测试
make test
```

**开发模式特性：**
- 🔄 前端热重载 (Vite) - 修改 Vue/TS 文件即时更新
- 🔄 后端热重载 (Air) - 修改 Go 文件自动重新编译
- 🌐 开发时访问 `http://localhost:3000`
- 📡 API 请求自动代理到后端

## 📚 文档

- [**使用指南**](docs/USAGE.md) - 包管理器配置和使用详解
- [**API 文档**](docs/API.md) - 完整的 API 参考
- [**部署指南**](docs/DEPLOYMENT.md) - 生产环境部署指南
- [**开发文档**](docs/DEVELOPMENT.md) - 开发者指南

## 🔒 安全建议

1. **修改 JWT 密钥**: 设置复杂的 `auth.jwt_secret`
2. **修改默认密码**: 首次启动后立即修改 admin 密码
3. **使用 HTTPS**: 生产环境建议配置反向代理
4. **限制网络访问**: 仅允许可信网络访问服务端口
5. **禁用自助注册**: 设置 `auth.allow_registration: false`

## 🗺️ 路线图

### v0.2.0 (计划中)
- [ ] RBAC 权限模型
- [ ] PostgreSQL 支持
- [ ] 操作审计日志
- [ ] 包作用域权限

### v0.3.0 (计划中)
- [ ] Redis 缓存
- [ ] S3/MinIO 存储
- [ ] 性能优化

### v1.0.0 (计划中)
- [ ] LDAP/OIDC 集成
- [ ] 高可用集群
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
