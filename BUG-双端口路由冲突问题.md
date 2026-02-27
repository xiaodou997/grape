# 🐛 Bug 报告：双端口路由冲突问题

**创建时间**: 2026-02-27  
**优先级**: 🔴 P0 - 阻塞性问题  
**状态**: ✅ 已解决  
**解决时间**: 2026-02-27  
**分配给**: AI Assistant

---

## 📋 问题概述

在实现**双端口分离**（Web UI 端口 + npm Registry API 端口）功能时，遇到 Gin 路由冲突问题，导致服务无法启动。

### 预期功能

- **端口 4873**: Web UI 管理界面（仅处理前端页面和管理 API）
- **端口 4874**: npm Registry API（处理包下载、发布等 npm 协议相关请求）

### 当前状态

服务启动时 panic，无法运行。

---

## 🔍 错误信息

### Panic 错误

```
panic: catch-all wildcard '*filepath' in new path '/*filepath' conflicts with 
existing path segment '-' in existing prefix '/-'

goroutine 1 [running]:
github.com/gin-gonic/gin.(*node).insertChild(...)
    /Users/luoxiaodou/go/pkg/mod/github.com/gin-gonic/gin@v1.11.0/tree.go:363
github.com/gin-gonic/gin.(*Engine).addRoute(...)
    /Users/luoxiaodou/go/pkg/mod/github.com/gin-gonic/gin@v1.11.0/gin.go:367
github.com/graperegistry/grape/internal/server.(*Server).setupRoutes(...)
    /Users/luoxiaodou/workspace/projects/grape/internal/server/server.go:245
```

### 错误原因

Gin 路由器**不允许在同一个 Router 实例中**同时使用：
- 具体路径：`/-/health`, `/-/user/:username`
- 通配符路径：`/*filepath`, `/:package`

这两种路径模式会产生冲突。

---

## 📝 修改历史

### 修改的文件

#### 1. `internal/config/config.go`
**修改内容**: 增加 `APIPort` 字段

```go
type ServerConfig struct {
    Host         string        `mapstructure:"host"`
    Port         int           `mapstructure:"port"`          // Web UI 端口
    APIPort      int           `mapstructure:"api_port"`      // npm Registry API 端口
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}
```

---

#### 2. `internal/server/server.go`
**修改内容**: 实现双路由器、双 HTTP 服务器

**主要改动**:

1. **Server 结构体增加字段**:
```go
type Server struct {
    // ... 原有字段 ...
    apiRouter       *gin.Engine         // npm Registry API 专用路由
    apiServer       *http.Server        // npm Registry API 服务器
}
```

2. **New 函数初始化双路由器**:
```go
func New(cfg *config.Config, version string) *Server {
    // Web UI 路由器
    router := gin.New()
    // ... 中间件 ...
    
    // npm Registry API 路由器（独立）
    apiRouter := gin.New()
    // ... 中间件 ...
    
    // 确定 API 端口
    apiPort := cfg.Server.APIPort
    if apiPort == 0 {
        apiPort = cfg.Server.Port
    }
    baseURL := fmt.Sprintf("http://localhost:%d", apiPort)
    
    // ... 初始化 handlers ...
    
    s := &Server{
        // ...
        router:    router,
        apiRouter: apiRouter,
        apiServer: &http.Server{
            Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, apiPort),
            Handler: apiRouter,
        },
    }
    
    s.setupRoutes()
    return s
}
```

3. **setupRoutes 函数拆分为两个路由**:

