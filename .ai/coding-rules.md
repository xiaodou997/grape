# 📝 Grape Registry - 编码规范

**文档版本**: v1.0  
**最后更新**: 2026-02-27  
**适用范围**: 所有参与 Grape 项目的开发者、AI 助手

---

## 🎯 代码风格总结

### Go 代码风格

| 规范 | 规则 | 示例 |
|------|------|------|
| **分号** | ❌ 不使用行尾分号 | `fmt.Println("hello")` |
| **缩进** | Tab 缩进 | Go 标准 |
| **行宽** | 建议 < 120 字符 | 自动换行 |
| **命名** | 驼峰式，首字母大写表示导出 | `NewServer`, `validatePath` |
| **注释** | 文档注释以函数名开头 | `// New 创建新服务器` |
| **错误处理** | 显式返回，不忽略错误 | `if err != nil { return err }` |
| **变量声明** | 短变量优先 | `i := 0` |
| **接口** | 小接口，1-2 个方法 | `io.Reader`, `Storage` |

### TypeScript 代码风格

| 规范 | 规则 | 示例 |
|------|------|------|
| **分号** | ❌ 不使用行尾分号 | `const x = 1` |
| **缩进** | 2 空格 | 前端标准 |
| **行宽** | 建议 < 120 字符 | 自动换行 |
| **命名** | 驼峰式，组件 PascalCase | `getApiBaseUrl`, `App.vue` |
| **类型** | 显式类型注解 | `const count: Ref<number>` |
| **异步** | async/await 优先 | `const data = await api.get()` |
| **导出** | 命名导出优先 | `export const foo = 1` |

---

## 📛 命名规范

### Go 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| **包名** | 小写，无下划线 | `package registry` |
| **结构体** | PascalCase，名词 | `type User struct` |
| **接口** | PascalCase，-er/-able 后缀 | `type Storage interface` |
| **函数** | PascalCase（导出）/ camelCase（内部） | `NewServer`, `validatePath` |
| **变量** | camelCase，简短优先 | `i`, `ctx`, `cfg` |
| **常量** | PascalCase 或 camelCase | `MaxRetries`, `defaultTimeout` |
| **错误** | 小写，不带标点 | `errNotFound`, `ErrUnauthorized` |

### TypeScript 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| **变量** | camelCase | `const userName = 'admin'` |
| **函数** | camelCase | `function getUserInfo()` |
| **组件** | PascalCase | `App.vue`, `PackageList.vue` |
| **类型** | PascalCase | `interface UserInfo` |
| **枚举** | PascalCase，值大写 | `enum Role { ADMIN }` |
| **文件** | kebab-case | `user-info.ts`, `App.vue` |

---

## 📁 目录规范

### Go 项目结构

```
grape/
├── cmd/                    # 可执行程序入口
│   └── grape/              # 主程序
│       └── main.go
├── internal/               # 内部包（不对外暴露）
│   ├── auth/               # 认证模块
│   ├── config/             # 配置管理
│   ├── db/                 # 数据库层
│   ├── logger/             # 日志系统
│   ├── metrics/            # 监控指标
│   ├── registry/           # Registry 核心
│   ├── server/             # HTTP 服务
│   │   ├── server.go
│   │   └── handler/        # HTTP Handler
│   ├── storage/            # 存储抽象
│   ├── webhook/            # Webhook 事件
│   └── web/                # 前端嵌入
├── pkg/                    # 公共包（可对外暴露）
│   └── apierr/             # 统一错误码
├── web/                    # 前端源码
│   ├── src/
│   │   ├── api/            # API 客户端
│   │   ├── views/          # 页面组件
│   │   ├── components/     # 通用组件
│   │   ├── stores/         # Pinia 状态
│   │   └── router/         # 路由配置
│   ├── public/             # 静态资源
│   └── dist/               # 构建输出
├── configs/                # 配置文件示例
├── docs/                   # 文档
├── scripts/                # 脚本工具
├── test-projects/          # 测试项目
└── data/                   # 运行时数据
```

### 文件命名

