<template>
  <div class="dashboard-wrapper fade-in">
    <!-- Top Stats Grid -->
    <div class="stats-grid">
      <div class="stat-box blue">
        <div class="stat-info">
          <span class="stat-label">{{ t('home.stats.localPackages') }}</span>
          <span class="stat-number">{{ stats.totalPackages }}</span>
        </div>
        <div class="stat-icon-bg"><el-icon><Box /></el-icon></div>
      </div>

      <div class="stat-box green">
        <div class="stat-info">
          <span class="stat-label">{{ t('home.stats.storageSize') }}</span>
          <span class="stat-number">{{ formatSize(stats.storageSize) }}</span>
        </div>
        <div class="stat-icon-bg"><el-icon><Download /></el-icon></div>
      </div>

      <div class="stat-box orange">
        <div class="stat-info">
          <span class="stat-label">{{ t('nav.users') }}</span>
          <span class="stat-number">{{ stats.users }}</span>
        </div>
        <div class="stat-icon-bg"><el-icon><User /></el-icon></div>
      </div>

      <div class="stat-box purple">
        <div class="stat-info">
          <span class="stat-label">{{ t('common.status') }}</span>
          <span class="stat-number status-text">
            <span class="status-dot"></span>
            {{ t('common.running') }}
          </span>
        </div>
        <div class="stat-icon-bg"><el-icon><Check /></el-icon></div>
      </div>
    </div>

    <!-- Main Content Split -->
    <div class="dashboard-main-grid">
      <!-- Left: Recent Activity -->
      <section class="dashboard-card activity-card">
        <div class="card-header-v2">
          <h3>{{ t('home.recentActivity') }}</h3>
          <el-button text type="primary" @click="$router.push('/packages')">{{ t('home.viewAll') }}</el-button>
        </div>
        <div class="activity-list" v-loading="loading">
          <el-empty v-if="recentPackages.length === 0" :image-size="80" :description="t('packages.empty')" />
          <div v-else class="timeline-v2">
            <div v-for="pkg in recentPackages" :key="pkg.name" class="timeline-v2-item" @click="$router.push(`/package/${pkg.name}`)">
              <div class="t-dot"></div>
              <div class="t-content">
                <div class="t-header">
                  <span class="t-pkg">{{ pkg.name }}</span>
                  <span class="t-ver">v{{ pkg.version }}</span>
                </div>
                <div class="t-time">{{ formatTime(pkg.updatedAt) }}</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Right: Quick Actions & System Info -->
      <div class="dashboard-side-col">
        <section class="dashboard-card quick-links">
          <h3>{{ t('home.quickActions.title') }}</h3>
          <div class="action-grid-small">
            <div v-for="action in quickActions" :key="action.path" class="action-item-small" @click="$router.push(action.path)">
              <div class="item-icon" :style="{ color: action.color, backgroundColor: action.color + '10' }">
                <el-icon><component :is="action.icon" /></el-icon>
              </div>
              <span>{{ action.name }}</span>
            </div>
          </div>
        </section>

        <section class="dashboard-card system-info-card">
          <h3>{{ t('admin.systemInfo') }}</h3>
          <div class="sys-detail-list">
            <div class="sys-item">
              <span class="s-label">{{ t('admin.version') }}</span>
              <span class="s-value"><el-tag size="small" round effect="light">{{ sysInfo.version || 'v0.1.0' }}</el-tag></span>
            </div>
            <div class="sys-item">
              <span class="s-label">{{ t('admin.uptime') }}</span>
              <span class="s-value">{{ sysInfo.uptime || '-' }}</span>
            </div>
            <div class="sys-item">
              <span class="s-label">{{ t('admin.storagePath') }}</span>
              <span class="s-value truncate">{{ sysInfo.storagePath || '-' }}</span>
            </div>
            <div class="sys-item">
              <span class="s-label">Database</span>
              <span class="s-value truncate">{{ sysInfo.databasePath || '-' }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Box, Download, User, Check, UserFilled, Key, Delete, Setting, Connection } from '@element-plus/icons-vue'
