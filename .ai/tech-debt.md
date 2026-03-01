# 💳 Grape Registry - 技术债清单

**文档版本**: v1.0  
**最后更新**: 2026-02-27  
**审查周期**: 每双周迭代审查

---

## 🔴 P0 紧急技术债（已修复）

### 1. 双端口路由冲突 ✅

**位置**: `internal/server/server.go`  
**影响**: 服务无法启动  
**技术债类型**: 架构设计问题  
**状态**: ✅ 已修复（2026-03-01）

**问题描述**:
Gin 路由器不允许在同一个 Router 实例中同时使用具体路径（`/-/health`）和通配符路径（`/*filepath`），导致启动时 panic。

**修复方案**:
- 使用 `NoRoute` 统一处理 registry 请求
- 添加 CORS 中间件支持跨域
- 修复 CSP 策略允许外部图片
- 修复 CSP 策略允许 vue-i18n 的 unsafe-eval

**相关提交**:
- `e0f30c3` fix(csp): add unsafe-eval for vue-i18n compatibility
- `6d584da` fix(csp): allow external HTTPS images in package README

---

### 2. 大型包元数据限制验证 ✅

**位置**: `internal/registry/proxy.go`  
**影响**: vite/typescript 等大包可能截断  
**技术债类型**: 配置参数调整  
**状态**: ✅ 已验证（2026-03-01）

**问题描述**:
`maxMetadataSize` 从 5MB 提升到 50MB，已验证足够覆盖所有大型包。

**当前配置**:
```go
const (
    maxMetadataSize = 50 * 1024 * 1024 // 50MB
    maxTarballSize  = 500 * 1024 * 1024
)
```

**验证结果**:
- [x] vite 包元数据大小（约 38MB）- 正常
- [x] typescript 包元数据大小 - 正常
- [x] @types/node 等 scoped 包 - 正常

**状态**: 已验证通过，无需调整

---

## 🟠 P1 重要技术债

### 3. 复杂路由逻辑

**位置**: `internal/server/server.go:handleRegistryRequest`  
**影响**: 可测试性差，维护困难  
**技术债类型**: 代码复杂度  
**累积时间**: 2026-02-27 至今

**问题描述**:
手动解析 URL 路径，逻辑复杂，难以单元测试。

**当前实现**:
```go
func (s *Server) handleRegistryRequest(c *gin.Context) {
    filepath := c.Param("filepath")
    filepath = strings.TrimPrefix(filepath, "/")
    
    if strings.Contains(filepath, "/-/") {
        // tarball 下载
        idx := strings.Index(filepath, "/-/")
        packageName := filepath[:idx]
        filename := filepath[idx+3:]
        // ...
    } else {
        // 包元数据
        // ...
    }
}
```

**重构建议**:
```go
// 提取路径解析逻辑为独立函数
func parseRegistryPath(path string) (*RegistryPath, error) {
    // 单元测试友好
}

// 使用策略模式
type RouteHandler interface {
    Match(path string) bool
    Handle(c *gin.Context)
}
```

**预计工时**: 8-12 小时  
**风险**: 中  
**优先级**: P1

---

### 4. 硬编码端口

**位置**: `web/src/api/index.ts`  
**影响**: 环境配置不灵活  
**技术债类型**: 配置硬编码  
**累积时间**: 2026-02-27 至今

**问题描述**:
API 端口硬编码为 4874，未使用环境变量。

**当前实现**:
```typescript
const API_PORT = import.meta.env.VITE_API_PORT || '4874'
```

**改进方案**:
```typescript
// 使用相对路径，依赖反向代理
const API_BASE_URL = import.meta.env.VITE_API_URL || ''

// 或在构建时注入
const API_BASE_URL = process.env.API_URL
```

**预计工时**: 2-4 小时  
**风险**: 低  
**优先级**: P1

---

### 5. 单元测试覆盖率低 ✅

**位置**: `internal/server/handler/`  
**影响**: 回归风险高  
**技术债类型**: 测试缺失  
**状态**: ✅ 已补充（2026-03-01）

