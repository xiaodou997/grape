<template>
  <div class="guide-page fade-in">
    <el-card class="guide-header">
      <h1>📖 {{ $t('guide.title') }}</h1>
      <p>{{ $t('guide.subtitle') }}</p>
    </el-card>

    <!-- Tab 切换不同包管理器 -->
    <el-tabs v-model="activeTab" class="guide-tabs">
      <!-- npm -->
      <el-tab-pane label="npm" name="npm">
        <GuideSection :title="$t('guide.configTitle')">
          <CodeBlock :code="npmCommands.config" />
          <p class="tip">{{ $t('guide.npm.scopeConfig') }}</p>
          <CodeBlock :code="npmCommands.scopeConfig" />
        </GuideSection>

        <GuideSection :title="$t('guide.authTitle')">
          <CodeBlock :code="npmCommands.login" />
          <p class="tip">{{ $t('guide.npm.defaultCreds') }}</p>
        </GuideSection>

        <GuideSection :title="$t('guide.installTitle')">
          <CodeBlock :code="npmCommands.install" />
        </GuideSection>

        <GuideSection :title="$t('guide.publishTitle')">
          <CodeBlock :code="npmCommands.publish" />
          <p class="tip">{{ $t('guide.npm.publishConfigHint') }}</p>
        </GuideSection>

        <GuideSection :title="$t('guide.unpublishTitle')">
          <CodeBlock :code="npmCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- pnpm -->
      <el-tab-pane label="pnpm" name="pnpm">
        <GuideSection :title="$t('guide.configTitle')">
          <CodeBlock :code="pnpmCommands.config" />
          <p class="tip">{{ $t('guide.pnpm.scopeConfig') }}</p>
          <CodeBlock :code="pnpmCommands.scopeConfig" />
        </GuideSection>

        <GuideSection :title="$t('guide.authTitle')">
          <CodeBlock :code="pnpmCommands.login" />
        </GuideSection>

        <GuideSection :title="$t('guide.installTitle')">
          <CodeBlock :code="pnpmCommands.install" />
        </GuideSection>

        <GuideSection :title="$t('guide.publishTitle')">
          <CodeBlock :code="pnpmCommands.publish" />
        </GuideSection>

        <GuideSection :title="$t('guide.unpublishTitle')">
          <CodeBlock :code="pnpmCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- yarn -->
      <el-tab-pane label="yarn" name="yarn">
        <GuideSection :title="$t('guide.yarn.configV1Title')">
          <CodeBlock :code="yarnCommands.config" />
          <p class="tip">{{ $t('guide.yarn.scopeConfig') }}</p>
          <CodeBlock :code="yarnCommands.scopeConfig" />
        </GuideSection>

        <GuideSection :title="$t('guide.yarn.configV2Title')">
          <CodeBlock :code="yarnCommands.configV2" />
        </GuideSection>

        <GuideSection :title="$t('guide.authTitle')">
          <CodeBlock :code="yarnCommands.login" />
        </GuideSection>

        <GuideSection :title="$t('guide.installTitle')">
          <CodeBlock :code="yarnCommands.install" />
        </GuideSection>

        <GuideSection :title="$t('guide.publishTitle')">
          <CodeBlock :code="yarnCommands.publish" />
        </GuideSection>

        <GuideSection :title="$t('guide.unpublishTitle')">
          <CodeBlock :code="yarnCommands.unpublish" />
        </GuideSection>
      </el-tab-pane>

      <!-- bun -->
      <el-tab-pane label="bun" name="bun">
        <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
          <template #title>
            <strong>{{ $t('guide.bun.warning') }}</strong>：{{ $t('guide.bun.warningDesc') }}
          </template>
        </el-alert>

        <GuideSection :title="$t('guide.configTitle')">
          <CodeBlock :code="bunCommands.config" />
          <p class="tip">bun {{ $t('guide.bun.configProject') }}</p>
          <CodeBlock :code="bunCommands.scopeConfig" />
        </GuideSection>

        <GuideSection :title="$t('guide.authTitle')">
          <CodeBlock :code="bunCommands.login" />
          <p class="tip">{{ $t('guide.npm.defaultCreds') }}</p>
        </GuideSection>

        <GuideSection :title="$t('guide.installTitle')">
          <CodeBlock :code="bunCommands.install" />
        </GuideSection>

        <GuideSection :title="$t('guide.publishTitle')">
          <CodeBlock :code="bunCommands.publish" />
        </GuideSection>
      </el-tab-pane>
    </el-tabs>

    <!-- 项目级配置 -->
    <el-card class="project-config">
      <template #header>
        <div class="card-header">
          <span>📁 {{ $t('guide.projectConfig') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ $t('guide.projectConfigDesc') }}</p>
      <CodeBlock :code="npmrcExample" language="ini" />
    </el-card>

    <!-- package.json 配置 -->
    <el-card class="package-json-config">
      <template #header>
        <div class="card-header">
          <span>📦 {{ $t('guide.packageJson') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ $t('guide.packageJsonDesc') }}</p>
      <CodeBlock :code="packageJsonExample" language="json" />
    </el-card>

    <!-- 命令速查表 -->
    <el-card class="command-table">
      <template #header>
        <div class="card-header">
          <span>📋 {{ $t('guide.commandTable') }}</span>
        </div>
      </template>
      <el-table :data="commandTable" stripe>
        <el-table-column prop="operation" :label="$t('guide.operation')" width="120" />
        <el-table-column prop="npm" label="npm" />
        <el-table-column prop="pnpm" label="pnpm" />
        <el-table-column prop="yarn" label="yarn" />
        <el-table-column prop="bun" label="bun" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import GuideSection from '@/components/GuideSection.vue'