import { adminApi, packageApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const stats = reactive({ totalPackages: 0, storageSize: 0, users: 0 })
const sysInfo = reactive({ version: '', uptime: '', host: '', storagePath: '', databasePath: '' })
const recentPackages = ref<any[]>([])

const quickActions = [
  { name: t('nav.users'), icon: UserFilled, color: '#7c3aed', path: '/admin/users' },
  { name: t('nav.tokens'), icon: Key, color: '#10b981', path: '/admin/tokens' },
  { name: t('nav.backup'), icon: Download, color: '#f59e0b', path: '/admin/backup' },
  { name: t('nav.gc'), icon: Delete, color: '#ef4444', path: '/admin/gc' },
  { name: t('nav.settings'), icon: Setting, color: '#64748b', path: '/admin/settings' },
  { name: t('nav.webhooks'), icon: Connection, color: '#3b82f6', path: '/admin/webhooks' },
]

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

const formatTime = (time?: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const loadData = async () => {
  loading.value = true
  try {
    const [statsRes, usersRes, pkgRes, sysRes] = await Promise.all([
      adminApi.getStats(),
      adminApi.getUsers(),
      packageApi.getPackages(),
      adminApi.getSystemInfo()
    ])
    stats.totalPackages = statsRes.data.totalPackages || 0
    stats.storageSize = statsRes.data.storageSize || 0
    stats.users = (usersRes.data.users || []).length
    recentPackages.value = (pkgRes.data.packages || []).slice(0, 8)
    Object.assign(sysInfo, sysRes.data)
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.dashboard-wrapper { display: flex; flex-direction: column; gap: 32px; }
.stats-grid { 
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.stat-box { 
  flex: 1 1 200px;
  min-width: 0;
  position: relative; 
  padding: 24px; 
  background: white; 
  border-radius: 20px; 
  border: 1px solid var(--g-border); 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  overflow: hidden; 
  transition: all 0.2s ease; 
}

@media (max-width: 1400px) {
  .stat-number { font-size: 22px; }
  .stat-label { font-size: 11px; }
  .stat-box { padding: 16px; }
}

@media (max-width: 900px) {
  .stat-box { flex-basis: calc(50% - 8px); }
}

@media (max-width: 600px) {
  .stat-box { flex-basis: 100%; }
}

.stat-box:hover { transform: translateY(-2px); border-color: var(--g-brand); box-shadow: var(--shadow-md); }
.stat-info { display: flex; flex-direction: column; z-index: 1; }
.stat-label { font-size: 12px; font-weight: 600; color: var(--g-text-muted); text-transform: uppercase; margin-bottom: 8px; }
.stat-number { font-size: 26px; font-weight: 800; color: var(--g-text-primary); }
.stat-icon-bg { font-size: 44px; opacity: 0.08; position: absolute; right: -10px; bottom: -10px; transform: rotate(-15deg); }

.status-text { display: flex; align-items: center; gap: 8px; font-size: 18px; }
.status-dot { width: 10px; height: 10px; background: var(--g-success); border-radius: 50%; box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.1); animation: pulse 2s infinite; }
@keyframes pulse { 0% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4); } 70% { box-shadow: 0 0 0 8px rgba(16, 185, 129, 0); } 100% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); } }

.dashboard-main-grid { display: grid; grid-template-columns: 1.5fr 1fr; gap: 24px; }
.dashboard-card { background: white; border-radius: 24px; border: 1px solid var(--g-border); padding: 28px; }
.dashboard-card h3 { font-size: 18px; font-weight: 700; margin-bottom: 24px; }

.card-header-v2 { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.timeline-v2 { display: flex; flex-direction: column; gap: 8px; }
.timeline-v2-item { padding: 12px 16px; background: var(--g-bg); border-radius: 12px; display: flex; gap: 12px; cursor: pointer; transition: all 0.2s; border: 1px solid transparent; }
.timeline-v2-item:hover { background: white; border-color: var(--g-brand); transform: translateX(4px); }
.t-dot { width: 6px; height: 6px; background: var(--g-brand); border-radius: 50%; margin-top: 6px; }
.t-header { display: flex; justify-content: space-between; margin-bottom: 2px; }
.t-pkg { font-weight: 600; font-size: 14px; color: var(--g-text-primary); }
.t-ver { font-size: 12px; color: var(--g-text-muted); }
.t-time { font-size: 12px; color: var(--g-text-muted); }

.dashboard-side-col { display: flex; flex-direction: column; gap: 24px; }
.action-grid-small { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
.action-item-small { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 12px; cursor: pointer; border-radius: 12px; transition: all 0.2s; }
.action-item-small:hover { background: var(--g-bg); }
.item-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.action-item-small span { font-size: 12px; font-weight: 600; color: var(--g-text-secondary); text-align: center; }

.sys-detail-list { display: flex; flex-direction: column; gap: 12px; }
.sys-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; background: var(--g-bg); border-radius: 10px; }
.s-label { font-size: 13px; color: var(--g-text-muted); }
.s-value { font-size: 13px; font-weight: 600; color: var(--g-text-primary); }
.truncate { max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

@media (max-width: 1024px) { .dashboard-main-grid { grid-template-columns: 1fr; } }
</style>
