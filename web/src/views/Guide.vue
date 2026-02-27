<template>
  <div class="guide-page fade-in">
    <el-card class="guide-header">
      <h1>📖 使用指南</h1>
      <p>如何配置 npm、pnpm、yarn、bun 使用 Grape 私有源</p>
    </el-card>

    <!-- Tab 切换不同包管理器 -->
    <el-tabs v-model="activeTab" class="guide-tabs">
      <!-- npm -->
      <el-tab-pane label="npm" name="npm">
        <GuideSection title="配置源">
          <CodeBlock :code="npmCommands.config" />
          <p class="tip">推荐使用 scope 配置，只将特定范围的包指向私有源</p>
          <CodeBlock :code="npmCommands.scopeConfig" />
        </GuideSection>

        <GuideSection title="用户认证">
          <CodeBlock :code="npmCommands.login" />
          <p class="tip">默认用户名/密码: admin / admin</p>
        </GuideSection>

        <GuideSection title="安装包">
          <CodeBlock :code="npmCommands.install" />
        </GuideSection>

        <GuideSection title="发布包">
          <CodeBlock :code="npmCommands.publish" />
          <p class="tip">在 package.json 中配置 publishConfig 可以省略 --registry 参数</p>
        </GuideSection>

        <GuideSection title="删除包">
          <CodeBlock :code="npmCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- pnpm -->
      <el-tab-pane label="pnpm" name="pnpm">
        <GuideSection title="配置源">
          <CodeBlock :code="pnpmCommands.config" />
          <p class="tip">推荐使用 scope 配置</p>
          <CodeBlock :code="pnpmCommands.scopeConfig" />
        </GuideSection>

        <GuideSection title="用户认证">
          <CodeBlock :code="pnpmCommands.login" />
        </GuideSection>

        <GuideSection title="安装包">
          <CodeBlock :code="pnpmCommands.install" />
        </GuideSection>

        <GuideSection title="发布包">
          <CodeBlock :code="pnpmCommands.publish" />
        </GuideSection>

        <GuideSection title="删除包">
          <CodeBlock :code="pnpmCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- yarn -->
      <el-tab-pane label="yarn" name="yarn">
        <GuideSection title="配置源 (Yarn v1)">
          <CodeBlock :code="yarnCommands.config" />
          <p class="tip">推荐使用 scope 配置</p>
          <CodeBlock :code="yarnCommands.scopeConfig" />
        </GuideSection>

        <GuideSection title="配置源 (Yarn v2+/berry)">
          <CodeBlock :code="yarnCommands.configV2" />
        </GuideSection>

        <GuideSection title="用户认证">
          <CodeBlock :code="yarnCommands.login" />
        </GuideSection>

        <GuideSection title="安装包">
          <CodeBlock :code="yarnCommands.install" />
        </GuideSection>

        <GuideSection title="发布包">
          <CodeBlock :code="yarnCommands.publish" />
        </GuideSection>

        <GuideSection title="删除包">
          <CodeBlock :code="yarnCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- bun -->
      <el-tab-pane label="bun" name="bun">
        <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
          <template #title>
            <strong>注意</strong>：bun 不支持 <code>bun config set</code> 命令，需要通过 bunfig.toml 配置文件或环境变量配置
          </template>
        </el-alert>

        <GuideSection title="配置源">
          <CodeBlock :code="bunCommands.config" />
          <p class="tip">bun 使用 bunfig.toml 配置文件（TOML 格式），不支持 .npmrc 文件</p>
          <CodeBlock :code="bunCommands.scopeConfig" />
        </GuideSection>

        <GuideSection title="用户认证">
          <CodeBlock :code="bunCommands.login" />
          <p class="tip">默认用户名/密码: admin / admin</p>
        </GuideSection>

        <GuideSection title="安装包">
          <CodeBlock :code="bunCommands.install" />
        </GuideSection>

        <GuideSection title="发布包">
          <CodeBlock :code="bunCommands.publish" />
        </GuideSection>
      </el-tab-pane>
    </el-tabs>

    <!-- 项目级配置 -->
    <el-card class="project-config">
      <template #header>
        <div class="card-header">
          <span>📁 项目级配置 (.npmrc)</span>
        </div>
      </template>
      <p class="section-desc">在项目根目录创建 .npmrc 文件，配置跟随项目，团队成员无需手动配置：</p>
      <CodeBlock :code="npmrcExample" language="ini" />
    </el-card>

    <!-- package.json 配置 -->
    <el-card class="package-json-config">
      <template #header>
        <div class="card-header">
          <span>📦 package.json 发布配置</span>
        </div>
      </template>
      <p class="section-desc">在 package.json 中配置 publishConfig，发布时无需指定 --registry：</p>
      <CodeBlock :code="packageJsonExample" language="json" />
    </el-card>

    <!-- 命令速查表 -->
    <el-card class="command-table">
      <template #header>
        <div class="card-header">
          <span>📋 命令速查表</span>
        </div>
      </template>
      <el-table :data="commandTable" stripe>
        <el-table-column prop="operation" label="操作" width="120" />
        <el-table-column prop="npm" label="npm" />
        <el-table-column prop="pnpm" label="pnpm" />
        <el-table-column prop="yarn" label="yarn" />
        <el-table-column prop="bun" label="bun" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import GuideSection from '@/components/GuideSection.vue'