import CodeBlock from '@/components/CodeBlock.vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const activeTab = ref('npm')

// npm 命令
const npmCommands = computed(() => ({
  config: `# ${t('guide.npm.configDesc')}
npm set registry http://localhost:4873

# ${t('guide.npm.configRestore')}
npm set registry https://registry.npmjs.org`,
  scopeConfig: `# ${t('guide.npm.scopeConfig')}
npm set @mycompany:registry http://localhost:4873

# ${t('guide.npm.configList')}
npm config list`,
  login: `# ${t('guide.npm.login')}
npm login --registry http://localhost:4873

# ${t('guide.npm.logout')}
npm logout --registry http://localhost:4873`,
  install: `# ${t('guide.npm.install')}
npm install lodash

# ${t('guide.npm.installVersion')}
npm install lodash@4.17.21

# ${t('guide.npm.installDev')}
npm install -D typescript`,
  publish: `# ${t('guide.npm.publish')}
npm publish --registry http://localhost:4873

# ${t('guide.npm.publishBeta')}
npm publish --tag beta`,
  unpublish: `# ${t('guide.npm.unpublishVersion')}
npm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873

# ${t('guide.npm.unpublishAll')}
npm unpublish @mycompany/my-package --force --registry http://localhost:4873`,
}))

// pnpm 命令
const pnpmCommands = computed(() => ({
  config: `# ${t('guide.pnpm.configDesc')}
pnpm config set registry http://localhost:4873

# ${t('guide.pnpm.configRestore')}
pnpm config set registry https://registry.npmjs.org`,
  scopeConfig: `# ${t('guide.pnpm.scopeConfig')}
pnpm config set @mycompany:registry http://localhost:4873

# ${t('guide.pnpm.configList')}
pnpm config list`,
  login: `# ${t('guide.npm.login')}
pnpm login --registry http://localhost:4873

# ${t('guide.npm.logout')}
pnpm logout --registry http://localhost:4873`,
  install: `# ${t('guide.npm.install')}
pnpm add lodash

# ${t('guide.npm.installVersion')}
pnpm add lodash@4.17.21

# ${t('guide.npm.installDev')}
pnpm add -D typescript`,
  publish: `# ${t('guide.npm.publish')}
pnpm publish --registry http://localhost:4873`,
  unpublish: `# ${t('guide.npm.unpublishVersion')}
pnpm unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873`,
}))

// yarn 命令
const yarnCommands = computed(() => ({
  config: `# ${t('guide.yarn.configV1Title')}
yarn config set registry http://localhost:4873

# ${t('guide.pnpm.configRestore')}
yarn config set registry https://registry.npmjs.org`,
  scopeConfig: `# ${t('guide.yarn.scopeConfig')}
yarn config set @mycompany:registry http://localhost:4873

# ${t('guide.pnpm.configList')}
yarn config list`,
  configV2: `# ${t('guide.yarn.configV2')}
yarn config set npmRegistryServer http://localhost:4873

# ${t('guide.yarn.scopeConfigV2')}
yarn config set npmScopes.mycompany.npmRegistryServer http://localhost:4873`,
  login: `# ${t('guide.npm.login')}
yarn login --registry http://localhost:4873

# ${t('guide.npm.logout')}
yarn logout`,
  install: `# ${t('guide.npm.install')}
yarn add lodash

# ${t('guide.npm.installVersion')}
yarn add lodash@4.17.21

# ${t('guide.npm.installDev')}
yarn add -D typescript`,
  publish: `# ${t('guide.npm.publish')}
yarn publish --registry http://localhost:4873

# ${t('guide.npm.publishBeta')}
yarn publish --new-version 1.0.1`,
  unpublish: `# ${t('guide.npm.unpublishVersion')}
yarn unpublish @mycompany/my-package@1.0.0 --registry http://localhost:4873`,
}))

