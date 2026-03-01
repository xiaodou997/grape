<template>
  <div class="system-info-page">
    <el-row :gutter="20">
      <!-- 基本信息 -->
      <el-col :span="12">
        <el-card shadow="hover" v-loading="loading">
          <template #header>
            <div class="section-header">
              <el-icon><InfoFilled /></el-icon>
              <span>{{ $t('admin.basicInfo') }}</span>
            </div>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="$t('admin.version')">
              <el-tag type="primary">{{ sysInfo.version }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('admin.uptime')">{{ sysInfo.uptime }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.startTime')">{{ formatTime(sysInfo.startTime) }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.host')">{{ sysInfo.host }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.storagePath')">{{ sysInfo.storagePath }}</el-descriptions-item>
            <el-descriptions-item :label="$t('admin.databasePath')">{{ sysInfo.databasePath }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 存储统计 -->
      <el-col :span="12">
        <el-card shadow="hover" v-loading="loading">
          <template #header>
            <div class="section-header">
              <el-icon><PieChart /></el-icon>
              <span>{{ $t('admin.storageStats') }}</span>
            </div>
          </template>
          <div class="storage-stats">
            <div class="stat-item">
              <div class="stat-label">{{ $t('home.stats.localPackages') }}</div>
              <div class="stat-value">{{ stats.totalPackages }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ $t('home.stats.storageSize') }}</div>
              <div class="stat-value">{{ formatSize(stats.storageSize) }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ $t('nav.users') }}</div>
              <div class="stat-value">{{ stats.users }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ $t('nav.webhooks') }}</div>
              <div class="stat-value">{{ stats.webhooks }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 上游源列表 -->
    <el-card shadow="hover" style="margin-top: 20px" v-loading="loading">
      <template #header>
        <div class="section-header">
          <el-icon><Link /></el-icon>
          <span>{{ $t('admin.upstreams') }}</span>
        </div>
      </template>
      <el-table :data="sysInfo.upstreams" stripe>
        <el-table-column prop="name" :label="$t('common.name')" width="120" />
        <el-table-column prop="url" :label="$t('settings.upstreamUrl')" />
        <el-table-column prop="scope" :label="$t('settings.upstreamScope')" width="120">
          <template #default="{ row }">{{ row.scope || $t('common.default') }}</template>
        </el-table-column>
        <el-table-column prop="enabled" :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="sysInfo.upstreams.length === 0" :description="$t('settings.noUpstreams')" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { InfoFilled, PieChart, Link } from '@element-plus/icons-vue'
import { adminApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)

const sysInfo = reactive({
  version: '',
  startTime: '',
  uptime: '',
  storagePath: '',
  databasePath: '',
  host: '',
  upstreams: [] as any[],
})

const stats = reactive({
  totalPackages: 0,
  storageSize: 0,
  users: 0,
  webhooks: 0,
})

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const loadData = async () => {
  loading.value = true
  try {
    const [sysRes, statsRes, usersRes, webhookRes] = await Promise.all([
      adminApi.getSystemInfo(),
      adminApi.getStats(),
      adminApi.getUsers(),
      adminApi.getWebhooks(),
    ])
    Object.assign(sysInfo, sysRes.data)
    stats.totalPackages = statsRes.data.totalPackages || 0
    stats.storageSize = statsRes.data.storageSize || 0
    stats.users = (usersRes.data.users || []).length
    stats.webhooks = (webhookRes.data.webhooks || []).length
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.system-info-page {
  padding: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.storage-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-item .stat-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-item .stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}
</style>
