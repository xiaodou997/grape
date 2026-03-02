<template>
  <div class="home-container">
    <!-- Ultra Modern Hero Section -->
    <section class="hero-section">
      <div class="hero-bg">
        <div class="blob blob-1"></div>
        <div class="blob blob-2"></div>
      </div>
      
      <div class="hero-content animate-slide-up">
        <div class="badge">v0.1.0 is now live</div>
        <h1 class="hero-title">
          {{ t('home.heroSubtitle') }} <br />
          <span class="gradient-text">{{ t('home.heroTitleAccent') }}</span>
        </h1>
        <p class="hero-description">
          {{ t('home.heroTagline') }}
        </p>
        
        <div class="hero-actions">
          <el-button type="primary" size="large" class="btn-primary-glow" @click="$router.push('/packages')">
            {{ t('home.browsePackages') }}
          </el-button>
          <el-button size="large" class="btn-secondary" @click="showSetupGuide = true">
            <el-icon><MagicStick /></el-icon> {{ t('home.quickStart') }}
          </el-button>
        </div>

        <div class="hero-stats">
          <div class="hero-stat-item">
            <span class="stat-value">{{ stats.localPackages }}</span>
            <span class="stat-label">{{ t('home.stats.localPackages') }}</span>
          </div>
          <div class="divider"></div>
          <div class="hero-stat-item">
            <span class="stat-value">{{ stats.storageSize }}<small>MB</small></span>
            <span class="stat-label">{{ t('home.stats.storageSize') }}</span>
          </div>
          <div class="divider"></div>
          <div class="hero-stat-item">
            <span class="stat-value">⚡ Fast</span>
            <span class="stat-label">Response Time</span>
          </div>
        </div>
      </div>
    </section>

    <div class="main-content">
      <!-- Feature Cards -->
      <section class="features-section">
        <h2 class="section-title">{{ t('home.whyGrape') }}</h2>
        <div class="feature-grid">
          <div class="modern-feature-card">
            <div class="feature-icon-wrapper p-purple">
              <el-icon><Monitor /></el-icon>
            </div>
            <h3>{{ t('home.features.adminPanel.title') }}</h3>
            <p>{{ t('home.features.adminPanel.desc') }}</p>
          </div>
          
          <div class="modern-feature-card">
            <div class="feature-icon-wrapper p-green">
              <el-icon><Lock /></el-icon>
            </div>
            <h3>{{ t('home.features.enterpriseSecurity.title') }}</h3>
            <p>{{ t('home.features.enterpriseSecurity.desc') }}</p>
          </div>
          
          <div class="modern-feature-card">
            <div class="feature-icon-wrapper p-blue">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <h3>{{ t('home.features.smartData.title') }}</h3>
            <p>{{ t('home.features.smartData.desc') }}</p>
          </div>
        </div>
      </section>

      <!-- Content Row: Activity & Actions -->
      <section class="dashboard-row">
        <div class="activity-feed">
          <div class="card-header-modern">
            <h3>{{ t('home.recentActivity') }}</h3>
            <el-button text @click="$router.push('/packages')">{{ t('home.viewAll') }}</el-button>
          </div>
          
          <el-table :data="recentPackages" class="modern-table" v-loading="recentLoading">
            <el-table-column prop="name" :label="t('table.package')">
              <template #default="{ row }">
                <div class="pkg-name-cell" @click="$router.push(`/package/${encodeURIComponent(row.name)}`)">
                  <div class="pkg-icon">📦</div>
                  <span>{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="version" :label="t('table.version')" width="100">
              <template #default="{ row }">
                <el-tag size="small" effect="plain" round>{{ row.version }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="updatedAt" :label="t('table.updated')" width="150">
              <template #default="{ row }">
                <span class="time-text">{{ formatTime(row.updatedAt) }}</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!recentLoading && recentPackages.length === 0" :image-size="60" />
        </div>

        <div class="quick-access">
          <h3>{{ t('home.quickActions.title') }}</h3>
          <div class="action-grid">
            <div v-for="action in quickActions" :key="action.path" class="action-tile" @click="$router.push(action.path)">
              <div class="action-tile-icon" :style="{ color: action.color, backgroundColor: action.color + '10' }">
                <el-icon><component :is="action.icon" /></el-icon>
              </div>
              <span>{{ t(action.title) }}</span>
            </div>
          </div>
          
          <div class="promo-card">
            <h4>{{ t('home.promo.needHelp') }}</h4>
            <p>{{ t('home.promo.guideDesc') }}</p>
            <el-button type="primary" size="small" @click="$router.push('/guide')">{{ t('home.promo.viewGuide') }}</el-button>
          </div>
        </div>
      </section>
    </div>

    <!-- Setup Guide Modal -->
    <el-dialog v-model="showSetupGuide" :title="t('home.quickStart')" width="540px" class="modern-dialog">
      <div class="setup-guide-content">
        <div class="guide-step">
          <span class="step-num">01</span>
          <p>{{ t('home.guide.step1') }}</p>
          <div class="code-box">
            <code>npm set registry http://localhost:4873</code>
            <el-icon class="copy-icon"><DocumentCopy /></el-icon>
          </div>
        </div>
        
        <div class="guide-step">
          <span class="step-num">02</span>
          <p>{{ t('home.guide.step2') }}</p>
          <div class="code-box">
            <code>npm install lodash</code>
          </div>
        </div>
        
        <div class="guide-step">
          <span class="step-num">03</span>
          <p>{{ t('home.guide.step3') }}</p>
          <div class="code-box">
            <code>npm publish</code>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { 
  MagicStick, Monitor, Lock, DataAnalysis, Box, 
  User, Key, Download, Setting, DocumentCopy 
} from '@element-plus/icons-vue'
import { adminApi, packageApi } from '@/api'

const { t } = useI18n()
const showSetupGuide = ref(false)
const loading = ref(false)
const recentLoading = ref(false)
const recentPackages = ref<any[]>([])

const stats = ref({
  localPackages: 0,
  cachedPackages: 0,
  storageSize: 0,
})

const loadStats = async () => {
  try {
    const res = await adminApi.getStats()
    stats.value = {
      localPackages: res.data.totalPackages || 0,
      cachedPackages: res.data.cachedPackages || 0,
      storageSize: res.data.storageSize || 0,
    }
  } catch {
    ElMessage.error(t('home.stats.loadError'))
  }
}

const loadRecentPackages = async () => {
  recentLoading.value = true
  try {
    const res = await packageApi.getPackages()
    recentPackages.value = (res.data.packages || [])
      .sort((a: any, b: any) => new Date(b.updatedAt || 0).getTime() - new Date(a.updatedAt || 0).getTime())
      .slice(0, 6)
  } catch {
    // Silent fail
  } finally {
    recentLoading.value = false
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

const quickActions = [
  { title: 'home.quickActions.users', icon: User, path: '/admin/users', color: '#7c3aed' },
  { title: 'home.quickActions.tokens', icon: Key, path: '/admin/tokens', color: '#10b981' },
  { title: 'home.quickActions.backup', icon: Download, path: '/admin/backup', color: '#f59e0b' },
  { title: 'home.quickActions.settings', icon: Setting, path: '/admin/settings', color: '#64748b' },
]

onMounted(() => {
  loadStats()
  loadRecentPackages()
})
</script>

<style scoped>
/* 原有样式保持不变 */
.hero-section {
  position: relative;
  padding: 100px 20px 80px;
  text-align: center;
  overflow: hidden;
  background: white;
}

.hero-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  opacity: 0.4;
}

.blob {
  position: absolute;
  filter: blur(80px);
  border-radius: 50%;
  z-index: 0;
}

.blob-1 {
  width: 400px;
  height: 400px;
  background: rgba(124, 58, 237, 0.2);
  top: -100px;
  right: -100px;
}

.blob-2 {
  width: 300px;
  height: 300px;
  background: rgba(16, 185, 129, 0.1);
  bottom: -50px;
  left: -50px;
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 800px;
  margin: 0 auto;
}

.badge {
  display: inline-block;
  padding: 6px 16px;
  background: var(--g-brand-light);
  color: var(--g-brand);
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 24px;
}

.hero-title {
  font-size: 64px;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -2px;
  color: var(--g-text-primary);
  margin-bottom: 24px;
}

.gradient-text {
  background: linear-gradient(135deg, var(--g-brand), #9333ea);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-description {
  font-size: 20px;
  color: var(--g-text-secondary);
  max-width: 600px;
  margin: 0 auto 40px;
  line-height: 1.6;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin-bottom: 60px;
}

.btn-primary-glow {
  box-shadow: 0 10px 15px -3px rgba(124, 58, 237, 0.3);
  padding: 0 32px !important;
  height: 48px !important;
}

.btn-secondary {
  border: 1px solid var(--g-border) !important;
  padding: 0 32px !important;
  height: 48px !important;
}

.hero-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 40px;
  padding: 24px;
  background: var(--g-bg);
  border-radius: 20px;
  max-width: fit-content;
  margin: 0 auto;
  border: 1px solid var(--g-border);
}

.hero-stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--g-text-primary);
}

.stat-value small {
  font-size: 12px;
  margin-left: 2px;
  color: var(--g-text-muted);
}

.stat-label {
  font-size: 12px;
  color: var(--g-text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.divider {
  width: 1px;
  height: 30px;
  background: var(--g-border);
}

.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 80px 20px;
}

.section-title {
  text-align: center;
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 48px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 32px;
  margin-bottom: 80px;
}

.modern-feature-card {
  padding: 40px;
  background: white;
  border-radius: 24px;
  border: 1px solid var(--g-border);
  transition: all 0.3s ease;
}

.modern-feature-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-lg);
}

