<template>
  <div class="home-page fade-in">
    <!-- Hero Section -->
    <section class="hero">
      <div class="hero-content">
        <h1 class="hero-title">
          <span class="grape-icon">🍇</span>
          Grape
        </h1>
        <p class="hero-subtitle">轻盈如风的企业级私有 npm 仓库</p>
        <p class="hero-tagline">One binary, zero debt. 一个二进制，零负担。</p>
        
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/packages')">
            浏览包
          </el-button>
          <el-button size="large" @click="showSetupGuide = true">
            快速开始
          </el-button>
        </div>
      </div>
    </section>

    <!-- Stats Section -->
    <section class="stats">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic title="本地包数量" :value="stats.localPackages">
              <template #prefix>
                <el-icon><Box /></el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic title="缓存包数量" :value="stats.cachedPackages">
              <template #prefix>
                <el-icon><Download /></el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic title="存储占用" :value="stats.storageSize" suffix="MB">
              <template #prefix>
                <el-icon><Folder /></el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <!-- Features Section -->
    <section class="features">
      <h2 class="section-title">为什么选择 Grape？</h2>
      <el-row :gutter="24">
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><Promotion /></el-icon>
                <span>极速部署</span>
              </div>
            </template>
            <p>单一二进制文件，无需安装 Node.js，无需 npm install。下载即用，开箱即部署。</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><Lock /></el-icon>
                <span>安全可控</span>
              </div>
            </template>
            <p>私有包完全隔离，精细的权限控制，完整的审计日志，企业级安全保障。</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><Connection /></el-icon>
                <span>智能代理</span>
              </div>
            </template>
            <p>自动缓存公共包，加速团队开发。支持多上游源，灵活配置。</p>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <!-- Setup Guide Dialog -->
    <el-dialog v-model="showSetupGuide" title="快速开始" width="600px">
      <div class="setup-guide">
        <h4>1. 配置 npm 使用 Grape</h4>
        <div class="code-block">
          <code>npm set registry http://localhost:4873</code>
        </div>
        
        <h4>2. 安装包</h4>
        <div class="code-block">
          <code>npm install lodash</code>
        </div>
        
        <h4>3. 发布私有包</h4>
        <div class="code-block">
          <code>npm publish --registry http://localhost:4873</code>
        </div>
        
        <h4>4. 恢复默认源</h4>
        <div class="code-block">
          <code>npm set registry https://registry.npmjs.org</code>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Box, Download, Folder, Promotion, Lock, Connection } from '@element-plus/icons-vue'

const showSetupGuide = ref(false)

const stats = ref({
  localPackages: 0,
  cachedPackages: 5,
  storageSize: 2.5,
})
</script>

<style scoped>
.home-page {
  max-width: 1200px;
  margin: 0 auto;
}

.hero {
  text-align: center;
  padding: 60px 20px;
  background: linear-gradient(135deg, #f8f5ff 0%, #fff 100%);
  border-radius: 16px;
  margin-bottom: 40px;
}

.hero-content {
  max-width: 600px;
  margin: 0 auto;
}

.hero-title {
  font-size: 48px;
  font-weight: 700;
  color: var(--grape-primary);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.grape-icon {
  font-size: 56px;
}

.hero-subtitle {
  font-size: 24px;
  color: var(--grape-text);
  margin-bottom: 8px;
}

.hero-tagline {
  font-size: 16px;
  color: #666;
  margin-bottom: 32px;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.stats {
  margin-bottom: 40px;
}

.stat-card {
  text-align: center;
}

.section-title {
  text-align: center;
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 32px;
  color: var(--grape-text);
}

.features {
  margin-bottom: 40px;
}

.feature-card {
  height: 100%;
}

.feature-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 600;
}

.setup-guide h4 {
  margin: 20px 0 8px;
  color: var(--grape-text);
}

.setup-guide h4:first-child {
  margin-top: 0;
}

.setup-guide .code-block {
  margin-bottom: 12px;
}
</style>