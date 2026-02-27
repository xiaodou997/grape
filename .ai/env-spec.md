# 🖥️ Grape Registry - 环境规范

**文档版本**: v1.0  
**最后更新**: 2026-02-27  
**目标读者**: 新开发者、DevOps 工程师、AI 助手

---

## 📋 环境要求

### 开发环境

| 组件 | 最低版本 | 推荐版本 | 用途 |
|------|----------|----------|------|
| **Go** | 1.21 | 1.25.0 | 后端开发 |
| **Node.js** | 18 | 20+ | 前端开发 |
| **npm** | 9 | 10+ | 前端依赖管理 |
| **Git** | 2.0 | 最新 | 版本控制 |
| **Make** | 3.8 | 4.0+ | 构建工具 |

### 运行环境

| 组件 | 最低要求 | 推荐配置 |
|------|----------|----------|
| **CPU** | 1 核 | 2 核+ |
| **内存** | 64MB | 128MB+ |
| **磁盘** | 500MB | 2GB+ (SSD) |
| **操作系统** | Linux/macOS/Windows | Linux (Alpine) |

### 生产环境

| 组件 | 要求 | 说明 |
|------|------|------|
| **CPU** | 2 核+ | 高并发场景 |
| **内存** | 256MB+ | 含缓存 |
| **磁盘** | 10GB+ | 包存储 |
| **数据库** | SQLite/PostgreSQL | 高并发推荐 PG |
| **存储** | 本地/S3 | 大规模推荐 S3 |

---

## 🔧 依赖管理

### Go 依赖

**管理工具**: Go Modules

```bash
# 下载依赖
go mod download

# 整理依赖（移除未使用）
go mod tidy

# 添加新依赖
go get github.com/gin-gonic/gin@v1.11.0

# 升级依赖
go get -u github.com/gin-gonic/gin

# 查看依赖
go list -m all
```

**依赖锁定**:
- `go.mod` - 直接依赖
- `go.sum` - 所有依赖的校验和

### Node.js 依赖

**管理工具**: npm

```bash
# 安装依赖
cd web && npm install

# 添加新依赖
npm install axios@^1.13.5

# 升级依赖
npm update

# 查看过时依赖
npm outdated

# 清理缓存
npm cache clean --force
```

**依赖锁定**:
- `web/package.json` - 直接依赖
- `web/package-lock.json` - 锁定版本

---

## 🚀 启动命令

### 开发环境

```bash
# 方式 1: 使用 Make（推荐）
make dev
# 后端：http://localhost:4873
# 前端：http://localhost:3000

# 方式 2: 手动启动
# 后端
go run ./cmd/grape

# 前端
cd web && npm run dev

# 方式 3: 使用配置文件
go run ./cmd/grape -c ./configs/config.yaml
```

### 生产环境

```bash
# 方式 1: 二进制运行
./bin/grape -c ./configs/config.yaml

# 方式 2: Docker 运行
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  -v grape-data:/data \
  graperegistry/grape:latest

# 方式 3: Docker Compose
docker-compose up -d
```

### 测试命令

```bash
# 单元测试
go test -v ./...

# 带覆盖率
go test -v -cover ./...

# 特定包测试
go test -v ./internal/auth/

# 前端测试
cd web && npm test
```

---

## 🌍 环境变量

### 核心环境变量

| 变量名 | 默认值 | 说明 | 示例 |
|--------|--------|------|------|
| `GRAPE_CONFIG` | `./configs/config.yaml` | 配置文件路径 | `/etc/grape/config.yaml` |
| `GRAPE_HOST` | `0.0.0.0` | 监听地址 | `127.0.0.1` |
| `GRAPE_PORT` | `4873` | Web UI 端口 | `8080` |
| `GRAPE_API_PORT` | `4874` | Registry API 端口 | `8081` |
| `GRAPE_JWT_SECRET` | (必须设置) | JWT 密钥 | `your-secret-key` |
| `GRAPE_LOG_LEVEL` | `info` | 日志级别 | `debug` |
| `GRAPE_STORAGE_PATH` | `./data` | 数据存储路径 | `/var/lib/grape` |
| `GRAPE_DATABASE_DSN` | `./data/grape.db` | 数据库连接 | `/var/lib/grape/grape.db` |

### 前端环境变量

| 变量名 | 默认值 | 说明 | 示例 |
|--------|--------|------|------|
| `VITE_API_URL` | (空，同源) | API 基础 URL | `http://localhost:4874` |
| `VITE_API_PORT` | `4874` | API 端口 | `8081` |

### 使用方式

**方式 1: 命令行设置**
```bash
export GRAPE_JWT_SECRET="your-secret-key"
export GRAPE_LOG_LEVEL="debug"
./bin/grape
```

**方式 2: .env 文件**
```bash
# .env
GRAPE_JWT_SECRET=your-secret-key
GRAPE_LOG_LEVEL=debug
GRAPE_STORAGE_PATH=/var/lib/grape
```

**方式 3: Docker 环境变量**
```yaml
# docker-compose.yml
services:
  grape:
    environment:
      - GRAPE_JWT_SECRET=your-secret-key
      - GRAPE_LOG_LEVEL=info
```

---

## 🔌 端口说明

### 默认端口

| 端口 | 用途 | 协议 | 说明 |
|------|------|------|------|
| **4873** | Web UI + 管理 API | HTTP | 浏览器访问、管理后台 |
| **4874** | npm Registry API | HTTP | npm/yarn/pnpm 客户端 |

