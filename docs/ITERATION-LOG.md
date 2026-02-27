# 迭代开发记录

**版本**: v0.2.0
**日期**: 2026-02-27
**目标**: 让一个研发团队可以放心地在生产环境使用 Grape 作为私有 npm 仓库

---

## 概述

本次迭代实现了 Phase 1 的三个核心功能，使 Grape 满足团队生产使用标准：

1. **CI/CD Token 管理** - 支持流水线自动发布
2. **数据备份与恢复** - 保障数据安全
3. **包级别访问控制** - 防止包被意外覆盖

---

## T1: CI/CD Token 管理

### 背景

当前 JWT 是每次 `npm login` 颁发的短期 token，CI 环境不能交互式登录。Verdaccio 支持 `npm token create` 生成持久化 token，这是 CI/CD 场景的刚需。

### 实现方案

#### 1. 数据库模型

新增 `tokens` 表 (`internal/db/models.go`):

```go
type Token struct {
    ID        uint       `gorm:"primaryKey" json:"id"`
    UserID    uint       `gorm:"index;not null" json:"userId"`
    User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Name      string     `gorm:"size:100;not null" json:"name"`         // 描述，如 "github-ci"
    TokenHash string     `gorm:"size:64;uniqueIndex;not null" json:"-"` // sha256 哈希存储
    Readonly  bool       `gorm:"default:false" json:"readonly"`         // 只读 token 不能发布
    ExpiresAt *time.Time `json:"expiresAt,omitempty"`                   // nil = 永不过期
    LastUsed  *time.Time `json:"lastUsed,omitempty"`
    CreatedAt time.Time  `json:"createdAt"`
}
```

**安全设计**:
- Token 原始值只在创建时返回一次，后续只存储 SHA256 哈希
- 支持设置过期时间和只读权限

#### 2. API 实现

新增 `internal/server/handler/token.go`:

```
POST   /-/npm/v1/tokens           创建 Token
GET    /-/npm/v1/tokens           列出当前用户的 Token
DELETE /-/npm/v1/tokens/token/:id 撤销 Token
```

**创建 Token 请求**:
```json
{
  "name": "github-actions",
  "readonly": false,
  "days": 30
}
```

**创建 Token 响应**:
```json
{
  "id": 1,
  "name": "github-actions",
  "readonly": false,
  "token": "e4b7132aaa7e6e37dcac0ba0d3cfd00e...",
  "expiresAt": "2026-03-29T07:56:11Z",
  "createdAt": "2026-02-27T07:56:11Z"
}
```

#### 3. 认证中间件改造

修改 `internal/auth/middleware.go`，支持双认证模式：

1. 优先检查 JWT Token（用户登录）
2. JWT 无效时检查持久化 Token（CI/CD）

```go
// TokenInfo Token 信息
type TokenInfo struct {
    ID       uint
    Name     string
    Readonly bool
}

// 认证流程
func AuthMiddleware(jwtService *JWTService, userStore UserStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 先尝试 JWT 验证
        claims, err := jwtService.ValidateToken(tokenString)
        if err == nil {
            // JWT 有效
            user, _ := userStore.Get(claims.Username)
            c.Set("user", user)
            c.Set("isTokenAuth", false)
            c.Next()
            return
        }

        // 2. JWT 无效，尝试 Token 验证
        user, tokenInfo, err := ValidateTokenByHash(tokenString)
        if err == nil {
            c.Set("user", user)
            c.Set("token", tokenInfo)
            c.Set("isTokenAuth", true)
        }
        c.Next()
    }
}
```

#### 4. 只读 Token 发布限制

修改 `internal/server/handler/publish.go`:

```go
func (h *PublishHandler) Publish(c *gin.Context) {
    // 检查是否使用只读 Token
    tokenInfo := auth.GetTokenInfo(c)
    if tokenInfo != nil && tokenInfo.Readonly {
        c.JSON(http.StatusForbidden, gin.H{"error": "read-only token cannot publish"})
        return
    }
    // ... 继续发布流程
}
```

#### 5. 前端 Token 管理页面

新增 `web/src/views/admin/Tokens.vue`:

- 列表展示所有 Token（名称、权限、过期时间、最后使用时间）
- 创建新 Token（支持设置名称、只读、有效期）
- 撤销 Token
- Token 创建后只显示一次，提示用户保存

### 使用方式

**命令行**:
```bash
# 方式1: 通过 API 创建（需要先登录获取 JWT）
curl -X POST "http://localhost:4873/-/npm/v1/tokens" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"github-ci","readonly":false,"days":30}'

# 方式2: Web UI 管理后台 → Token 管理
```

**GitHub Actions 示例**:
```yaml
- name: Publish to Grape
  env:
    GRAPE_TOKEN: ${{ secrets.GRAPE_TOKEN }}
  run: |
    npm config set //grape.company.com/:_authToken $GRAPE_TOKEN
    npm publish
```

---

## T2: 数据备份与恢复

### 背景

生产环境必须有数据保护。Grape 的数据就两部分：SQLite 文件 + 包文件目录。

### 实现方案

#### 1. 命令行工具

新增 `cmd/backup/main.go`，支持三个子命令：

```bash
# 创建备份
grape backup -o grape-backup-20260227.tar.gz

# 查看备份内容
grape list -i grape-backup-20260227.tar.gz

# 恢复备份
grape restore -i grape-backup-20260227.tar.gz --force
```

**备份内容**:
- `data/grape.db` - SQLite 数据库
- `data/packages/` - 所有包元数据和 tarball 文件
- `BACKUP-META` - 备份元信息（时间戳）

#### 2. Web UI 备份恢复

新增 `internal/server/handler/backup.go` 和 `web/src/views/admin/Backup.vue`:

