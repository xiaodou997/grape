# 🤖 Grape Registry - AI 交接指南

**文档版本**: v1.0  
**最后更新**: 2026-02-27  
**目标读者**: 新加入的 AI 助手、人类开发者

---

## 📚 推荐阅读顺序

### 第一天：快速上手

```
1. .ai/ai-handoff.md (本文档)     # AI 交接指南
2. .ai/project-context.md          # 项目上下文
3. .ai/env-spec.md                 # 环境规范
4. README.md                       # 项目说明
```

### 第二天：深入理解

```
1. .ai/architecture.md             # 系统架构
2. .ai/coding-rules.md             # 编码规范
3. docs/API.md                     # API 文档
4. internal/server/server.go       # 服务器入口代码
```

### 第三天：实战开发

```
1. .ai/tech-debt.md                # 技术债清单
2. .ai/roadmap.md                  # 产品路线图
3. internal/registry/proxy.go      # Registry 核心
4. internal/auth/middleware.go     # 认证中间件
```

---

## 🎯 核心设计原则

### 1. 单一二进制原则

**设计目标**: 一个二进制文件，无需额外运行时。

**实现方式**:
- Go 语言编译为静态二进制
- 前端嵌入（`embed` 指令）
- SQLite 嵌入式数据库

**AI 注意事项**:
```go
// ✅ 正确：使用 embed 嵌入前端
//go:embed dist/*
var webDist embed.FS

// ❌ 错误：依赖外部文件
http.ServeFile("/app", "/path/to/frontend")
```

### 2. 双端口分离原则

**设计目标**: Web UI 与 Registry API 独立监听，职责分离。

**端口分配**:
- **4873**: Web UI + 管理 API
- **4874**: npm Registry API

**AI 注意事项**:
```go
// Web UI 路由器
router := gin.New()
router.Use(securityHeadersMiddleware())  // Web 安全头

// Registry API 路由器
apiRouter := gin.New()
// 无安全头，npm 客户端不需要
```

### 3. 零配置启动原则

**设计目标**: 默认配置即可使用。

**实现方式**:
- 配置文件可选
- 合理默认值
- 环境变量覆盖

**AI 注意事项**:
```go
// ✅ 正确：提供默认值
port := cfg.Server.Port
if port == 0 {
    port = 4873  // 默认端口
}

// ❌ 错误：强制要求配置
if cfg.Server.Port == "" {
    return errors.New("port required")
}
```

### 4. 数据持久化原则

**设计目标**: 重启不丢失数据。

**实现方式**:
- SQLite 存储元数据
- 文件系统存储包文件
- 原子写入避免损坏

**AI 注意事项**:
```go
// ✅ 正确：原子写入
tmpPath := path + ".tmp"
os.WriteFile(tmpPath, data, 0644)
os.Rename(tmpPath, path)  // 原子操作

// ❌ 错误：直接写入
os.WriteFile(path, data, 0644)  // 可能损坏
```

### 5. npm 兼容原则

**设计目标**: 完整兼容 npm/yarn/pnpm/bun。

**实现方式**:
- 遵循 npm registry 协议
- 正确响应格式
- 支持所有客户端

**AI 注意事项**:
```go
// ✅ 正确：npm 兼容响应
c.JSON(http.StatusOK, gin.H{
    "_id": packageName,
    "_rev": rev,
    "dist-tags": map[string]string{
        "latest": latestVersion,
    },
    "versions": versions,
})

// ❌ 错误：自定义格式
c.JSON(http.StatusOK, gin.H{
    "name": packageName,  // npm 客户端不识别
})
```

---

## 🚫 不要做的事情

### 1. 不要修改核心协议

**禁止**:
- 修改 npm registry API 响应格式
- 移除必需字段（`_id`, `_rev`, `dist-tags`）
- 更改认证流程（除非兼容）

**原因**: 会破坏 npm 客户端兼容性。

### 2. 不要引入重型依赖

**禁止**:
- 引入完整的 Web 框架（已有 Gin）
- 引入其他 ORM（已有 GORM）
- 引入其他日志库（已有 Zap）

**原因**: 违背轻量级设计原则。

### 3. 不要忽略错误

**禁止**:
```go
// ❌ 禁止：忽略错误
doSomething()

// ❌ 禁止：panic 滥用
if err != nil {
    panic(err)
}
```

**正确做法**:
```go
// ✅ 正确：显式返回
if err != nil {
    return fmt.Errorf("failed to do: %w", err)
}
```

### 4. 不要硬编码配置

