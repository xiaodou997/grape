# Grape 配置指南

本文档详细介绍 Grape 的所有配置项和使用方法。

## 目录

- [配置文件位置](#配置文件位置)
- [完整配置示例](#完整配置示例)
- [配置项详解](#配置项详解)
- [环境变量](#环境变量)
- [配置最佳实践](#配置最佳实践)

---

## 配置文件位置

### 方式一：命令行指定

```bash
# 使用 -c 或 --config 参数
./grape -c /path/to/config.yaml
./grape --config /path/to/config.yaml
```

### 方式二：默认位置

如果不指定配置文件，Grape 会按以下顺序查找：

1. `./config.yaml` (当前目录)
2. `./configs/config.yaml`
3. `/etc/grape/config.yaml` (系统目录)

### 方式三：环境变量

部分配置可以通过环境变量覆盖（详见 [环境变量](#环境变量) 节）。

---

## 完整配置示例

```yaml
# ============================================
# Grape 配置文件
# ============================================

# --------------------------------------------
# 1. 服务器配置
# --------------------------------------------
server:
  host: "0.0.0.0"              # 监听地址
  port: 4873                    # 监听端口
  read_timeout: 30s             # 请求读取超时
  write_timeout: 30s            # 响应写入超时

# --------------------------------------------
# 2. npm Registry 配置
# --------------------------------------------
registry:
  # 向后兼容的单一上游配置（不推荐）
  # upstream: "https://registry.npmjs.org"

  # 多上游配置（推荐）
  upstreams:
    # 默认上游（公共包）
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""                 # 空字符串表示默认上游
      timeout: 30s              # 请求超时
      enabled: true             # 是否启用

    # 淘宝镜像（可选，加速访问）
    - name: "npmmirror"
      url: "https://registry.npmmirror.com"
      scope: ""
      timeout: 15s
      enabled: false

    # 公司私有包
    - name: "company-private"
      url: "https://npm.company.com"
      scope: "@company"         # 匹配 @company/* 包
      timeout: 30s
      enabled: true

    # 内部工具包
    - name: "internal-tools"
      url: "https://npm-internal.company.com"
      scope: "@internal"
      timeout: 30s
      enabled: true

# --------------------------------------------
# 3. 存储配置
# --------------------------------------------
storage:
  type: "local"                 # 存储类型：local
  path: "./data"                # 数据存储目录

# --------------------------------------------
# 4. 日志配置
# --------------------------------------------
log:
  level: "info"                 # 日志级别：debug, info, warn, error

# --------------------------------------------
# 5. 认证配置
# --------------------------------------------
auth:
  jwt_secret: "your-secret-key-change-in-production"  # JWT 签名密钥
  jwt_expiry: 24h               # Token 有效期
  allow_registration: false     # 是否允许自助注册

# --------------------------------------------
# 6. 数据库配置
# --------------------------------------------
database:
  type: "sqlite"                # 数据库类型：sqlite (目前仅支持 SQLite)
  dsn: "./data/grape.db"        # 数据库连接字符串
```

---

## 配置项详解

### 1. 服务器配置 (server)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `host` | string | `0.0.0.0` | 否 | 监听地址。生产环境建议设置为 `0.0.0.0` 以允许外部访问 |
| `port` | int | `4873` | 否 | 监听端口。确保防火墙开放此端口 |
| `read_timeout` | duration | `30s` | 否 | 请求读取超时时间。大型包上传可能需要更长时间 |
| `write_timeout` | duration | `30s` | 否 | 响应写入超时时间 |

**示例：**

```yaml
server:
  host: "0.0.0.0"
  port: 4873
  read_timeout: 60s      # 大型包上传建议设置更长
  write_timeout: 60s
```

### 2. Registry 配置 (registry)

#### 2.1 单一上游（向后兼容）

```yaml
registry:
  upstream: "https://registry.npmjs.org"
```

#### 2.2 多上游配置（推荐）

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `name` | string | - | 是 | 上游名称（标识符） |
| `url` | string | - | 是 | 上游仓库地址 |
| `scope` | string | `""` | 否 | 匹配的 scope，如 `@company`。空字符串表示默认上游 |
| `timeout` | duration | `30s` | 否 | 请求超时时间 |
| `enabled` | bool | `true` | 否 | 是否启用 |

**scope 路由规则：**

- `scope` 为空：作为默认上游，处理所有未匹配 scope 的包
- `scope` 为 `@company`：处理所有 `@company/*` 包
- 多个上游配置相同 scope：使用第一个匹配的上游

**示例：**

```yaml
registry:
  upstreams:
    # 默认上游
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""
      enabled: true

    # 公司私有包
    - name: "company"
      url: "https://npm.company.com"
      scope: "@company"
      enabled: true

    # 合作伙伴包
    - name: "partner"
      url: "https://npm.partner.com"
      scope: "@partner"
      enabled: true
```

### 3. 存储配置 (storage)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `type` | string | `local` | 否 | 存储类型。目前仅支持 `local` |
| `path` | string | `./data` | 否 | 数据存储目录。建议使用绝对路径 |

**存储目录结构：**

```
data/
├── grape.db                    # SQLite 数据库
└── packages/
    ├── lodash/
    │   ├── metadata.json       # 包元数据
    │   └── tarballs/
    │       └── lodash-4.17.21.tgz
    └── @babel/
        └── core/
            ├── metadata.json
            └── tarballs/
                └── core-7.23.0.tgz
```

**示例：**

```yaml
storage:
  type: "local"
  path: "/var/lib/grape/data"   # 生产环境建议使用绝对路径
```

### 4. 日志配置 (log)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `level` | string | `info` | 否 | 日志级别 |

**日志级别说明：**

- `debug`: 调试级别，输出所有日志（开发环境）
- `info`: 信息级别，输出常规日志（生产环境推荐）
- `warn`: 警告级别，仅输出警告和错误
- `error`: 错误级别，仅输出错误

**示例：**

```yaml
log:
  level: "debug"    # 开发环境
```

```yaml
log:
  level: "info"     # 生产环境
```

### 5. 认证配置 (auth)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `jwt_secret` | string | - | 是 | JWT 签名密钥。**生产环境必须修改** |
| `jwt_expiry` | duration | `24h` | 否 | Token 有效期 |
| `allow_registration` | bool | `false` | 否 | 是否允许自助注册 |

**安全建议：**

1. **JWT 密钥**: 使用强随机字符串（至少 32 字符）
   ```bash
   # 生成随机密钥
   openssl rand -base64 32
   ```

2. **Token 有效期**: 根据安全策略调整
   - 开发环境：`24h` 或更长
   - 生产环境：`1h` - `8h`

3. **自助注册**: 
   - 内部团队：可设置为 `true`
   - 生产环境：建议 `false`，由管理员创建用户

**示例：**

```yaml
auth:
  jwt_secret: "x7K#mP9$vL2@nQ5&wR8*tY3^uI6%oA1"  # 强随机密钥
  jwt_expiry: 8h                               # 8 小时有效期
  allow_registration: false                    # 禁用自助注册
```

### 6. 数据库配置 (database)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
|--------|------|--------|------|------|
| `type` | string | `sqlite` | 否 | 数据库类型。目前仅支持 `sqlite` |
| `dsn` | string | `./data/grape.db` | 否 | 数据库连接字符串 |

**示例：**

```yaml
database:
  type: "sqlite"
  dsn: "/var/lib/grape/data/grape.db"  # 生产环境建议使用绝对路径
```

---

## 环境变量

Grape 支持通过环境变量覆盖配置文件中的设置：

| 环境变量 | 对应配置 | 示例 |
|----------|----------|------|
| `GRAPE_SERVER_HOST` | `server.host` | `export GRAPE_SERVER_HOST=0.0.0.0` |
| `GRAPE_SERVER_PORT` | `server.port` | `export GRAPE_SERVER_PORT=8080` |
| `GRAPE_LOG_LEVEL` | `log.level` | `export GRAPE_LOG_LEVEL=debug` |
| `GRAPE_AUTH_JWT_SECRET` | `auth.jwt_secret` | `export GRAPE_AUTH_JWT_SECRET=secret` |
| `GRAPE_STORAGE_PATH` | `storage.path` | `export GRAPE_STORAGE_PATH=/data` |
| `GRAPE_DATABASE_DSN` | `database.dsn` | `export GRAPE_DATABASE_DSN=/data/grape.db` |

**优先级：** 环境变量 > 配置文件 > 默认值

---

## 配置最佳实践

### 开发环境配置

```yaml
# config.dev.yaml
server:
  host: "localhost"
  port: 4873
  read_timeout: 60s
  write_timeout: 60s

log:
  level: "debug"

auth:
  jwt_secret: "dev-secret-key-not-for-production"
  jwt_expiry: 24h
  allow_registration: true

storage:
  path: "./data"

database:
  dsn: "./data/grape.db"
```

### 生产环境配置

```yaml
# config.prod.yaml
server:
  host: "0.0.0.0"
  port: 4873
  read_timeout: 120s      # 大型包上传需要更长时间
  write_timeout: 120s

log:
  level: "info"           # 生产环境不建议使用 debug

auth:
  jwt_secret: "${JWT_SECRET}"  # 从环境变量读取
  jwt_expiry: 8h
  allow_registration: false    # 禁用自助注册

storage:
  path: "/var/lib/grape/data"  # 使用绝对路径

database:
  dsn: "/var/lib/grape/data/grape.db"
```

### Docker 环境配置

```yaml
# config.docker.yaml
server:
  host: "0.0.0.0"
  port: 4873

log:
  level: "info"

auth:
  jwt_secret: "${GRAPE_AUTH_JWT_SECRET}"
  jwt_expiry: 8h
  allow_registration: false

storage:
  path: "/data"

database:
  dsn: "/data/grape.db"
```

配合 Docker Compose 使用：

```yaml
# docker-compose.yml
version: '3.8'

services:
  grape:
    image: graperegistry/grape:latest
    ports:
      - "4873:4873"
    volumes:
      - grape-data:/data
      - ./config.docker.yaml:/etc/grape/config.yaml:ro
    environment:
      - GRAPE_AUTH_JWT_SECRET=${GRAPE_AUTH_JWT_SECRET}
    command: -c /etc/grape/config.yaml
    restart: unless-stopped

volumes:
  grape-data:
```

### 安全加固配置

```yaml
# config.secure.yaml
server:
  host: "127.0.0.1"       # 仅监听本地，通过 nginx 反向代理
  port: 4873
  read_timeout: 120s
  write_timeout: 120s

log:
  level: "warn"           # 仅记录警告和错误

auth:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiry: 4h          # 缩短 token 有效期
  allow_registration: false

storage:
  path: "/var/lib/grape/data"

database:
  dsn: "/var/lib/grape/data/grape.db"
```

配合 nginx 反向代理：

```nginx
# /etc/nginx/conf.d/grape.conf
server {
    listen 80;
    server_name npm.example.com;

    # 强制 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name npm.example.com;

    ssl_certificate /etc/ssl/certs/npm.example.com.crt;
    ssl_certificate_key /etc/ssl/private/npm.example.com.key;

    location / {
        proxy_pass http://127.0.0.1:4873;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 限制请求体大小（可选）
    client_max_body_size 50M;
}
```

---

## 配置检查

启动时，Grape 会自动检查配置的有效性：

```bash
# 启动服务
./grape -c config.yaml

# 如果配置有误，会输出错误信息
Failed to load config: invalid jwt_secret length
```

### 常见配置错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `invalid jwt_secret length` | JWT 密钥太短 | 使用至少 32 字符的密钥 |
| `port already in use` | 端口被占用 | 修改 `server.port` 或关闭占用程序 |
| `permission denied` | 数据目录无权限 | `chown -R grape:grape /var/lib/grape` |
| `invalid upstream url` | 上游 URL 格式错误 | 检查 URL 是否完整（包含协议） |

---

## 配置热重载

⚠️ **当前版本不支持配置热重载**

修改配置后需要重启服务：

```bash
# 优雅重启
systemctl restart grape

# 或手动重启
pkill grape
./grape -c config.yaml
```

---

## 获取帮助

如遇配置问题，请查看：

- [完整文档](../docs/)
- [常见问题](../docs/FAQ.md)
- [GitHub Issues](https://github.com/graperegistry/grape/issues)