import CodeBlock from '@/components/CodeBlock.vue'

const activeTab = ref('npm')

// npm 命令
const npmCommands = {
  config: `# 设置全局默认源
npm set registry http://localhost:4873

# 恢复官方源
npm set registry https://registry.npmjs.org`,
  scopeConfig: `# 仅设置特定 scope 的源（推荐）
npm set @mycompany:registry http://localhost:4873

# 查看当前配置
npm config list`,
  login: `# 登录
npm login --registry http://localhost:4873

# 登出
npm logout --registry http://localhost:4873`,
  install: `# 安装包
npm install lodash

# 安装指定版本
npm install lodash@4.17.21

# 安装为开发依赖
npm install -D typescript`,
  publish: `# 发布包
npm publish --registry http://localhost:4873

# 发布 beta 版本
npm publish --tag beta`,
  unpublish: `# 删除特定版本
npm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# 删除整个包（谨慎操作）
npm unpublish @mycompany/my-package --force --registry http://localhost:4873`,
}

// pnpm 命令
const pnpmCommands = {
  config: `# 设置全局默认源
pnpm config set registry http://localhost:4873

# 恢复官方源
pnpm config set registry https://registry.npmjs.org`,
  scopeConfig: `# 仅设置特定 scope 的源（推荐）
pnpm config set @mycompany:registry http://localhost:4873

# 查看当前配置
pnpm config list`,
  login: `# 登录
pnpm login --registry http://localhost:4873

# 登出
pnpm logout --registry http://localhost:4873`,
  install: `# 安装包
pnpm add lodash

# 安装指定版本
pnpm add lodash@4.17.21

# 安装为开发依赖
pnpm add -D typescript`,
  publish: `# 发布包
pnpm publish --registry http://localhost:4873`,
  unpublish: `# 删除特定版本
pnpm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873`,
}

// yarn 命令
const yarnCommands = {
  config: `# 设置全局默认源 (Yarn v1)
yarn config set registry http://localhost:4873

# 恢复官方源
yarn config set registry https://registry.npmjs.org`,
  scopeConfig: `# 仅设置特定 scope 的源（推荐）
yarn config set @mycompany:registry http://localhost:4873

# 查看当前配置
yarn config list`,
  configV2: `# Yarn v2+/berry 配置
yarn config set npmRegistryServer http://localhost:4873

# 设置特定 scope 的源
yarn config set npmScopes.mycompany.npmRegistryServer http://localhost:4873`,
  login: `# 登录
yarn login --registry http://localhost:4873

# 登出
yarn logout`,
  install: `# 安装包
yarn add lodash

# 安装指定版本
yarn add lodash@4.17.21

# 安装为开发依赖
yarn add -D typescript`,
  publish: `# 发布包
yarn publish --registry http://localhost:4873

# 发布并指定新版本
yarn publish --new-version 1.0.1`,
  unpublish: `# 删除包
yarn unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873`,
}