// bun 命令
const bunCommands = computed(() => ({
  config: `# ${t('guide.bun.configMethod1')}
# ${t('guide.bun.configProject')}
cat > bunfig.toml << 'EOF'
[install]
registry = "http://localhost:4873"
EOF

# ${t('guide.bun.configGlobal')}

# ${t('guide.bun.configMethod2')}
export BUN_CONFIG_REGISTRY=http://localhost:4873

# ${t('guide.bun.configMethod3')}
bun add lodash --registry http://localhost:4873`,
  scopeConfig: `# ${t('guide.bun.noScopeConfig')}
# ${t('guide.bun.viewConfig')}
cat ~/.bunfig.toml`,
  login: `# bun ${t('guide.npm.login')}
npm login --registry http://localhost:4873

# ${t('guide.npm.logout')}
npm logout --registry http://localhost:4873`,
  install: `# ${t('guide.npm.install')}
bun add lodash

# ${t('guide.npm.installVersion')}
bun add lodash@4.17.21

# ${t('guide.npm.installDev')}
bun add -d typescript`,
  publish: `# ${t('guide.npm.publish')}
bun publish`,
}))

// .npmrc 示例
const npmrcExample = computed(() => `# .npmrc

# ${t('guide.npm.configDesc')}
registry=http://localhost:4873

# ${t('guide.npm.scopeConfig')}
@mycompany:registry=http://localhost:4873
@internal:registry=http://localhost:4873

# ${t('guide.yarn.scopeConfig')}
@partner:registry=https://npm.partner.com`)

// package.json 示例
const packageJsonExample = computed(() => `{
  "name": "@mycompany/my-package",
  "version": "1.0.0",
  "private": false,
  "publishConfig": {
    "registry": "http://localhost:4873"
  }
}`)

// 命令速查表
const commandTable = computed(() => [
  { operation: t('guide.setRegistry'), npm: 'npm set registry <url>', pnpm: 'pnpm config set registry <url>', yarn: 'yarn config set registry <url>', bun: t('guide.bun.configMethod1') },
  { operation: t('guide.restoreRegistry'), npm: 'npm set registry https://registry.npmjs.org', pnpm: 'pnpm config set registry https://registry.npmjs.org', yarn: 'yarn config set registry https://registry.npmjs.org', bun: 'Edit ~/.bunfig.toml' },
  { operation: t('guide.npm.login'), npm: 'npm login', pnpm: 'pnpm login', yarn: 'yarn login', bun: 'npm login' },
  { operation: t('guide.npm.logout'), npm: 'npm logout', pnpm: 'pnpm logout', yarn: 'yarn logout', bun: 'npm logout' },
  { operation: t('guide.npm.install'), npm: 'npm install <pkg>', pnpm: 'pnpm add <pkg>', yarn: 'yarn add <pkg>', bun: 'bun add <pkg>' },
  { operation: t('guide.npm.installDev'), npm: 'npm install -D <pkg>', pnpm: 'pnpm add -D <pkg>', yarn: 'yarn add -D <pkg>', bun: 'bun add -d <pkg>' },
  { operation: t('guide.npm.publish'), npm: 'npm publish', pnpm: 'pnpm publish', yarn: 'yarn publish', bun: 'bun publish' },
  { operation: t('guide.npm.unpublishVersion'), npm: 'npm unpublish <pkg>@<version>', pnpm: 'pnpm unpublish <pkg>@<version>', yarn: 'yarn unpublish <pkg>@<version>', bun: '-' },
  { operation: t('guide.pnpm.configList'), npm: 'npm config list', pnpm: 'pnpm config list', yarn: 'yarn config list', bun: 'cat ~/.bunfig.toml' },
])
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