**禁止**:
```go
// ❌ 禁止：硬编码
port := "4873"
jwtSecret := "default-secret"
```

**正确做法**:
```go
// ✅ 正确：从配置读取
port := cfg.Server.Port
jwtSecret := cfg.Auth.JWTSecret
```

### 5. 不要破坏双端口设计

**禁止**:
- 在 4873 端口处理 Registry API
- 在 4874 端口返回 HTML

**原因**: 职责分离是核心架构。

---

## 📜 隐含业务规则

### 1. 包名验证规则

**规则**:
- 包名不能以 `.` 或 `_` 开头
- 包名不能包含大写字母
- scoped 包格式：`@scope/name`

**实现**:
```go
func validatePackageName(name string) error {
    if name[0] == '.' || name[0] == '_' {
        return errors.New("name cannot start with . or _")
    }
    if strings.ToUpper(name) != name {
        return errors.New("name cannot contain uppercase")
    }
    // ...
}
```

### 2. Token 权限规则

**规则**:
- JWT Token: 完整权限（24h 有效期）
- Persistent Token: 可设置只读权限
- 只读 Token 不能发布包

**实现**:
```go
// 检查 Token 权限
if token.Readonly && c.Request.Method == http.MethodPut {
    c.JSON(http.StatusForbidden, gin.H{
        "error": "readonly token cannot publish",
    })
    c.Abort()
    return
}
```

### 3. 包 Owner 规则

**规则**:
- 发布者自动成为 owner
- 非 owner 不能发布已有包
- admin 可以管理所有包

**实现**:
```go
// 检查包权限
var owner PackageOwner
err := db.DB.Where("package_name = ? AND user_id = ?", packageName, userID).First(&owner).Error
if err != nil && !user.IsAdmin {
    return errors.New("not owner of this package")
}
```

### 4. 上游路由规则

**规则**:
- 空 scope → 默认上游（npmjs）
- `@company/*` → company-private 上游
- `@internal/*` → internal-tools 上游

**实现**:
```go
func (p *Proxy) selectUpstream(packageName string) *Upstream {
    if strings.HasPrefix(packageName, "@company/") {
        return p.upstreams["company-private"]
    }
    if strings.HasPrefix(packageName, "@internal/") {
        return p.upstreams["internal-tools"]
    }
    return p.upstreams["npmjs"]  // 默认
}
```

### 5. 审计日志规则

**规则**:
- 登录操作必须记录
- 发布/删除操作必须记录
- 配置修改必须记录

**实现**:
```go
func (h *AuthHandler) Login(c *gin.Context) {
    // ... 登录逻辑
    
    // 记录审计日志
    db.Create(&AuditLog{
        Action: "login",
        Username: username,
        IP: c.ClientIP(),
        Detail: "User logged in",
    })
}
```

---

## 🗺️ 当前系统边界

### 已实现边界

```
┌─────────────────────────────────────────────────────────┐
│                    Grape Registry                        │
│                                                          │
│  ✅ 用户认证 (JWT + Token)                               │
│  ✅ 包发布/下载/删除                                     │
│  ✅ 多上游代理                                           │
│  ✅ 包 Owner 管理                                         │
│  ✅ 审计日志                                             │
│  ✅ Webhook 通知                                          │
│  ✅ Prometheus 监控                                       │
│  ✅ Web UI 管理后台                                       │
│  ✅ 备份/恢复                                            │
│  ✅ 垃圾回收                                             │
│                                                          │
│  ⚠️  双端口路由 (Bug 待修复)                              │
│  ❌ LDAP/AD 集成                                         │
│  ❌ S3/OSS 存储                                          │
│  ❌ Redis 缓存                                           │
│  ❌ PostgreSQL 支持                                      │
│  ❌ 高可用集群                                           │
└─────────────────────────────────────────────────────────┘
```

### 系统交互

```
Grape Registry
    │
    ├── npm/yarn/pnpm/bun 客户端 (Port 4874)
    │   └── 协议：npm registry API
    │
    ├── 浏览器用户 (Port 4873)
    │   └── 协议：HTTP + JSON
    │
    ├── 上游 Registry
    │   ├── registry.npmjs.org
    │   ├── registry.npmmirror.com
    │   └── 私有 Registry
    │
    └── 外部系统
        ├── Prometheus (指标收集)
        └── Webhook 接收方 (事件通知)
```

---

## 🔮 未来演进方向

### 短期演进（1-3 月）

**方向**: 稳定性与性能

1. **修复双端口路由 Bug**
   - 当前阻塞问题
   - 优先级：P0

