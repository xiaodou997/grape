# 🐳 Docker 镜像构建使用指南

**工作流**: `.github/workflows/docker.yml`  
**触发方式**: 手动触发  
**镜像仓库**: `ghcr.io/xiaodou997/grape`

---

## 🚀 如何手动触发构建

### 步骤 1: 进入 Actions 页面

访问：https://github.com/xiaodou997/grape/actions

### 步骤 2: 选择 Docker 工作流

左侧栏 → **Docker** → 点击 **"Run workflow"** 按钮

### 步骤 3: 填写参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| **镜像版本标签** | 镜像的 tag | `latest` | `v0.1.0`, `latest`, `dev` |
| **推送到 Registry** | 是否推送到 GHCR | `true` | `true` / `false` |
| **构建平台** | 目标平台架构 | `linux/amd64,linux/arm64` | 逗号分隔 |

### 步骤 4: 运行

点击 **"Run workflow"** 开始构建

---

## 📋 使用场景

### 场景 1: 发布正式版本

```yaml
版本标签：v0.1.0
推送到 Registry: ✓
构建平台：linux/amd64,linux/arm64
```

**结果**:
- 镜像标签：`ghcr.io/xiaodou997/grape:v0.1.0`
- 多平台构建：AMD64 + ARM64
- 自动推送到 GHCR

### 场景 2: 本地测试构建

```yaml
版本标签：test-build
推送到 Registry: ✗
构建平台：linux/amd64
```

**结果**:
- 仅在 GitHub Actions 构建，不推送
- 适合测试 Dockerfile 是否有问题
- 节省存储空间

### 场景 3: 更新 latest 标签

```yaml
版本标签：latest
推送到 Registry: ✓
构建平台：linux/amd64,linux/arm64
```

**结果**:
- 镜像标签：`ghcr.io/xiaodou997/grape:latest`
- 覆盖旧的 latest 标签

---

## 🎯 常用命令

### 拉取镜像

```bash
# 登录 GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u xiaodou997 --password-stdin

# 拉取最新版本
docker pull ghcr.io/xiaodou997/grape:latest

# 拉取特定版本
docker pull ghcr.io/xiaodou997/grape:v0.1.0
```

### 运行容器

```bash
# 基本运行
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  ghcr.io/xiaodou997/grape:latest

# 持久化数据
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  -v grape-data:/data \
  ghcr.io/xiaodou997/grape:latest

# 使用配置文件
docker run -d \
  --name grape \
  -p 4873:4873 \
  -p 4874:4874 \
  -v ./configs:/app/configs \
  -v grape-data:/data \
  ghcr.io/xiaodou997/grape:latest \
  --config /app/configs/config.yaml
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  grape:
    image: ghcr.io/xiaodou997/grape:v0.1.0
    container_name: grape
    ports:
      - "4873:4873"
      - "4874:4874"
    volumes:
      - grape-data:/data
      - ./configs:/app/configs
    restart: unless-stopped
```

**启动命令**:
```bash
docker-compose up -d

# 查看日志
docker-compose logs -f grape

# 停止
docker-compose down
```

---

## 📊 镜像标签策略

### 推荐标签

| 标签 | 用途 | 更新频率 |
|------|------|----------|
| `latest` | 最新稳定版 | 每次正式发布 |
| `v0.1.0` | 特定版本 | 固定不变 |
| `dev` | 开发版 | 随时更新 |
| `sha-abc123` | Commit SHA | 每次 commit |

### 标签管理

**查看可用标签**:
```bash
# 访问 GitHub Packages 页面
https://github.com/xiaodou997/grape/pkgs/container/grape
```

**删除旧标签**:
```
1. GitHub Packages → grape
2. 点击版本 → Delete version
```

---

## 🔧 高级配置

### 自定义构建平台

**可用平台**:
- `linux/amd64` - Intel/AMD x86_64
- `linux/arm64` - ARM 64-bit (Raspberry Pi, M1/M2)
- `linux/arm/v7` - ARM 32-bit
- `linux/ppc64le` - PowerPC
- `linux/s390x` - IBM Z

**多平台构建**:
```
linux/amd64,linux/arm64,linux/arm/v7
```

### 本地测试构建

**导出到本地**:
```yaml
推送到 Registry: ✗
```

**下载镜像**:
1. 构建完成后进入 Actions 页面
2. 点击对应构建任务
3. 下载 artifacts 中的镜像文件

### 自动清理旧镜像

**设置保留策略**:
```
GitHub Settings → Packages → grape
→ Container registry cleanup
→ Delete untagged images after: 30 days
```

---

## 🐛 故障排查

### 问题 1: 构建失败

**检查 Dockerfile**:
```bash
# 本地测试构建
docker build -t grape:test .
```

**查看构建日志**:
```
Actions → Docker → 对应构建 → 查看详细日志
```

### 问题 2: 推送失败

**检查权限**:
```
Settings → Actions → General
→ Workflow permissions
→ Read and write permissions ✓
```

**检查 Token**:
```
Settings → Developer settings → Personal access tokens
→ 确保有 read:packages, write:packages 权限
```

### 问题 3: 镜像拉取失败

**认证问题**:
```bash
# 重新登录
docker logout ghcr.io
echo $GITHUB_TOKEN | docker login ghcr.io -u xiaodou997 --password-stdin
```

**镜像不存在**:
```bash
# 检查标签是否正确
docker pull ghcr.io/xiaodou997/grape:v0.1.0

# 查看可用标签
https://github.com/xiaodou997/grape/pkgs/container/grape
```

---

## 📈 最佳实践

### 1. 版本发布流程

```
1. 本地测试构建 ✓
2. 推送测试版本到 GHCR ✓
3. 测试镜像运行正常 ✓
4. 正式构建 Release 版本 ✓
5. 更新 latest 标签 ✓
```

### 2. 标签命名规范

- ✅ `v0.1.0` - 语义化版本
- ✅ `latest` - 最新稳定版
- ✅ `dev-20240101` - 开发版带日期
- ❌ `test`, `abc`, `123` - 无意义标签

### 3. 平台选择

| 场景 | 推荐平台 |
|------|----------|
| 本地开发 | `linux/amd64` |
| 生产环境 | `linux/amd64,linux/arm64` |
| 嵌入式设备 | `linux/arm/v7,linux/arm64` |

### 4. 存储管理

- 定期清理旧版本镜像
- 使用标签而非 `latest` 生产部署
- 设置自动清理策略（30 天）

---

## 🔗 相关资源

- [GitHub Packages](https://github.com/xiaodou997/grape/pkgs/container/grape)
- [Docker 工作流配置](../.github/workflows/docker.yml)
- [Dockerfile](../Dockerfile)
- [Docker Compose 配置](../docker-compose.yml)

---

**最后更新**: 2026-03-01  
**维护者**: Grape Team