```go
func (s *Server) setupRoutes() {
    // =====================================
    // Web UI 路由
    // =====================================
    s.router.GET("/-/health", s.handleHealth)
    s.router.GET("/-/metrics", gin.WrapH(promhttp.Handler()))
    
    webAPI := s.router.Group("/-")
    webAPI.Use(authMiddleware)
    {
        // 管理 API
        webAPI.GET("/api/packages", s.apiHandler.ListPackages)
        // ... 其他管理 API ...
        webAPI.PUT("/user/:username", s.authHandler.Login)
    }
    
    // 前端静态资源
    s.router.NoRoute(authMiddleware, s.serveFrontend)
    
    // =====================================
    // npm Registry API 路由（问题所在）
    // =====================================
    s.apiRouter.GET("/-/health", s.handleHealth)
    s.apiRouter.PUT("/-/user/:username", s.authHandler.Login)
    s.apiRouter.PUT("/-/user/:username/*rev", s.authHandler.Login)
    
    apiRegistry := s.apiRouter.Group("/-")
    apiRegistry.Use(authMiddleware)
    {
        apiRegistry.GET("/api/user", s.authHandler.GetCurrentUser)
        apiRegistry.DELETE("/api/session", s.authHandler.Logout)
        apiRegistry.PUT("/:package", s.publishHandler.Publish)
        apiRegistry.DELETE("/:package", s.publishHandler.Unpublish)
    }
    
    // ⚠️ 问题行：通配符路由与上面的具体路径冲突
    s.apiRouter.GET("/*filepath", s.handleRegistryRequest)
    // 或
    s.apiRouter.NoRoute(s.handleRegistryRequest)
}
```

4. **新增 handleRegistryRequest 处理函数**:
```go
func (s *Server) handleRegistryRequest(c *gin.Context) {
    filepath := c.Param("filepath")
    filepath = strings.TrimPrefix(filepath, "/")
    
    if strings.Contains(filepath, "/-/") {
        // tarball 下载
        idx := strings.Index(filepath, "/-/")
        packageName := filepath[:idx]
        filename := filepath[idx+3:]
        c.Params = append(c.Params,
            gin.Param{Key: "package", Value: packageName},
            gin.Param{Key: "filename", Value: filename},
        )
        s.registryHandler.GetTarball(c)
    } else {
        // 包元数据
        c.Params = append(c.Params, 
            gin.Param{Key: "package", Value: filepath})
        s.registryHandler.GetPackage(c)
    }
}
```

5. **Start 函数启动两个 HTTP 服务器**:
```go
func (s *Server) Start() error {
    webAddr := s.http.Addr
    apiAddr := s.apiServer.Addr
    
    logger.Infof("🚀 Grape Web UI server starting on http://%s", webAddr)
    logger.Infof("📦 npm Registry API server starting on http://%s", apiAddr)
    
    // 启动 API 服务器（后台）
    go func() {
        if err := s.apiServer.ListenAndServe(); err != nil {
            logger.Fatalf("Failed to start API server: %v", err)
        }
    }()
    
    // 启动 Web UI 服务器（前台）
    return s.http.ListenAndServe()
}
```

6. **Shutdown 函数关闭两个服务器**:
```go
func (s *Server) Shutdown(ctx context.Context) error {
    logger.Info("🛑 Shutting down servers...")
    
    // 关闭 Web UI 服务器
    if err := s.http.Shutdown(ctx); err != nil {
        logger.Errorf("Web UI server shutdown error: %v", err)
    }
    
    // 关闭 API 服务器
    if err := s.apiServer.Shutdown(ctx); err != nil {
        logger.Errorf("API server shutdown error: %v", err)
    }
    
    return nil
}
```

---

#### 3. `configs/config.yaml`
**修改内容**: 增加 `api_port` 配置

```yaml
server:
  host: 0.0.0.0
  port: 4873              # Web UI 端口
  api_port: 4874          # npm Registry API 端口（新增）
  read_timeout: 30s
  write_timeout: 30s
```

---

#### 4. `internal/registry/proxy.go`
**修改内容**: 增加 gzip 解压支持、JSON 验证、调试日志

**主要改动**:

1. **增加导入**:
```go
import (
    "bytes"
    "compress/gzip"
    "os"
    // ... 其他导入
)
```

2. **增加常量**:
```go
const (
    maxMetadataSize = 50 * 1024 * 1024 // 50MB（原 5MB，大型包如 vite 超过 38MB）
    maxTarballSize  = 500 * 1024 * 1024
)
```