### 端口配置

```yaml
# config.yaml
server:
  host: 0.0.0.0
  port: 4873              # Web UI 端口
  api_port: 4874          # Registry API 端口
  read_timeout: 30s
  write_timeout: 30s
```

### 端口冲突处理

```bash
# 检查端口占用
lsof -i :4873
lsof -i :4874

# 修改端口
export GRAPE_PORT=8080
export GRAPE_API_PORT=8081
./bin/grape
```

---

## 🐳 Docker 使用

### 构建镜像

```bash
# 本地构建
docker build -t grape:latest .

# 多平台构建
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t grape:latest \
  --push .
```

### 运行容器

```bash
# 基本运行
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  -v grape-data:/data \
  grape:latest

# 带配置文件
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  -v ./configs/config.yaml:/app/configs/config.yaml \
  -v grape-data:/data \
  grape:latest \
  --config /app/configs/config.yaml

# 开发模式（热重载）
docker run -d \
  --name grape-dev \
  -p 4873:4873 \
  -p 4874:4874 \
  -v $(pwd):/app \
  -e GRAPE_LOG_LEVEL=debug \
  grape:latest
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  grape:
    image: graperegistry/grape:latest
    container_name: grape
    ports:
      - "4873:4873"
      - "4874:4874"
    volumes:
      - grape-data:/data
      - ./configs:/app/configs
    environment:
      - GRAPE_JWT_SECRET=your-secret-key
      - GRAPE_LOG_LEVEL=info
    restart: unless-stopped

volumes:
  grape-data:
```

**启动命令**:
```bash
docker-compose up -d

# 查看日志
docker-compose logs -f grape

# 停止
docker-compose down

# 重启
docker-compose restart
```

---

## 💾 数据库初始化

### SQLite（默认）

**自动初始化**:
```bash
# 首次启动自动创建数据库文件
./bin/grape

# 数据库文件位置
./data/grape.db
```

**手动初始化**:
```bash
# 创建数据目录
mkdir -p ./data

# 设置权限（Linux）
chown -R $(whoami):$(whoami) ./data
chmod -R 755 ./data
```

**验证**:
```bash
# 检查数据库文件
ls -la ./data/grape.db

# 使用 sqlite3 查看
sqlite3 ./data/grape.db ".tables"
```

### PostgreSQL（规划中）

**初始化脚本**（规划）:
```sql
-- 创建数据库
CREATE DATABASE grape;

-- 创建用户
CREATE USER grape_user WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE grape TO grape_user;

-- 运行迁移
./bin/grape migrate --database postgres
```

---

## 📁 目录结构

### 运行时目录

```
/data/                      # 数据根目录
├── grape.db               # SQLite 数据库
├── packages/              # 包文件存储
│   ├── lodash/
│   │   └── lodash-4.17.21.tgz
│   └── @babel/
│       └── core/
│           └── core-7.20.0.tgz
└── backups/               # 备份文件
    └── grape-20260227.tar.gz
```

### 配置目录

```
/configs/
├── config.yaml            # 主配置文件
└── config.prod.yaml       # 生产环境配置
```

### 日志目录

```
/logs/
├── grape.log              # 主日志文件
└── grape-error.log        # 错误日志
```

---

## 🔍 健康检查

### HTTP 健康检查

```bash
# Web UI 健康检查
curl http://localhost:4873/-/health
# 返回：{"status":"ok","time":"2026-02-27T00:00:00Z"}

# Registry API 健康检查
curl http://localhost:4874/-/health
# 返回：{"status":"ok","time":"2026-02-27T00:00:00Z"}
```

### 功能验证

```bash
# 验证包下载
curl http://localhost:4874/lodash

# 验证认证
curl -u admin:admin http://localhost:4873/-/api/user

# 验证监控
curl http://localhost:4873/-/metrics
```

---

## 🛠️ 故障排查

### 常见问题

**问题 1: 端口被占用**
```bash
# 错误：bind: address already in use
# 解决：检查并关闭占用进程
lsof -i :4873
kill -9 <PID>

# 或修改端口
export GRAPE_PORT=8080
```

**问题 2: 数据库锁定**
```bash
# 错误：database is locked
# 解决：检查是否有其他进程访问
lsof ./data/grape.db

# 或重启服务
./bin/grape
```

**问题 3: 权限不足**
```bash
# 错误：permission denied
# 解决：修改目录权限
chown -R $(whoami):$(whoami) ./data
chmod -R 755 ./data
```

**问题 4: 配置文件无效**
```bash
# 错误：failed to load config
# 解决：验证 YAML 语法
yamllint configs/config.yaml

# 或使用默认配置
./bin/grape
```

---

## 📊 性能调优

### Go 运行时

```bash
# 设置 GOMAXPROCS
export GOMAXPROCS=4

# 设置内存限制
export GOMEMLIMIT=512MiB
```

### SQLite 优化

```sql
-- 启用 WAL 模式（提升并发）
PRAGMA journal_mode = WAL;

-- 设置缓存大小
PRAGMA cache_size = -64000;  -- 64MB

-- 启用异步写入
PRAGMA synchronous = NORMAL;
```

### 系统级优化

```bash
# 增加文件描述符限制
ulimit -n 65536

# 增加 TCP 连接队列
sysctl -w net.core.somaxconn=65535
```

---

**最后更新**: 2026-02-27  
**下次审查**: 环境变更时更新
