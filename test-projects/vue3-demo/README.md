# Vue3 Demo - Grape Registry 测试项目

## 📦 项目介绍

这是一个用于测试 **Grape Registry** 功能的 Vue 3 基础项目，包含常用的前端依赖和完整的项目结构。

## 🚀 快速开始

### 前置要求

1. **Node.js** >= 18.0.0
2. **npm** >= 9.0.0
3. **Grape Registry** 正在运行（端口 4874）

### 安装依赖

```bash
# 进入项目目录
cd test-projects/vue3-demo

# 安装依赖（从 Grape Registry 下载）
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000 查看应用。

### 构建生产版本

```bash
npm run build
```

## 📋 依赖清单

### 运行时依赖

| 包名 | 版本 | 用途 |
|------|------|------|
| vue | ^3.5.25 | Vue 3 框架 |
| vue-router | ^4.5.1 | 路由管理 |
| pinia | ^3.0.3 | 状态管理 |
| axios | ^1.11.0 | HTTP 请求 |
| element-plus | ^2.9.10 | UI 组件库 |
| @element-plus/icons-vue | ^2.3.1 | Element Plus 图标 |

### 开发依赖

| 包名 | 版本 | 用途 |
|------|------|------|
| vite | ^7.3.1 | 构建工具 |
| @vitejs/plugin-vue | ^6.0.2 | Vue 3 插件 |
| typescript | ^5.9.3 | TypeScript 支持 |
| vue-tsc | ^3.0.2 | Vue TypeScript 检查 |
| @types/node | ^24.1.0 | Node.js 类型定义 |
| sass | ^1.90.0 | Sass 预处理器 |

## 🔧 配置说明

### .npmrc 配置

项目使用 `.npmrc` 文件配置私有 Registry：

```
registry=http://localhost:4874
audit=false
fund=false
progress=false
```

这确保所有依赖都从本地 Grape Registry 下载，不影响全局 npm 配置。

## 📁 项目结构

```
vue3-demo/
├── .npmrc                 # npm 配置（指向 Grape Registry）
├── package.json           # 项目依赖配置
├── tsconfig.json          # TypeScript 配置
├── vite.config.ts         # Vite 配置
├── index.html             # HTML 入口
└── src/
    ├── main.ts            # 应用入口
    ├── App.vue            # 根组件
    ├── router.ts          # 路由配置
    ├── stores/
    │   └── counter.ts     # Pinia Store 示例
    └── views/
        └── AboutView.vue  # 关于页面
```

## ✅ 测试清单

安装完成后，验证以下功能：

- [ ] 所有依赖成功安装（无错误）
- [ ] `npm run dev` 正常启动
- [ ] 页面正常显示
- [ ] Element Plus 组件正常工作
- [ ] Pinia 状态管理正常
- [ ] Axios 请求正常
- [ ] Vue Router 路由切换正常

## 🐛 故障排查

### 依赖安装失败

```bash
# 清理缓存
npm cache clean --force

# 删除 node_modules 和 lock 文件
rm -rf node_modules package-lock.json

# 重新安装
npm install
```

### 检查 Grape Registry 状态

```bash
# 检查健康状态
curl http://localhost:4874/-/health

# 检查特定包
curl http://localhost:4874/vue
```

## 📝 测试报告

如果测试通过，请记录以下信息：

- ✅ 安装耗时：___ 秒
- ✅ 下载包数量：___ 个
- ✅ 下载总量：___ MB
- ✅ 是否有失败：是/否

## 🔗 相关链接

- [Grape Registry 项目](https://github.com/graperegistry/grape)
- [Vue 3 文档](https://vuejs.org/)
- [Vite 文档](https://vitejs.dev/)
- [Element Plus 文档](https://element-plus.org/)
