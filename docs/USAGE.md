# Grape 使用指南

本文档详细介绍如何使用 Grape 私有 npm 仓库，包括各种包管理器的配置和使用方法。

## 目录

- [快速开始](#快速开始)
- [npm 配置和使用](#npm-配置和使用)
- [pnpm 配置和使用](#pnpm-配置和使用)
- [yarn 配置和使用](#yarn-配置和使用)
- [bun 配置和使用](#bun-配置和使用)
- [项目级配置](#项目级配置)
- [用户认证](#用户认证)
- [发布私有包](#发布私有包)
- [多上游配置](#多上游配置)
- [常见问题](#常见问题)

---

## 快速开始

### 1. 启动 Grape 服务

```bash
# 本地开发
./grape

# 或使用配置文件
./grape -c config.yaml
```

### 2. 配置包管理器

```bash
# npm
npm set registry http://localhost:4873

# pnpm
pnpm config set registry http://localhost:4873

# yarn
yarn config set registry http://localhost:4873
```

### 3. 安装包

```bash
npm install lodash
```

### 4. 登录（发布包需要）

```bash
npm login --registry http://localhost:4873
# 用户名：admin
# 密码：admin
```

---

## npm 配置和使用

### 全局配置

```bash
# 设置 registry
npm set registry http://localhost:4873

# 查看当前配置
npm config list

# 查看 registry
npm get registry
```

### 按 scope 配置

```bash
# 仅 @mycompany 包使用私有源
npm set @mycompany:registry http://localhost:4873

# 其他包仍使用官方源
npm get registry  # https://registry.npmjs.org
npm get @mycompany:registry  # http://localhost:4873
```

### 安装包

```bash
# 安装公共包（自动缓存）
npm install lodash
npm install express

# 安装 scoped 包
npm install @babel/core
npm install @mycompany/utils

# 安装指定版本
npm install lodash@4.17.21

# 安装为开发依赖
npm install -D typescript

# 全局安装
npm install -g pnpm
```

### 发布包

```bash
# 登录
npm login --registry http://localhost:4873

# 发布
npm publish --registry http://localhost:4873

# 发布 beta 版本
npm publish --tag beta --registry http://localhost:4873

# 发布 scoped 包
npm publish --access public --registry http://localhost:4873
```

### 删除包

```bash
# 删除特定版本
npm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# 删除整个包（谨慎操作）
npm unpublish @mycompany/my-package --force --registry http://localhost:4873
```

### 恢复官方源

```bash
npm set registry https://registry.npmjs.org
```

---

## pnpm 配置和使用

### 全局配置

```bash
# 设置 registry
pnpm config set registry http://localhost:4873

# 查看配置
pnpm config list

# 按 scope 配置
pnpm config set @mycompany:registry http://localhost:4873
```

### 安装包

```bash
# 安装包
pnpm add lodash
pnpm add @mycompany/utils

# 安装指定版本
pnpm add lodash@4.17.21

# 开发依赖
pnpm add -D typescript

# 全局安装
pnpm add -g typescript
```

### 发布包

```bash
# 登录
pnpm login --registry http://localhost:4873

# 发布
pnpm publish --registry http://localhost:4873

# 发布 beta 版本
pnpm publish --tag beta --registry http://localhost:4873
```

### 删除包

```bash
# 删除特定版本
pnpm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# 删除整个包
pnpm unpublish @mycompany/my-package --force --registry http://localhost:4873
```

---

## yarn 配置和使用

### Yarn v1 (Classic)

```bash
# 设置 registry
yarn config set registry http://localhost:4873

# 按 scope 配置
yarn config set @mycompany:registry http://localhost:4873

# 查看配置
yarn config list
```

### Yarn v2+ (Berry)

```bash
# 设置 registry
yarn config set npmRegistryServer http://localhost:4873

# 设置特定 scope 的源
yarn config set npmScopes.mycompany.npmRegistryServer http://localhost:4873

# 查看配置
yarn config
```

### 安装包

```bash
# 安装包
yarn add lodash
yarn add @mycompany/utils

# 安装指定版本
yarn add lodash@4.17.21

# 开发依赖
yarn add -D typescript

# 全局安装（仅 yarn v1）
yarn global add typescript
```

### 发布包

```bash
# 登录
yarn login --registry http://localhost:4873

# 发布
yarn publish --registry http://localhost:4873

# 指定新版本
yarn publish --new-version 1.0.1 --registry http://localhost:4873
```

### 删除包

```bash
# 删除特定版本
yarn unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# 删除整个包
yarn unpublish @mycompany/my-package --force --registry http://localhost:4873
```

---

## bun 配置和使用

### 配置文件方式（推荐）

创建 `bunfig.toml` 文件：

```toml
# 项目级配置（项目根目录）
[install]
registry = "http://localhost:4873"

# 或全局配置（~/.bunfig.toml）
```

### 环境变量方式

```bash
export BUN_CONFIG_REGISTRY=http://localhost:4873
```

### 命令行方式

```bash
# 安装时指定源
bun add lodash --registry http://localhost:4873
```

### 安装包

```bash
# 安装包
bun add lodash
bun add @mycompany/utils

# 安装指定版本
bun add lodash@4.17.21

# 开发依赖
bun add -d typescript

# 全局安装
bun add -g typescript
```

### 发布包

```bash
# 登录（使用 npm）
npm login --registry http://localhost:4873

# 发布
bun publish --registry http://localhost:4873
```

### 注意事项

⚠️ **bun 不支持 .npmrc 文件**，必须使用 `bunfig.toml` 或环境变量。

---

## 项目级配置

### .npmrc 文件

在项目根目录创建 `.npmrc` 文件：

```ini
# .npmrc 示例

# 默认 registry
registry=http://localhost:4873

# 特定 scope 使用私有源
@mycompany:registry=http://localhost:4873
@internal:registry=http://localhost:4873

# 另一个 scope 使用其他源
@partner:registry=https://npm.partner.com

# 认证 token（登录后自动生成）
//localhost:4873/:_authToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### .npmrc 位置

| 位置 | 说明 |
|------|------|
| `./.npmrc` | 项目级配置（推荐） |
| `~/.npmrc` | 用户级配置 |
| `$PREFIX/etc/npmrc` | 全局配置 |

### 优势

- ✅ 配置跟随项目，团队成员无需手动配置
- ✅ 支持 CI/CD 环境自动使用正确的源
- ✅ 不同 scope 可以路由到不同的源

### 示例：多 registry 配置

```ini
# .npmrc
registry=https://registry.npmjs.org

# 公司私有包使用私有源
@mycompany:registry=http://localhost:4873

# 合作伙伴包使用其他源
@partner:registry=https://npm.partner.com
```

---

## 用户认证

### 登录

```bash
# npm
npm login --registry http://localhost:4873

# pnpm
pnpm login --registry http://localhost:4873

# yarn v1
yarn login --registry http://localhost:4873

# bun（使用 npm 登录）
npm login --registry http://localhost:4873
```

**登录提示：**

```
Username: admin
Password: admin
Email: (optional)
```

### 登出

```bash
# npm
npm logout --registry http://localhost:4873

# pnpm
pnpm logout --registry http://localhost:4873

# yarn
yarn logout
```

### 查看当前用户

```bash
npm whoami --registry http://localhost:4873
```

### Token 管理

Token 保存在以下位置：

- **npm/pnpm**: `~/.npmrc`
- **yarn**: `~/.yarnrc`
- **bun**: `~/.bunfig.toml`（需手动添加）

```ini
//localhost:4873/:_authToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token 有效期

默认 24 小时，可通过配置调整：

```yaml
# config.yaml
auth:
  jwt_expiry: 8h  # 8 小时
```

---

## 发布私有包

### 准备工作

1. **配置 package.json**

```json
{
  "name": "@mycompany/my-package",
  "version": "1.0.0",
  "private": false,
  "description": "My private package",
  "main": "index.js",
  "publishConfig": {
    "registry": "http://localhost:4873"
  }
}
```

2. **登录认证**

```bash
npm login --registry http://localhost:4873
```

### 发布流程

```bash
# 1. 确保代码已准备好
npm test
npm run build

# 2. 更新版本号
npm version patch  # 或 minor/major

# 3. 发布
npm publish

# 如果 package.json 没有配置 publishConfig
npm publish --registry http://localhost:4873

# 发布 scoped 包（需要指定 access）
npm publish --access public --registry http://localhost:4873
```

### 发布 scoped 包

```json
{
  "name": "@mycompany/utils",
  "version": "1.0.0"
}
```

```bash
# 发布 scoped 包
npm publish --access public --registry http://localhost:4873
```

### 发布 beta 版本

```bash
# 发布 beta 标签
npm publish --tag beta --registry http://localhost:4873

# 安装 beta 版本
npm install @mycompany/utils@beta --registry http://localhost:4873
```

### 权限说明

| 操作 | 权限要求 |
|------|----------|
| 安装包 | 无需认证 |
| 发布包 | 已认证用户 |
| 删除包 | 包维护者或管理员 |
| 删除他人包 | 仅管理员 |

---

## 多上游配置

### 配置示例

```yaml
# config.yaml
registry:
  upstreams:
    # 默认上游（公共包）
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""
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
      scope: "@company"
      timeout: 30s
      enabled: true

    # 内部工具包
    - name: "internal-tools"
      url: "https://npm-internal.company.com"
      scope: "@internal"
      timeout: 30s
      enabled: true
```

### 路由规则

| 包名 | 匹配上游 |
|------|---------|
| `lodash` | npmjs (默认) |
| `express` | npmjs (默认) |
| `@babel/core` | npmjs (默认) |
| `@company/utils` | company-private |
| `@company/ui-kit` | company-private |
| `@internal/cli` | internal-tools |

### 查看上游配置

```bash
curl http://localhost:4873/-/api/upstreams
```

---

## 常见问题

### 1. 404 Not Found

```bash
npm ERR! 404 Not Found - GET http://localhost:4873/lodash
```

**原因：**

- Grape 服务未启动
- 配置错误的端口或地址
- 上游连接失败

**解决：**

```bash
# 检查 Grape 是否运行
ps aux | grep grape

# 检查端口
netstat -tlnp | grep 4873

# 测试健康检查
curl http://localhost:4873/-/health

# 检查上游连接
curl -I https://registry.npmjs.org/lodash
```

### 2. 401 Unauthorized

```bash
npm ERR! 401 Unauthorized
```

**原因：**

- 未登录或 token 过期
- token 配置错误

**解决：**

```bash
# 重新登录
npm logout --registry http://localhost:4873
npm login --registry http://localhost:4873

# 检查 .npmrc 配置
cat ~/.npmrc | grep authToken
```

### 3. 403 Forbidden

```bash
npm ERR! 403 Forbidden
```

**原因：**

- 尝试发布已存在的版本
- 权限不足

**解决：**

```bash
# 检查版本是否已存在
npm view @mycompany/package versions

# 更新版本号后重新发布
npm version patch
npm publish
```

### 4. 网络超时

```bash
npm ERR! network timeout
```

**解决：**

```bash
# 增加超时时间
npm config set fetch-timeout 600000
npm config set fetch-retry-mintimeout 20000
npm config set fetch-retry-maxtimeout 120000

# 或使用更快的上游
# config.yaml
registry:
  upstreams:
    - name: "npmmirror"
      url: "https://registry.npmmirror.com"
      scope: ""
      enabled: true
```

### 5. SSL 证书问题

如果使用 HTTPS 和自签名证书：

```bash
# 禁用严格 SSL（仅开发环境）
npm config set //localhost:4873/:strict-ssl false

# 或设置 CA 证书
npm config set cafile /path/to/ca.crt
```

### 6. bun 配置不生效

**原因：** bun 不支持 .npmrc

**解决：**

```toml
# 使用 bunfig.toml
[install]
registry = "http://localhost:4873"
```

或

```bash
# 使用环境变量
export BUN_CONFIG_REGISTRY=http://localhost:4873
```

### 7. scoped 包发布失败

```bash
npm ERR! 400 Bad Request
```

**解决：**

```bash
# scoped 包需要指定 access
npm publish --access public --registry http://localhost:4873

# 或私有包
npm publish --access restricted --registry http://localhost:4873
```

---

## 命令速查表

| 操作 | npm | pnpm | yarn v1 | bun |
|------|-----|------|---------|-----|
| 设置源 | `npm set registry <url>` | `pnpm config set registry <url>` | `yarn config set registry <url>` | 编辑 `bunfig.toml` |
| 登录 | `npm login` | `pnpm login` | `yarn login` | `npm login` |
| 登出 | `npm logout` | `pnpm logout` | `yarn logout` | `npm logout` |
| 安装包 | `npm install <pkg>` | `pnpm add <pkg>` | `yarn add <pkg>` | `bun add <pkg>` |
| 安装依赖 | `npm install` | `pnpm install` | `yarn` | `bun install` |
| 发布包 | `npm publish` | `pnpm publish` | `yarn publish` | `bun publish` |
| 删除包 | `npm unpublish <pkg>` | `pnpm unpublish <pkg>` | `yarn unpublish <pkg>` | - |
| 查看配置 | `npm config list` | `pnpm config list` | `yarn config list` | `cat ~/.bunfig.toml` |

---

## Web UI 操作

### 访问 Web 界面

```
http://localhost:4873
```

### 功能

1. **包浏览** - 查看已缓存的包列表和详情
2. **包搜索** - 搜索包名和描述
3. **用户登录** - 登录获取认证 token
4. **用户管理** - 管理员创建/删除用户
5. **系统设置** - 查看统计信息和配置

### 登录 Web UI

1. 访问 http://localhost:4873
2. 点击右上角「登录」
3. 输入用户名和密码（默认：admin / admin）
4. 登录后访问管理功能

---

## 最佳实践

### 1. 团队配置

```ini
# .npmrc（项目根目录）
registry=https://registry.npmjs.org
@mycompany:registry=http://localhost:4873
```

### 2. CI/CD 配置

```yaml
# .github/workflows/ci.yml
- name: Setup npm registry
  run: |
    npm set registry http://localhost:4873
    npm set //localhost:4873/:_authToken ${{ secrets.NPM_TOKEN }}
```

### 3. 多环境配置

```bash
# 开发环境
npm set registry http://dev-grape:4873

# 生产环境
npm set registry http://prod-grape:4873
```

### 4. 包版本管理

```bash
# 使用语义化版本
npm version patch  # 1.0.0 -> 1.0.1
npm version minor  # 1.0.0 -> 1.1.0
npm version major  # 1.0.0 -> 2.0.0
```

---

## 相关文档

- [API 文档](API.md) - API 参考
- [配置指南](../configs/README.md) - 配置文件详解
- [部署指南](DEPLOYMENT.md) - 部署说明
- [Webhook 文档](WEBHOOKS.md) - Webhook 使用