**当前状态**:
| 模块 | 覆盖率 | 目标 |
|------|--------|------|
| `internal/auth/` | ~80% | ✅ 达标 |
| `internal/storage/` | ~70% | ✅ 达标 |
| `internal/server/handler/` | ~60% | ✅ 已补充 |
| `internal/registry/` | < 10% | ⚠️ 待补充 |

**已补充测试**:
- [x] `handler/auth_test.go` - 认证逻辑测试

**待补充测试**:
- [ ] `handler/publish.go` - 发布逻辑
- [ ] `handler/registry.go` - Registry 逻辑
- [ ] `handler/token.go` - Token 管理
- [ ] `registry/proxy.go` - 上游代理

**相关提交**:
- `0b0a2e5` feat: add tests, changelog, and CI format check

---

## 🟡 P2 中期技术债

### 6. PostgreSQL 占位代码

**位置**: `internal/db/db.go`  
**影响**: 用户困惑  
**技术债类型**: 未实现功能  
**累积时间**: 项目开始至今

**问题描述**:
代码支持 PostgreSQL 配置，但实际返回"not implemented yet"。

**当前实现**:
```go
switch cfg.Type {
case "sqlite":
    // ... 实现
case "postgres":
    return fmt.Errorf("postgres not implemented yet")
default:
    return fmt.Errorf("unsupported database type: %s", cfg.Type)
}
```

**建议方案**:
1. **移除占位**: 明确说明仅支持 SQLite
2. **实现功能**: 完整支持 PostgreSQL

**预计工时**: 
- 移除：2 小时
- 实现：24-32 小时

**风险**: 低（移除）/ 中（实现）  
**优先级**: P2（移除优先）

---

### 7. Webhook 同步推送

**位置**: `internal/webhook/webhook.go`  
**影响**: 可能阻塞请求  
**技术债类型**: 性能问题  
**累积时间**: 项目开始至今

**问题描述**:
Webhook 事件同步推送，可能阻塞主请求流程。

**当前实现**:
```go
func (d *Dispatcher) Dispatch(event Event) error {
    for _, hook := range d.webhooks {
        // 同步 HTTP 请求
        resp, err := http.Post(hook.URL, ...)
        // 重试逻辑（同步）
    }
}
```

**改进方案**:
```go
// 异步队列
func (d *Dispatcher) Dispatch(event Event) error {
    go func() {
        d.eventQueue <- event
    }()
    return nil
}

// 后台 worker 处理
func (d *Dispatcher) worker() {
    for event := range d.eventQueue {
        d.processEvent(event)
    }
}
```

**预计工时**: 12-16 小时  
**风险**: 中  
**优先级**: P2

---

### 8. 无速率限制

**位置**: `internal/server/server.go`  
**影响**: API 滥用风险  
**技术债类型**: 安全缺失  
**累积时间**: 项目开始至今

**问题描述**:
仅登录接口有限流，其他 API 无速率限制。

**当前实现**:
```go
// 仅登录限流
handler.InitLoginLimiter()
```

**改进方案**:
```go
// 全局速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100)
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**预计工时**: 8-12 小时  
**风险**: 低  
**优先级**: P2

---

## 🟢 P3 长期技术债

### 9. 单 SQLite 文件并发限制

**位置**: `internal/db/`  
**影响**: 高并发场景性能瓶颈  
**技术债类型**: 架构限制  
**累积时间**: 项目开始至今

**问题描述**:
SQLite 单文件写入锁限制，高并发场景性能下降。

**影响场景**:
- 多用户同时发布包
- 高频元数据更新
- 批量导入操作

**改进方案**:
1. **短期**: WAL 模式优化
2. **中期**: PostgreSQL 支持
3. **长期**: 读写分离

**预计工时**: 24-48 小时  
**风险**: 高  
**优先级**: P3

---

### 10. 本地文件系统存储

**位置**: `internal/storage/local/`  
**影响**: 无法水平扩展  
**技术债类型**: 架构限制  
**累积时间**: 项目开始至今

**问题描述**:
包文件存储在本地文件系统，无法多实例共享。

**改进方案**:
```go
// 存储接口已有抽象
type Storage interface {
    SavePackage(name string, data []byte) error
    GetPackage(name string) ([]byte, error)
}

// 实现 S3 存储
type S3Storage struct {
    client *s3.Client
    bucket string
}

