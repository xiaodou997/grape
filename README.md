# 🍇 Grape

> 轻盈如风的企业级私有 npm 仓库  
> One binary, zero debt. 一个二进制，零负担。

## 特性

- 🚀 **单一二进制** - 无需 Node.js，下载即用
- 📦 **npm 兼容** - 完整支持 npm/yarn/pnpm/bun
- 🔐 **用户认证** - JWT 认证，支持发布私有包
- 💾 **智能缓存** - 自动缓存公共包，加速团队开发
- 🌐 **现代 Web UI** - Vue 3 + Element Plus 管理界面
- 🪶 **轻量级** - 内存占用 < 10MB

## 快速开始

### 下载运行

```bash
# 下载 (macOS/Linux)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-$(uname -s)-$(uname -m) -o grape
chmod +x grape

# 运行
./grape

# 访问
open http://localhost:4873
```

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/graperegistry/grape.git
cd grape

# 构建 (包含前端)
make build

# 运行
./bin/grape
```

## 使用方法

### 配置 npm

```bash
# 设置 registry
npm set registry http://localhost:4873

# 或者使用作用域
npm set @company:registry http://localhost:4873
```

### 安装包

```bash
npm install lodash
npm install @babel/core
```

### 发布私有包

```bash
# 登录
npm login --registry http://localhost:4873
# 用户名: admin
# 密码: admin

# 发布
npm publish --registry http://localhost:4873
```

### 恢复默认源

```bash
npm set registry https://registry.npmjs.org
```

## 配置

### 配置文件 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 4873
  read_timeout: 30s
  write_timeout: 30s

registry:
  upstream: "https://registry.npmjs.org"

storage:
  type: "local"
  path: "./data"

auth:
  jwt_secret: "your-secret-key"
  jwt_expiry: 24h

log:
  level: "info"
```

### 命令行参数

```bash
./grape -c /path/to/config.yaml
./grape -h  # 查看帮助
```

## API 端点

### npm Registry API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/:package` | GET | 获取包元数据 |
| `/:package/-/:filename` | GET | 下载 tarball |
| `/:package` | PUT | 发布包 |
| `/:package` | DELETE | 删除包 |
| `/-/user/:username` | PUT | 用户登录 |

### 管理 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/-/health` | GET | 健康检查 |
| `/-/api/packages` | GET | 包列表 |
| `/-/api/stats` | GET | 统计信息 |
| `/-/api/user` | GET | 当前用户信息 |

## 项目结构

```
grape/
├── cmd/grape/           # 程序入口
├── internal/
│   ├── auth/           # 用户认证
│   ├── config/         # 配置管理
│   ├── logger/         # 日志系统
│   ├── registry/       # npm 核心
│   ├── server/         # HTTP 服务
│   ├── storage/        # 存储
│   └── web/            # 前端嵌入
├── web/                # 前端源码
├── configs/            # 配置示例
├── docs/               # 文档
└── Makefile
```

## 开发

### 环境要求

- Go 1.21+
- Node.js 18+ (仅前端开发需要)

### 开发命令

```bash
# 构建后端
make build-only

# 构建前端
make build-frontend

# 完整构建
make build

# 运行开发环境
make dev

# 清理
make clean
```

## 路线图

### v0.1.0 (当前)
- ✅ npm 代理缓存
- ✅ 用户认证 (JWT)
- ✅ 包发布/删除
- ✅ Web 管理界面
- ✅ 单一二进制部署

### v0.2.0 (计划)
- 🔲 SQLite/PostgreSQL 支持
- 🔲 数据持久化

### v0.3.0 (计划)
- 🔲 RBAC 权限模型
- 🔲 审计日志

### v1.0.0 (计划)
- 🔲 LDAP/OIDC 集成
- 🔲 高可用集群
- 🔲 Redis 缓存
- 🔲 S3 存储

## 与 Verdaccio 对比

| 维度 | Grape | Verdaccio |
|------|-------|-----------|
| 技术栈 | Go | Node.js |
| 内存占用 | < 10MB | ~ 50MB |
| 部署方式 | 单一二进制 | npm install |
| 权限模型 | JWT + 内存 | 配置文件 |
| 数据库 | 文件系统 (计划: PG/SQLite) | 文件系统 |

## 贡献

欢迎贡献代码！请查看 [贡献指南](docs/CONTRIBUTING.md)。

## 许可证

[Apache 2.0](LICENSE)

---

<p align="center">
  Made with ❤️ by the Grape Team
</p>
