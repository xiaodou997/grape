<template>
  <div class="dashboard-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" v-loading="loading">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon blue">
            <el-icon :size="32"><Box /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalPackages }}</div>
            <div class="stat-label">{{ $t('home.stats.localPackages') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon green">
            <el-icon :size="32"><Download /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatSize(stats.storageSize) }}</div>
            <div class="stat-label">{{ $t('home.stats.storageSize') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon orange">
            <el-icon :size="32"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.users }}</div>
            <div class="stat-label">{{ $t('nav.users') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon purple">
            <el-icon :size="32"><Check /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ $t('common.running') }}</div>
            <div class="stat-label">{{ $t('common.status') }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷操作 -->
    <el-card class="quick-actions" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="header-title">
            <el-icon><Grid /></el-icon>
            {{ $t('home.quickActions.title') }}
          </span>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :span="4" v-for="action in quickActions" :key="action.name">
          <div class="quick-action-item" @click="$router.push(action.path)">
            <el-icon :size="24" :color="action.color">
              <component :is="action.icon" />
            </el-icon>
            <span class="action-name">{{ action.name }}</span>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 最近活动和系统信息 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="hover" class="recent-activity">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><Clock /></el-icon>
                {{ $t('home.recentActivity') }}
              </span>
              <el-button text type="primary" @click="$router.push('/packages')">
                {{ $t('home.viewAll') }}
              </el-button>
            </div>
          </template>
          <el-empty v-if="recentPackages.length === 0" :description="$t('packages.empty')" />
          <el-timeline v-else>
            <el-timeline-item
              v-for="pkg in recentPackages"
              :key="pkg.name"
              :timestamp="formatTime(pkg.updatedAt)"
              type="primary"
            >
              <div class="timeline-item" @click="$router.push(`/package/${pkg.name}`)">
                <el-link type="primary">{{ pkg.name }}</el-link>
                <el-tag size="small" type="info">v{{ pkg.version }}</el-tag>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="system-overview">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><InfoFilled /></el-icon>
                {{ $t('admin.systemInfo') }}
              </span>
              <el-button text type="primary" @click="$router.push('/admin/system')">
                {{ $t('common.more') }}
              </el-button>
            </div>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="$t('admin.version')">{{ sysInfo.version }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.uptime')">{{ sysInfo.uptime }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.host')">{{ sysInfo.host }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.storagePath')">{{ sysInfo.storagePath }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Box,
  Download,
  User,
  Check,
  Grid,
  Clock,
  InfoFilled,
  UserFilled,
  Key,
  Download as DownloadIcon,
  Delete,
  Setting,
  Link,
  Connection,
  Document,
} from '@element-plus/icons-vue'
import { adminApi, packageApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)

const stats = reactive({
  totalPackages: 0,
  storageSize: 0,
  users: 0,
})

const sysInfo = reactive({
  version: '',
  uptime: '',
  host: '',
  storagePath: '',
})

const recentPackages = ref<any[]>([])

const quickActions = [
  { name: t('nav.users'), icon: UserFilled, color: '#409eff', path: '/admin/users' },
  { name: t('nav.tokens'), icon: Key, color: '#67c23a', path: '/admin/tokens' },
  { name: t('nav.backup'), icon: DownloadIcon, color: '#e6a23c', path: '/admin/backup' },
  { name: t('nav.gc'), icon: Delete, color: '#f56c6c', path: '/admin/gc' },
  { name: t('nav.settings'), icon: Setting, color: '#909399', path: '/admin/settings' },
  { name: t('nav.webhooks'), icon: Connection, color: '#8e44ad', path: '/admin/webhooks' },
]

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 MB'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

const loadStats = async () => {
  loading.value = true
  try {
    const [statsRes, usersRes, pkgRes] = await Promise.all([
      adminApi.getStats(),
      adminApi.getUsers(),
      packageApi.getPackages(),
    ])

    stats.totalPackages = statsRes.data.totalPackages || 0
    stats.storageSize = statsRes.data.storageSize || 0
    stats.users = (usersRes.data.users || []).length
    const allPackages = pkgRes.data.objects || []
    recentPackages.value = allPackages.slice(0, 5)
  } catch {
    ElMessage.error(t('home.stats.loadError'))
  } finally {
    loading.value = false
  }
}

const loadSystemInfo = async () => {
  try {
    const res = await adminApi.getSystemInfo()
    Object.assign(sysInfo, res.data)
  } catch {
    // ignore
  }
}

onMounted(() => {
  loadStats()
  loadSystemInfo()
})
</script>

<style scoped>
.dashboard-page {
  padding: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 8px;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
}

.stat-icon.blue {
  background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
  color: #1976d2;
}

.stat-icon.green {
  background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%);
  color: #388e3c;
}

.stat-icon.orange {
  background: linear-gradient(135deg, #fff3e0 0%, #ffe0b2 100%);
  color: #f57c00;
}

.stat-icon.purple {
  background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%);
  color: #7b1fa2;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.quick-actions {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.quick-action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  background: #f5f7fa;
}

.quick-action-item:hover {
  background: #e4e7ed;
  transform: translateY(-2px);
}

.action-name {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}

.recent-activity,
.system-overview {
  margin-top: 20px;
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

:deep(.el-timeline-item__content) {
  padding-left: 8px;
}
</style>