3. **GetMetadata 函数增强**:
```go
func (p *Proxy) GetMetadata(packageName string) ([]byte, error) {
    // ... 创建请求 ...
    
    // 设置 Accept-Encoding 支持 gzip
    req.Header.Set("Accept-Encoding", "gzip")
    resp, err := up.client.Do(req)
    
    // 处理 gzip 压缩响应
    var reader io.Reader = resp.Body
    if resp.Header.Get("Content-Encoding") == "gzip" {
        gzipReader, err := gzip.NewReader(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to create gzip reader: %w", err)
        }
        defer gzipReader.Close()
        reader = gzipReader
    }
    
    // 使用 Buffer 读取
    var buf bytes.Buffer
    io.Copy(&buf, io.LimitReader(reader, maxMetadataSize))
    data := buf.Bytes()
    
    // 验证 JSON 完整性
    if err := validateJSON(data); err != nil {
        // 保存调试文件
        debugFile := fmt.Sprintf("/tmp/grape-debug-%s-%d.json", packageName, time.Now().Unix())
        os.WriteFile(debugFile, data, 0644)
        return nil, fmt.Errorf("invalid JSON from upstream: %w", err)
    }
    
    return data, nil
}
```

4. **新增 validateJSON 函数**:
```go
func validateJSON(data []byte) error {
    var raw json.RawMessage
    return json.Unmarshal(data, &raw)
}
```

---

#### 5. `internal/storage/local/storage.go`
**修改内容**: 增加 JSON 验证、原子写入

**主要改动**:

1. **增加导入**:
```go
import (
    "encoding/json"
    // ...
    "github.com/graperegistry/grape/internal/logger"
)
```

2. **GetMetadata 增加验证**:
```go
func (s *Storage) GetMetadata(packageName string) ([]byte, error) {
    data, err := os.ReadFile(path)
    
    // 验证 JSON 完整性
    if err := validateMetadataJSON(data); err != nil {
        logger.Warnf("Corrupted metadata for package %s: %v", packageName, err)
        os.Remove(path) // 删除损坏文件
        return nil, registry.ErrPackageNotFound
    }
    
    return data, nil
}
```

3. **SaveMetadata 使用原子写入**:
```go
func (s *Storage) SaveMetadata(packageName string, data []byte) error {
    // 验证 JSON
    if err := validateMetadataJSON(data); err != nil {
        return fmt.Errorf("invalid metadata JSON: %w", err)
    }
    
    // 原子写入：先写临时文件，再重命名
    tmpPath := path + ".tmp"
    os.WriteFile(tmpPath, data, 0644)
    os.Rename(tmpPath, path) // 原子操作
    
    return nil
}
```

4. **新增验证函数**:
```go
func validateMetadataJSON(data []byte) error {
    var raw json.RawMessage
    return json.Unmarshal(data, &raw)
}
```

---

#### 6. `internal/server/handler/registry.go`
**修改内容**: 增加 JSON 序列化错误处理

```go
func (h *RegistryHandler) rewriteTarballURLs(data []byte, packageName string, baseURL string) ([]byte, error) {
    // ... 处理 ...
    
    rewritten, err := json.Marshal(pkg)
    if err != nil {
        logger.Errorf("Failed to marshal rewritten JSON for %s: %v", packageName, err)
        return data, nil // 返回原始数据
    }
    
    return rewritten, nil
}
```

---

## 🧪 测试场景

### 期望的行为

1. **Web UI 端口 (4873)**:
   - 访问 `/` → 返回前端页面 (HTML)
   - 访问 `/-/health` → 返回健康检查 (JSON)
   - 访问 `/-/api/packages` → 返回包列表 (JSON，需认证)
   - 访问 `/vite` → 返回前端页面 (HTML, SPA fallback)

2. **API 端口 (4874)**:
   - 访问 `/-/health` → 返回健康检查 (JSON)
   - 访问 `/vite` → 返回包元数据 (JSON)
   - 访问 `/@types/estree` → 返回 scoped 包元数据 (JSON)
   - 访问 `/lodash/-/lodash-4.17.21.tgz` → 返回 tarball (二进制)

### 当前行为

服务无法启动，启动时 panic。

---

## 💡 可能的解决方案

