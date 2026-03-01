<template>
  <div class="home-page fade-in">
    <!-- Hero Section -->
    <section class="hero">
      <div class="hero-content">
        <h1 class="hero-title">
          <span class="grape-icon">🍇</span>
          Grape
        </h1>
        <p class="hero-subtitle">{{ t('home.heroSubtitle') }}</p>
        <p class="hero-tagline">{{ t('home.heroTagline') }}</p>
        
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/packages')">
            {{ t('home.browsePackages') }}
          </el-button>
          <el-button size="large" @click="showSetupGuide = true">
            {{ t('home.quickStart') }}
          </el-button>
        </div>
      </div>
    </section>

    <!-- Stats Section -->
    <section class="stats" v-loading="loading">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic :title="t('home.stats.localPackages')" :value="stats.localPackages">
              <template #prefix>
                <el-icon><Box /></el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic :title="t('home.stats.cachedPackages')" :value="stats.cachedPackages">
              <template #prefix>
                <el-icon><Download /></el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <el-statistic :title="t('home.stats.storageSize')" :value="stats.storageSize" suffix="MB">
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
      <h2 class="section-title">{{ t('home.whyGrape') }}</h2>
      <el-row :gutter="24">
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><Monitor /></el-icon>
                <span>{{ t('home.features.adminPanel.title') }}</span>
              </div>
            </template>
            <p>{{ t('home.features.adminPanel.desc') }}</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><Lock /></el-icon>
                <span>{{ t('home.features.enterpriseSecurity.title') }}</span>
              </div>
            </template>
            <p>{{ t('home.features.enterpriseSecurity.desc') }}</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <template #header>
              <div class="feature-header">
                <el-icon :size="32" color="var(--grape-primary)"><DataAnalysis /></el-icon>
                <span>{{ t('home.features.smartData.title') }}</span>
              </div>
            </template>
            <p>{{ t('home.features.smartData.desc') }}</p>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <!-- Recent Activity & Quick Actions -->
    <section class="activity-section">
      <el-row :gutter="24">
        <!-- Recent Packages -->
        <el-col :span="16">
          <el-card v-loading="recentLoading">
            <template #header>
              <div class="card-header">
                <span>{{ t('home.recentActivity') }}</span>
                <el-button text type="primary" @click="$router.push('/packages')">
                  {{ t('home.viewAll') }}
                </el-button>
              </div>
            </template>
            <el-table :data="recentPackages" style="width: 100%" :show-header="true">
              <el-table-column prop="name" :label="t('packages.name')" min-width="200">
                <template #default="{ row }">
                  <el-link type="primary" @click="$router.push(`/package/${encodeURIComponent(row.name)}`)">
                    {{ row.name }}
                  </el-link>
                </template>
              </el-table-column>
              <el-table-column prop="version" :label="t('packages.version')" width="120" />
              <el-table-column prop="updatedAt" :label="t('packages.lastModified')" width="150">
                <template #default="{ row }">
                  {{ formatTime(row.updatedAt) }}
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!recentLoading && recentPackages.length === 0" :description="t('packages.noPackages')" />
          </el-card>
        </el-col>

        <!-- Quick Actions -->
        <el-col :span="8">
          <el-card>
            <template #header>
              <span>{{ t('home.quickActions.title') }}</span>
            </template>
            <div class="quick-actions">
              <div
                v-for="action in quickActions"
                :key="action.path"
                class="quick-action-item"
                @click="$router.push(action.path)"
              >
                <div class="action-icon" :style="{ backgroundColor: action.color + '20', color: action.color }">
                  <el-icon :size="20">
                    <component :is="action.icon" />
                  </el-icon>
                </div>
                <span class="action-title">{{ t(action.title) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <!-- Setup Guide Dialog -->
    <el-dialog v-model="showSetupGuide" :title="t('home.quickStart')" width="600px">
      <div class="setup-guide">
        <h4>{{ t('home.guide.step1') }}</h4>
        <div class="code-block">
          <code>npm set registry http://localhost:4873</code>
        </div>
        
        <h4>{{ t('home.guide.step2') }}</h4>
        <div class="code-block">
          <code>npm install lodash</code>
        </div>
        
        <h4>{{ t('home.guide.step3') }}</h4>
        <div class="code-block">
          <code>npm publish --registry http://localhost:4873</code>
        </div>
        
        <h4>{{ t('home.guide.step4') }}</h4>
        <div class="code-block">
          <code>npm set registry https://registry.npmjs.org</code>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Box, Download, Folder, Monitor, Lock, DataAnalysis } from '@element-plus/icons-vue'
import { adminApi, packageApi } from '@/api'

const { t } = useI18n()
const showSetupGuide = ref(false)
const loading = ref(false)

const stats = ref({
  localPackages: 0,
  cachedPackages: 0,
  storageSize: 0,
})

const loadStats = async () => {
  loading.value = true
  try {
    const res = await adminApi.getStats()
    const data = res.data
    stats.value = {
      localPackages: data.totalPackages || 0,
      cachedPackages: data.cachedPackages || data.totalPackages || 0,
      storageSize: data.storageSize || 0, // API 已转换为 MB
    }
  } catch {
    ElMessage.error(t('home.stats.loadError'))
  } finally {
    loading.value = false
  }
}

// 最近活动
const recentPackages = ref<any[]>([])
const recentLoading = ref(false)

const loadRecentPackages = async () => {
  recentLoading.value = true
  try {
    const res = await packageApi.getPackages()
    // 取最近更新的 5 个包
    recentPackages.value = (res.data.packages || [])
      .sort((a: any, b: any) => new Date(b.updatedAt || 0).getTime() - new Date(a.updatedAt || 0).getTime())
      .slice(0, 5)
  } catch {
    // 静默失败，不显示错误
  } finally {
    recentLoading.value = false
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleDateString()
}

// 快捷入口
const quickActions = [
  { title: 'home.quickActions.users', icon: 'User', path: '/admin/users', color: '#409EFF' },
  { title: 'home.quickActions.tokens', icon: 'Key', path: '/admin/tokens', color: '#67C23A' },
  { title: 'home.quickActions.backup', icon: 'Download', path: '/admin/backup', color: '#E6A23C' },
  { title: 'home.quickActions.settings', icon: 'Setting', path: '/admin/settings', color: '#909399' },
]

onMounted(() => {
  loadStats()
  loadRecentPackages()
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

.activity-section {
  margin-bottom: 40px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.quick-action-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.quick-action-item:hover {
  background-color: #f5f7fa;
}

.action-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-title {
  font-size: 14px;
  color: #606266;
}
</style>