```
GET  /-/api/admin/backup/info      # 获取备份信息（包数量、存储大小）
GET  /-/api/admin/backup/download  # 下载备份文件
POST /-/api/admin/backup/restore   # 上传恢复备份
GET  /-/api/admin/backup/list      # 列出自动备份记录
```

**Web UI 功能**:
- 数据概览（总包数、存储占用、数据库大小）
- 一键下载备份文件
- 上传备份文件恢复数据
- 使用说明（定时备份建议、命令行操作）

#### 3. 恢复保护机制

- 恢复前自动备份当前数据到 `.restore-backup-时间戳` 目录
- 需要 `--force` 参数或 Web UI 确认才能覆盖
- 恢复后提示重启服务

### 定时备份建议

```bash
# crontab -e
# 每天凌晨 3 点自动备份
0 3 * * * /path/to/grape backup -o /backup/grape-$(date +\%Y\%m\%d).tar.gz
```

---

## T3: 包级别访问控制

### 背景

当前任意 developer 都能发布/覆盖任何包。对于有多个小组共用一个 registry 的团队，需要控制"只有前端组能发 @ui/*"。

### 实现方案

#### 1. 数据库模型

新增 `package_owners` 表 (`internal/db/models.go`):

```go
type PackageOwner struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    PackageName string    `gorm:"uniqueIndex:idx_pkg_user;size:255;not null" json:"packageName"`
    Username    string    `gorm:"uniqueIndex:idx_pkg_user;size:100;not null" json:"username"`
    CreatedAt   time.Time `json:"createdAt"`
}
```

#### 2. 发布时自动设 Owner

修改 `internal/server/handler/publish.go`:

```go
func (h *PublishHandler) Publish(c *gin.Context) {
    // ... 认证检查 ...
    
    // 检查权限
    if h.storage.HasPackage(packageName) {
        // 已存在的包：检查是否为 owner 或 admin
        if !h.isPackageOwner(packageName, user.Username) && user.Role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "you are not a maintainer of this package",
            })
            return
        }
    } else {
        // 新包：发布后自动设置为 owner
        h.addPackageOwner(packageName, user.Username)
    }
    
    // ... 发布流程 ...
}
```

#### 3. Owner 管理 API

新增 `internal/server/handler/owner.go`:

```
# npm 协议兼容 API
GET    /-/package/:name/collaborators         列出 owner
PUT    /-/package/:name/collaborators/:user   添加 owner
DELETE /-/package/:name/collaborators/:user   移除 owner

# 管理后台 API（admin 专用）
GET    /-/api/admin/packages/:name/owners         列出 owner
POST   /-/api/admin/packages/:name/owners         添加 owner
DELETE /-/api/admin/packages/:name/owners/:user   移除 owner
```

#### 4. 前端 Owner 管理

修改 `web/src/views/PackageDetail.vue`，添加 Owner 管理卡片：

- 显示当前包的所有 Owner
- Admin 可以添加/移除 Owner
- Owner 可以添加其他 Owner

### 权限规则

| 角色 | 权限 |
|------|------|
| admin | 可以发布/删除任何包，管理所有 Owner |
| developer | 只能发布自己拥有的包（首次发布自动获得所有权） |
| readonly | 不能发布任何包 |

---

## 文件变更清单

### 新增文件

| 文件 | 说明 |
|------|------|
| `cmd/backup/main.go` | 备份恢复命令实现 |
| `internal/server/handler/token.go` | Token 管理 API |
| `internal/server/handler/owner.go` | Owner 管理 API |
| `internal/server/handler/backup.go` | Web 备份恢复 API |
| `web/src/views/admin/Tokens.vue` | Token 管理页面 |
| `web/src/views/admin/Backup.vue` | 备份恢复页面 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `internal/db/models.go` | 新增 Token、PackageOwner 模型 |
| `internal/auth/middleware.go` | 支持 Token 认证、TokenInfo |
| `internal/auth/user.go` | 新增 ID 字段、ValidatePassword、ValidateRole |
| `internal/auth/db_user.go` | 新增 ID 字段映射 |
| `internal/server/server.go` | 注册新路由 |
| `internal/server/handler/publish.go` | Owner 检查、只读 Token 检查 |
| `cmd/grape/main.go` | 子命令支持、新模型迁移 |
| `web/src/api/index.ts` | 新增 Token、Owner、Backup API |
| `web/src/router/index.ts` | 新增路由 |
| `web/src/views/Admin.vue` | 新增标签页 |
| `web/src/views/PackageDetail.vue` | Owner 管理界面 |

---

## 验收测试结果

### T1 Token 管理

| 测试项 | 结果 |
|--------|------|
| 创建 Token | ✅ 通过 |
| 列出 Token | ✅ 通过 |
| 撤销 Token | ✅ 通过 |
| Token 认证 | ✅ 通过 |
| 只读 Token 发布被拒绝 | ✅ 通过 |

### T2 备份恢复

| 测试项 | 结果 |
|--------|------|
| 命令行备份 | ✅ 通过 |
| 命令行列表 | ✅ 通过 |
| 命令行恢复 | ✅ 通过 |
| Web 下载备份 | ✅ 通过 |
| Web 备份信息 | ✅ 通过 |

### T3 包级别 ACL

| 测试项 | 结果 |
|--------|------|
| 新包发布自动设 owner | ✅ 通过 |
| 非 owner 发布被拒绝 | ✅ 通过 |
| Admin 可以管理所有包 | ✅ 通过 |
| Owner 可以添加其他 owner | ✅ 通过 |

---

## 后续计划

Phase 2 运维增强（按需推进）:

- [ ] LDAP/AD 集成
- [ ] Helm Chart
- [ ] 包清理 (GC)
- [ ] S3/OSS 存储
- [ ] Redis 元数据缓存