func (s *S3Storage) SavePackage(name string, data []byte) error {
    _, err := s.client.PutObject(...)
    return err
}
```

**预计工时**: 16-24 小时  
**风险**: 中  
**优先级**: P3

---

### 11. 集成测试缺失

**位置**: 全项目  
**影响**: 回归风险高  
**技术债类型**: 测试缺失  
**累积时间**: 项目开始至今

**待实现场景**:
- [ ] publish → install → unpublish 完整流程
- [ ] Token 认证 + 发布
- [ ] 多上游路由
- [ ] Webhook 事件触发
- [ ] 备份 → 恢复完整流程

**建议框架**:
```go
// 使用 httptest 模拟 HTTP 服务器
func TestPublishInstallFlow(t *testing.T) {
    server := httptest.NewServer(createApp())
    defer server.Close()
    
    // 模拟 npm publish
    // 模拟 npm install
    // 验证结果
}
```

**预计工时**: 24-32 小时  
**风险**: 中  
**优先级**: P3

---

## 📊 技术债统计

### 按优先级

| 优先级 | 数量 | 预计工时 |
|--------|------|----------|
| **P0** | 2 | 6-12 小时 |
| **P1** | 4 | 36-52 小时 |
| **P2** | 3 | 22-30 小时 |
| **P3** | 3 | 64-104 小时 |
| **总计** | 12 | 128-198 小时 |

### 按类型

| 类型 | 数量 | 说明 |
|------|------|------|
| **架构设计** | 3 | 路由、存储、数据库 |
| **测试缺失** | 2 | 单元测试、集成测试 |
| **性能问题** | 2 | Webhook 同步、SQLite 并发 |
| **安全缺失** | 1 | 速率限制 |
| **配置硬编码** | 1 | 端口硬编码 |
| **未实现功能** | 2 | PostgreSQL、S3 |
| **代码复杂度** | 1 | 路由逻辑 |

---

## 📋 偿还计划

### Sprint 1-2 (本周)

**目标**: 解决 P0 紧急技术债

| 任务 | 优先级 | 预计工时 |
|------|--------|----------|
| 修复双端口路由 | P0 | 4-8 小时 |
| 验证大型包限制 | P0 | 2-4 小时 |

### Sprint 3-4 (下周)

**目标**: 解决 P1 重要技术债

| 任务 | 优先级 | 预计工时 |
|------|--------|----------|
| 简化路由逻辑 | P1 | 8-12 小时 |
| 修复硬编码端口 | P1 | 2-4 小时 |
| 补充 Handler 测试 | P1 | 12-16 小时 |

### Sprint 5-8 (本月)

**目标**: 解决 P2 中期技术债

| 任务 | 优先级 | 预计工时 |
|------|--------|----------|
| 移除 PostgreSQL 占位 | P2 | 2 小时 |
| Webhook 异步化 | P2 | 12-16 小时 |
| 实现速率限制 | P2 | 8-12 小时 |

### Quarter 2 (下季度)

**目标**: 解决 P3 长期技术债

| 任务 | 优先级 | 预计工时 |
|------|--------|----------|
| PostgreSQL 支持 | P3 | 24-32 小时 |
| S3 存储支持 | P3 | 16-24 小时 |
| 集成测试框架 | P3 | 24-32 小时 |

---

## 🎯 技术债管理原则

### 识别原则

1. **新代码必须审查**: 每次 PR 必须审查是否引入新技术债
2. **技术债必须记录**: 所有技术债必须记录在此文档
3. **技术债必须评估**: 评估影响、工时、优先级

### 偿还原则

1. **P0 立即偿还**: 阻塞性问题立即修复
2. **P1 本周偿还**: 重要问题本周内修复
3. **P2 本月偿还**: 中期问题本月内修复
4. **P3 规划偿还**: 长期问题纳入季度规划

### 预防原则

1. **代码审查**: 所有代码必须经过审查
2. **测试覆盖**: 新功能必须有测试
3. **文档同步**: 代码变更必须同步文档
4. **定期审查**: 每双周审查技术债清单

---

**最后更新**: 2026-03-01  
**下次审查**: 2026-03-15  
**负责人**: Grape Team