2. **补充测试覆盖**
   - 单元测试 > 70%
   - 集成测试核心流程

3. **性能优化**
   - 大型包处理
   - 并发能力提升

### 中期演进（3-6 月）

**方向**: 企业特性

1. **LDAP/AD 集成**
   - 企业账号系统对接

2. **S3/OSS 存储**
   - 云原生部署支持

3. **PostgreSQL 支持**
   - 高并发场景

### 长期演进（6-12 月）

**方向**: 生态建设

1. **插件系统**
   - 可扩展架构

2. **高可用集群**
   - 多实例部署

3. **Kubernetes Operator**
   - K8s 原生支持

---

## 🔍 推荐分析流程

### 接到新任务时

```
1. 理解需求
   ↓
2. 查看相关文档
   - .ai/project-context.md
   - .ai/architecture.md
   ↓
3. 定位相关代码
   - grep 搜索关键词
   - 查看对应 handler
   ↓
4. 分析现有实现
   - 阅读代码
   - 运行测试
   ↓
5. 设计方案
   - 是否符合核心原则
   - 是否引入技术债
   ↓
6. 实现 + 测试
   - 遵循编码规范
   - 补充测试覆盖
   ↓
7. 更新文档
   - 代码注释
   - 相关文档
```

### 遇到 Bug 时

```
1. 复现问题
   - 记录复现步骤
   - 收集日志
   ↓
2. 定位原因
   - 查看错误堆栈
   - grep 搜索相关代码
   ↓
3. 分析根因
   - 代码逻辑问题
   - 配置问题
   - 环境问题
   ↓
4. 设计修复
   - 最小改动原则
   - 不引入新 Bug
   ↓
5. 验证修复
   - 本地测试
   - 回归测试
   ↓
6. 记录 Bug
   - 更新 BUG-*.md
   - 更新 tech-debt.md
```

### 添加新功能时

```
1. 需求分析
   - 是否核心功能
   - 是否违背设计原则
   ↓
2. 设计 API
   - RESTful 风格
   - 统一响应格式
   ↓
3. 设计数据模型
   - 数据库表结构
   - GORM 模型定义
   ↓
4. 实现 Handler
   - 认证检查
   - 参数验证
   - 业务逻辑
   - 响应返回
   ↓
5. 实现 Service
   - 核心业务逻辑
   - 数据访问
   ↓
6. 编写测试
   - 单元测试
   - 集成测试
   ↓
7. 更新文档
   - API 文档
   - 使用文档
```

---

## ⚠️ 风险提醒

### 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **双端口路由 Bug** | 服务无法启动 | 优先修复或回退单端口 |
| **SQLite 并发限制** | 高并发性能下降 | 短期 WAL 模式，长期 PG 支持 |
| **测试覆盖率低** | 回归风险高 | 逐步补充核心测试 |
| **大型包截断** | 元数据不完整 | 已提升到 50MB，需监控 |

### 业务风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **npm 协议变更** | 兼容性问题 | 关注 npm 官方动态 |
| **安全漏洞** | 数据泄露 | 定期安全审计 |
| **性能瓶颈** | 用户体验下降 | 持续性能测试 |

### 运维风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **数据丢失** | 包文件丢失 | 定期备份 |
| **单点故障** | 服务不可用 | 高可用方案规划 |
| **监控缺失** | 问题发现延迟 | Prometheus 告警配置 |

---

## 📞 获取帮助

### 文档资源

- **项目文档**: `docs/` 目录
- **AI 交接文档**: `.ai/` 目录
- **Bug 报告**: `BUG-*.md` 文件
- **迭代计划**: `迭代计划.md`

### 代码资源

- **入口文件**: `cmd/grape/main.go`
- **服务器**: `internal/server/server.go`
- **Registry 核心**: `internal/registry/proxy.go`
- **认证中间件**: `internal/auth/middleware.go`

### 调试技巧

```bash
# 启用 debug 日志
export GRAPE_LOG_LEVEL=debug
./bin/grape

# 查看实时日志
tail -f /tmp/grape.log

# 性能分析
go tool pprof http://localhost:4873/debug/pprof/heap
```

---

## 🎓 学习资源

### Go 语言

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Gin 框架

- [Gin 官方文档](https://gin-gonic.com/docs/)

### GORM

- [GORM 官方文档](https://gorm.io/docs/)

### npm Registry 协议

- [npm Registry Spec](https://github.com/npm/registry/blob/master/docs/README.md)

---

**最后更新**: 2026-02-27  
**维护者**: Grape Team  
**联系方式**: graperegistry@github