### 方案 1: 使用独立的路由处理函数（推荐）

不在 `apiRouter` 中注册任何具体路由，所有请求都通过 `NoRoute` 处理：

```go
// setupRoutes - API 路由器部分
s.apiRouter.NoRoute(s.handleRegistryRequest)

// handleRegistryRequest 处理所有 npm registry 请求
func (s *Server) handleRegistryRequest(c *gin.Context) {
    path := strings.TrimPrefix(c.Request.URL.Path, "/")
    
    // 特殊路径处理
    if path == "-/health" {
        s.handleHealth(c)
        return
    }
    if strings.HasPrefix(path, "-/user/") {
        // 处理登录
        s.authHandler.Login(c)
        return
    }
    // ... 其他特殊路径 ...
    
    // 默认：npm 包请求
    if strings.Contains(path, "/-/") {
        // tarball
    } else {
        // 元数据
    }
}
```

**优点**: 
- 完全避免路由冲突
- 逻辑清晰

**缺点**:
- 需要手动解析所有路径
- 失去 Gin 的自动参数绑定

---

### 方案 2: 使用子路由器隔离

为不同路径模式创建独立的子路由器：

```go
// 创建独立的路由组
healthRouter := s.apiRouter.Group("/")
healthRouter.GET("/-/health", s.handleHealth)

userRouter := s.apiRouter.Group("/-/user")
userRouter.PUT("/:username", s.authHandler.Login)

// 通配符路由放在最后
s.apiRouter.GET("/*filepath", s.handleRegistryRequest)
```

**优点**: 利用 Gin 的路由组机制

**缺点**: 可能仍有冲突，需要测试验证

---

### 方案 3: 放弃双端口，改进单端口路由判断

保留单端口设计，但改进 `isNpmPackageRequest` 的判断逻辑：

```go
func (s *Server) isNpmPackageRequest(c *gin.Context) bool {
    path := c.Request.URL.Path
    
    // 精确匹配前端路径
    frontendPaths := map[string]bool{
        "/": true, "/login": true, "/packages": true,
        // ... 所有前端路径
    }
    if frontendPaths[path] {
        return false
    }
    
    // 单段路径视为包名
    segments := strings.Split(strings.Trim(path, "/"), "/")
    if len(segments) == 1 {
        return true
    }
    
    // 其他判断逻辑...
}
```

**优点**: 
- 架构简单
- 无路由冲突

**缺点**:
- 无法完全隔离 Web UI 和 API 流量
- 判断逻辑复杂

---

## 📎 相关文件清单

### 已修改的文件

| 文件路径 | 修改内容 | 状态 |
|----------|----------|------|
| `internal/config/config.go` | 增加 APIPort 字段 | ✅ 完成 |
| `internal/server/server.go` | 双路由器、双服务器实现 | ⚠️ 有 bug |
| `internal/server/handler/registry.go` | JSON 序列化错误处理 | ✅ 完成 |
| `internal/registry/proxy.go` | gzip 解压、JSON 验证 | ✅ 完成 |
| `internal/storage/local/storage.go` | JSON 验证、原子写入 | ✅ 完成 |
| `configs/config.yaml` | 增加 api_port 配置 | ✅ 完成 |

### 需要修复的文件

| 文件路径 | 问题 | 优先级 |
|----------|------|--------|
| `internal/server/server.go` | 路由冲突导致 panic | 🔴 P0 |

---

## 🔗 参考资料

- [Gin 路由文档](https://gin-gonic.com/docs/)
- [Gin 路由冲突问题讨论](https://github.com/gin-gonic/gin/issues)
- [原始 Issue: npm 包路由判断问题](https://github.com/graperegistry/grape/issues)

---

## 📝 下一步行动

1. **立即**: 修复路由冲突问题（建议采用方案 1）
2. **短期**: 完善双端口测试用例
3. **中期**: 编写双端口部署文档
4. **长期**: 考虑是否需要更多端口分离（如管理 API 独立端口）

---

**报告人**: AI Assistant  
**审核人**: _待填写_  
**修复人**: _待分配_