| 类型 | 规则 | 示例 |
|------|------|------|
| **Go 源文件** | 小写，下划线分隔 | `user_store.go` |
| **Go 测试文件** | `*_test.go` | `user_store_test.go` |
| **TypeScript** | 小写，kebab-case | `user-info.ts` |
| **Vue 组件** | PascalCase | `UserInfo.vue` |
| **配置文件** | 小写 | `config.yaml` |
| **文档** | 大写或 kebab-case | `README.md`, `api-spec.md` |

---

## 🌐 API 设计规范

### RESTful 风格

| 操作 | HTTP 方法 | 路径 | 说明 |
|------|----------|------|------|
| 列表 | GET | `/-/api/packages` | 获取包列表 |
| 详情 | GET | `/:package` | 获取包元数据 |
| 创建 | PUT | `/:package` | 发布包 |
| 删除 | DELETE | `/:package` | 删除包 |
| 搜索 | GET | `/-/api/search?q=` | 搜索包 |

### 响应格式

**成功响应**:
```json
{
  "name": "lodash",
  "version": "4.17.21",
  "description": "Lodash modular utilities."
}
```

**错误响应**:
```json
{
  "error": "package not found",
  "code": "E_NOT_FOUND"
}
```

### 状态码使用

| 状态码 | 场景 |
|--------|------|
| 200 OK | 成功获取 |
| 201 Created | 成功创建 |
| 204 No Content | 成功删除 |
| 400 Bad Request | 请求参数错误 |
| 401 Unauthorized | 未认证 |
| 403 Forbidden | 无权限 |
| 404 Not Found | 资源不存在 |
| 409 Conflict | 资源冲突（如包已存在） |
| 500 Internal Server Error | 服务器错误 |

### 认证头

```http
Authorization: Bearer <jwt_token>
# 或
Authorization: Bearer <persistent_token>
```

---

## ⚠️ 错误处理规范

### Go 错误处理

**基本原则**:
```go
// ✅ 正确：显式返回错误
if err != nil {
    return err
}

// ✅ 正确：包装错误上下文
if err != nil {
    return fmt.Errorf("failed to save package: %w", err)
}

// ❌ 错误：忽略错误
doSomething()  // 错误！

// ❌ 错误：panic 滥用
if err != nil {
    panic(err)  // 仅在 main 中使用
}
```

**错误类型**:
```go
// 定义错误变量
var (
    ErrNotFound = fmt.Errorf("not found")
    ErrUnauthorized = fmt.Errorf("unauthorized")
    ErrForbidden = fmt.Errorf("forbidden")
)

// 错误判断
if errors.Is(err, ErrNotFound) {
    // 处理特定错误
}
```

**日志记录**:
```go
// 使用结构化日志
logger.Errorf("Failed to save package: %v", err)
logger.Warnf("Invalid JSON from upstream: %v", err)
logger.Infof("Package published successfully: %s", packageName)
```

### TypeScript 错误处理

**基本原则**:
```typescript
// ✅ 正确：try-catch 包裹异步操作
try {
    const data = await api.get('/package')
} catch (error) {
    if (axios.isAxiosError(error)) {
        console.error('API error:', error.response?.data)
    }
}

// ✅ 正确：统一错误拦截
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            // 跳转登录
        }
        return Promise.reject(error)
    }
)
```

---

## 📊 数据结构约定

### Go 结构体

**基本规则**:
```go
// ✅ 正确：使用指针表示可选
type User struct {
    ID        uint       `gorm:"primaryKey" json:"id"`
    Username  string     `gorm:"uniqueIndex;size:100;not null" json:"username"`
    Email     string     `gorm:"size:255" json:"email"`
    LastLogin *time.Time `json:"lastLogin,omitempty"`  // 可选字段
}

// ✅ 正确：JSON 标签使用 camelCase
type Package struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
}
```

**验证标签**:
```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=100"`
    Password string `json:"password" binding:"required,min=6"`
    Email    string `json:"email" binding:"omitempty,email"`
    Role     string `json:"role" binding:"oneof=admin developer readonly"`
}
```

### TypeScript 类型

**接口定义**:
```typescript
// ✅ 正确：使用 interface 定义对象形状
interface Package {
    name: string
    version: string
    description?: string  // 可选字段
    private?: boolean
}