// bun 命令（bun 不支持 config set，使用 bunfig.toml 或环境变量）
const bunCommands = {
  config: `# 方式一：使用 bunfig.toml 配置文件（推荐）
# 在项目根目录创建 bunfig.toml
cat > bunfig.toml << 'EOF'
[install]
registry = "http://localhost:4873"
EOF

# 全局配置：~/.bunfig.toml 或 $XDG_CONFIG_HOME/.bunfig.toml

# 方式二：使用环境变量
export BUN_CONFIG_REGISTRY=http://localhost:4873

# 方式三：安装时指定源
bun add lodash --registry http://localhost:4873`,
  scopeConfig: `# 在 bunfig.toml 中设置特定 scope 的源
# bun 目前不支持 scope 级别的 registry 配置
# 建议使用环境变量或安装时指定

# 查看当前配置
cat ~/.bunfig.toml`,
  login: `# bun 使用 npm 登录（凭证需要手动添加到 bunfig.toml）
npm login --registry http://localhost:4873

# 登出
npm logout --registry http://localhost:4873`,
  install: `# 安装包
bun add lodash

# 安装指定版本
bun add lodash@4.17.21

# 安装为开发依赖
bun add -d typescript`,
  publish: `# 发布包
bun publish`,
}

// .npmrc 示例
const npmrcExample = `# .npmrc 文件示例

# 默认源
registry=http://localhost:4873

# 特定 scope 使用私有源
@mycompany:registry=http://localhost:4873
@internal:registry=http://localhost:4873

# 另一个 scope 使用其他源
@partner:registry=https://npm.partner.com`

// package.json 示例
const packageJsonExample = `{
  "name": "@mycompany/my-package",
  "version": "1.0.0",
  "private": false,
  "publishConfig": {
    "registry": "http://localhost:4873"
  }
}`

// 命令速查表
const commandTable = [
  { operation: '设置源', npm: 'npm set registry <url>', pnpm: 'pnpm config set registry <url>', yarn: 'yarn config set registry <url>', bun: '编辑 bunfig.toml 或环境变量' },
  { operation: '登录', npm: 'npm login', pnpm: 'pnpm login', yarn: 'yarn login', bun: '使用 npm login' },
  { operation: '登出', npm: 'npm logout', pnpm: 'pnpm logout', yarn: 'yarn logout', bun: '使用 npm logout' },
  { operation: '安装包', npm: 'npm install <pkg>', pnpm: 'pnpm add <pkg>', yarn: 'yarn add <pkg>', bun: 'bun add <pkg>' },
  { operation: '安装依赖', npm: 'npm install', pnpm: 'pnpm install', yarn: 'yarn', bun: 'bun install' },
  { operation: '发布包', npm: 'npm publish', pnpm: 'pnpm publish', yarn: 'yarn publish', bun: 'bun publish' },
  { operation: '删除包', npm: 'npm unpublish <pkg>', pnpm: 'pnpm unpublish <pkg>', yarn: 'yarn unpublish <pkg>', bun: '-' },
  { operation: '查看配置', npm: 'npm config list', pnpm: 'pnpm config list', yarn: 'yarn config list', bun: 'cat ~/.bunfig.toml' },
]
</script>

<style scoped>
.guide-page {
  max-width: 900px;
  margin: 0 auto;
}

.guide-header {
  text-align: center;
  margin-bottom: 24px;
}

.guide-header h1 {
  margin: 0 0 8px;
  font-size: 28px;
  color: var(--grape-text);
}

.guide-header p {
  margin: 0;
  color: #666;
}

.guide-tabs {
  margin-bottom: 24px;
}

.tip {
  font-size: 13px;
  color: #909399;
  margin: 8px 0 16px;
  padding-left: 12px;
  border-left: 3px solid var(--grape-primary);
}

.section-desc {
  color: #666;
  margin-bottom: 16px;
}

.project-config,
.package-json-config,
.command-table {
  margin-bottom: 24px;
}

.card-header {
  font-size: 16px;
  font-weight: 600;
}

:deep(.el-table .cell) {
  font-size: 13px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
}
</style>
