<template>
  <div class="audit-logs-page">
    <div class="page-header">
      <div>
        <h2>{{ $t('audit.title') }}</h2>
        <p class="description">{{ $t('audit.description') }}</p>
      </div>
    </div>

    <el-card shadow="hover" v-loading="loading">
      <el-table :data="logs" stripe>
        <el-table-column :label="$t('audit.action')" width="150">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)" size="small">
              {{ formatAction(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="username" :label="$t('audit.user')" width="130" />
        <el-table-column prop="ip" :label="$t('audit.ip')" width="140" />
        <el-table-column prop="detail" :label="$t('audit.detail')" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createdAt" :label="$t('common.createdAt')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && logs.length === 0" :description="$t('audit.noLogs')" />

      <div class="pagination" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="limit"
          :total="total"
          layout="prev, pager, next, total"
          @current-change="loadLogs"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)
const logs = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

const formatAction = (action: string): string => {
  const actionMap: Record<string, string> = {
    login: t('audit.actionLogin'),
    logout: t('audit.actionLogout'),
    package_publish: t('audit.actionPackagePublish'),
    package_unpublish: t('audit.actionPackageUnpublish'),
    user_create: t('audit.actionUserCreate'),
    user_update: t('audit.actionUserUpdate'),
    user_delete: t('audit.actionUserDelete'),
    config_update: t('audit.actionConfigUpdate'),
    token_create: t('audit.actionTokenCreate'),
    token_delete: t('audit.actionTokenDelete'),
    webhook_create: t('audit.actionWebhookCreate'),
    webhook_update: t('audit.actionWebhookUpdate'),
    webhook_delete: t('audit.actionWebhookDelete'),
  }
  return actionMap[action] || action
}

const getActionType = (action: string): 'danger' | 'warning' | 'success' | 'primary' | 'info' => {
  if (action === 'login' || action === 'logout') return 'success'
  if (action === 'package_publish') return 'primary'
  if (action === 'package_unpublish' || action === 'user_delete' || action === 'token_delete' || action === 'webhook_delete') return 'danger'
  if (action.includes('create') || action.includes('update')) return 'warning'
  return 'info'
}

const loadLogs = async (currentPage = page.value) => {
  loading.value = true
  try {
    const res = await adminApi.getAuditLogs(currentPage, limit)
    logs.value = res.data.logs || []
    total.value = res.data.total || 0
    page.value = currentPage
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.audit-logs-page {
  padding: 0;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 20px;
}

.page-header .description {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