// ✅ 正确：使用 type 定义联合类型
type Role = 'admin' | 'developer' | 'readonly'

// ✅ 正确：泛型约束
interface ApiResponse<T> {
    data: T
    error?: string
    code?: string
}
```

---

## 📦 提交规范

### Git Commit Message

**格式**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**:
| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响功能） |
| `refactor` | 重构 |
| `test` | 测试相关 |
| `chore` | 构建/工具/配置 |

**示例**:
```
feat(auth): add persistent token support

- Add Token model and database migration
- Implement /-/npm/v1/tokens API
- Add token validation middleware
- Update frontend with token management UI

Closes #42
```

### 分支命名

| 分支类型 | 命名规则 | 示例 |
|----------|----------|------|
| **主分支** | `main` | `main` |
| **功能分支** | `feature/<name>` | `feature/token-auth` |
| **修复分支** | `fix/<name>` | `fix/route-conflict` |
| **发布分支** | `release/<version>` | `release/v0.2.0` |

---

## 🔧 推荐补充规范

### 日志规范

**日志级别使用**:
```go
// DEBUG: 详细调试信息
logger.Debugf("Request headers: %v", c.Request.Header)

// INFO: 重要运行信息
logger.Infof("Package published: %s@%s", packageName, version)

// WARN: 警告信息（不影响功能）
logger.Warnf("Using default JWT secret, please change in production")

// ERROR: 错误信息（功能受影响）
logger.Errorf("Failed to save package: %v", err)

// FATAL: 致命错误（程序退出）
logger.Fatalf("Database connection failed: %v", err)
```

### 配置规范

**环境变量优先级**:
```
1. 命令行参数 (--config)
2. 环境变量 (GRAPE_JWT_SECRET)
3. 配置文件 (config.yaml)
4. 默认值
```

**配置项命名**:
```yaml
# ✅ 正确：使用下划线分隔
auth:
    jwt_secret: "your-secret"
    jwt_expiry: 24h

# ❌ 错误：使用驼峰
auth:
    jwtSecret: "your-secret"
```

### 测试规范

**Go 测试**:
```go
// ✅ 正确：表格驱动测试
func TestValidatePackageName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid simple", "lodash", false},
        {"valid scoped", "@babel/core", false},
        {"invalid empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validatePackageName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validatePackageName() error = %v", err)
            }
        })
    }
}
```

**TypeScript 测试**:
```typescript
// ✅ 正确：使用 describe 组织测试
describe('Auth API', () => {
    describe('login', () => {
        it('should return token on valid credentials', async () => {
            const response = await authApi.login('admin', 'admin')
            expect(response.data).toHaveProperty('token')
        })

        it('should reject invalid credentials', async () => {
            await expect(authApi.login('admin', 'wrong'))
                .rejects.toThrow()
        })
    })
})
```

---

## 🚫 禁止事项

### Go 禁止

| 禁止 | 原因 | 替代方案 |
|------|------|----------|
| `panic()` 业务逻辑 | 难以恢复 | 返回 error |
| 忽略 error | 隐藏问题 | 显式处理 |
| 全局变量滥用 | 难以测试 | 依赖注入 |
| `interface{}` 滥用 | 类型不安全 | 泛型或具体类型 |
| 循环导入 | 编译错误 | 重构包结构 |

### TypeScript 禁止

| 禁止 | 原因 | 替代方案 |
|------|------|----------|
| `any` 类型 | 失去类型检查 | 定义具体类型 |
| 直接修改 props | Vue 警告 | 使用 emit |
| 硬编码 API URL | 环境耦合 | 环境变量 |
| 嵌套超过 3 层 | 难以维护 | 提取函数 |

---

## 📚 参考资源

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Vue 3 Style Guide](https://vuejs.org/style-guide/)
- [TypeScript Deep Dive](https://basarat.gitbook.io/typescript/)

---

**最后更新**: 2026-02-27  
**下次审查**: 团队规模扩大或技术栈变更时