.feature-icon-wrapper {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-bottom: 24px;
}

.p-purple { background: #f5f3ff; color: #7c3aed; }
.p-green { background: #ecfdf5; color: #10b981; }
.p-blue { background: #eff6ff; color: #3b82f6; }

.modern-feature-card h3 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 12px;
}

.modern-feature-card p {
  color: var(--g-text-secondary);
  line-height: 1.6;
}

.dashboard-row {
  display: grid;
  grid-template-columns: 1.8fr 1fr;
  gap: 32px;
}

.activity-feed, .quick-access {
  background: white;
  padding: 32px;
  border-radius: 24px;
  border: 1px solid var(--g-border);
}

.card-header-modern {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.card-header-modern h3 {
  font-size: 20px;
  font-weight: 700;
}

.pkg-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  font-weight: 500;
  color: var(--g-text-primary);
}

.pkg-name-cell:hover {
  color: var(--g-brand);
}

.pkg-icon {
  font-size: 18px;
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin: 24px 0;
}

.action-tile {
  padding: 20px;
  background: var(--g-bg);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-tile:hover {
  background: white;
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.action-tile-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.action-tile span {
  font-size: 13px;
  font-weight: 600;
  color: var(--g-text-secondary);
}

.promo-card {
  margin-top: 32px;
  padding: 24px;
  background: linear-gradient(135deg, var(--g-brand), #9333ea);
  border-radius: 16px;
  color: white;
}

.promo-card h4 {
  font-size: 18px;
  margin-bottom: 8px;
}

.promo-card p {
  font-size: 13px;
  opacity: 0.9;
  margin-bottom: 16px;
}

.setup-guide-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.guide-step {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-num {
  font-size: 12px;
  font-weight: 800;
  color: var(--g-brand);
  letter-spacing: 1px;
}

.code-box {
  background: #0f172a;
  padding: 16px;
  border-radius: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.code-box code {
  color: #e2e8f0;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
}

.copy-icon {
  color: #94a3b8;
  cursor: pointer;
}

.copy-icon:hover {
  color: white;
}

/* Responsive fixes */
@media (max-width: 1024px) {
  .dashboard-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 40px;
  }
  .hero-stats {
    flex-direction: column;
    gap: 20px;
  }
  .divider {
    width: 60px;
    height: 1px;
  }
}
</style>
