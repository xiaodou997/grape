<template>
  <div class="webhooks-page fade-in">
    <div class="page-header-modern">
      <div class="header-info">
        <h3>{{ t('webhooks.title') }}</h3>
        <p class="subtitle">{{ t('webhooks.description') }}</p>
      </div>
      <el-button type="primary" @click="openCreateDialog" class="btn-with-shadow">
        <el-icon><Plus /></el-icon>
        {{ t('webhooks.createWebhook') }}
      </el-button>
    </div>

    <div class="table-container">
      <el-table :data="webhooks" v-loading="loading" class="modern-table">
        <el-table-column prop="name" :label="t('common.name')" width="150">
          <template #default="{ row }">
            <span class="webhook-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="url" :label="t('webhooks.webhookUrl')" min-width="250">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" :href="row.url" target="_blank" class="url-link">
              {{ row.url }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column :label="t('webhooks.events')" width="200">
          <template #default="{ row }">
            <div class="event-tags">
              <el-tag v-for="event in formatEvents(row.events)" :key="event" size="small" effect="light" round>
                {{ event }}
              </el-tag>
              <span v-if="!row.events" class="status-muted">{{ t('webhooks.allEvents') }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" size="small" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="lastDeliveryAt" :label="t('webhooks.lastDelivery')" width="160">
          <template #default="{ row }">
            <span class="time-cell">{{ row.lastDeliveryAt ? formatTime(row.lastDeliveryAt) : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button text type="primary" size="small" @click="handleEdit(row)">
                {{ t('common.edit') }}
              </el-button>
              <el-button text type="success" size="small" @click="handleTest(row)" :loading="row.testing">
                {{ t('webhooks.test') }}
              </el-button>
              <el-button text type="danger" size="small" @click="handleDelete(row)">
                {{ t('common.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && webhooks.length === 0" :description="t('webhooks.noWebhooks')" />
    </div>

    <!-- Webhook Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingWebhook ? t('webhooks.editWebhook') : t('webhooks.createWebhook')"
      width="520px"
      class="modern-dialog"
    >
      <el-form :model="form" label-position="top" :rules="rules" ref="formRef">
        <el-form-item :label="t('common.name')" prop="name" required>
          <el-input v-model="form.name" :placeholder="t('webhooks.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('webhooks.webhookUrl')" prop="url" required>
          <el-input v-model="form.url" placeholder="https://example.com/webhook" />
        </el-form-item>
        <el-form-item :label="t('webhooks.secret')">
          <el-input
            v-model="form.secret"
            type="password"
            show-password
            :placeholder="t('webhooks.secretPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('webhooks.events')">
          <el-select
            v-model="form.eventsArr"
            multiple
            collapse-tags
            collapse-tags-indicator
            :placeholder="t('webhooks.eventsPlaceholder')"
            style="width: 100%"
          >
            <el-option label="package:published" value="package:published" />
            <el-option label="package:unpublished" value="package:unpublished" />
            <el-option label="user:created" value="user:created" />
            <el-option label="user:deleted" value="user:deleted" />
          </el-select>
          <div class="form-hint">{{ t('webhooks.eventsHint') }}</div>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <div class="left-actions">
            <el-button 
              v-if="editingWebhook" 
              type="success" 
              plain 
              @click="handleTest(editingWebhook)" 
              :loading="editingWebhook.testing"
            >
              {{ t('webhooks.testWebhook') }}
            </el-button>
          </div>
          <div class="right-actions">
            <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
            <el-button type="primary" @click="saveWebhook" :loading="saving">
              {{ t('common.save') }}
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingWebhook = ref<any>(null)
const formRef = ref<any>()

const webhooks = ref<any[]>([])

const form = reactive({
  name: '',
  url: '',
  secret: '',
  eventsArr: [] as string[],
  enabled: true,
})

const rules = {
  name: [{ required: true, message: t('webhooks.nameRequired'), trigger: 'blur' }],
  url: [
    { required: true, message: t('webhooks.urlRequired'), trigger: 'blur' },
    { type: 'url', message: t('webhooks.urlInvalid'), trigger: 'blur' },
  ],
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  return new Date(time).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

const formatEvents = (events?: string): string[] => {
  if (!events) return []
  return events.split(',').filter(Boolean)
}

const loadWebhooks = async () => {
  loading.value = true
  try {
    const res = await adminApi.getWebhooks()
    webhooks.value = (res.data.webhooks || []).map((w: any) => ({ ...w, testing: false }))
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

const toggleEnabled = async (row: any) => {
  try {
    await adminApi.updateWebhook(row.id, { enabled: row.enabled })
    ElMessage.success(t('common.saveSuccess'))
  } catch {
    row.enabled = !row.enabled
    ElMessage.error(t('errors.saveFailed'))
  }
}

const openCreateDialog = () => {
  editingWebhook.value = null
  Object.assign(form, { name: '', url: '', secret: '', eventsArr: [], enabled: true })
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  editingWebhook.value = row
  Object.assign(form, {
    name: row.name,
    url: row.url,
    secret: '',
    eventsArr: row.events ? row.events.split(',').filter(Boolean) : [],
    enabled: row.enabled,
  })
  dialogVisible.value = true
}

const saveWebhook = async () => {
  const valid = await formRef.value?.validate()
  if (!valid) return

  saving.value = true
  try {
    const payload = {
      ...form,
      events: form.eventsArr.join(','),
    }
    if (editingWebhook.value) {
      await adminApi.updateWebhook(editingWebhook.value.id, payload)
      ElMessage.success(t('webhooks.webhookUpdated'))
    } else {
      await adminApi.createWebhook(payload)
      ElMessage.success(t('webhooks.webhookCreated'))
    }
    dialogVisible.value = false
    loadWebhooks()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || t('errors.saveFailed'))
  } finally {
    saving.value = false
  }
}

const handleTest = async (row: any) => {
  row.testing = true
  try {
    await adminApi.testWebhook(row.id)
    ElMessage.success(t('webhooks.webhookTested'))
  } catch {
    ElMessage.error(t('errors.testFailed'))
  } finally {
    row.testing = false
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(t('webhooks.deleteConfirm', { name: row.name }), t('common.warning'), {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
    await adminApi.deleteWebhook(row.id)
    ElMessage.success(t('webhooks.webhookDeleted'))
    loadWebhooks()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.response?.data?.error || t('errors.deleteFailed'))
  }
}

onMounted(loadWebhooks)
</script>

<style scoped>
.webhooks-page { padding: 0; }
.page-header-modern { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.subtitle { font-size: 14px; color: var(--g-text-secondary); margin-top: 4px; }

.webhook-name { font-weight: 600; color: var(--g-text-primary); }
.url-link { font-size: 13px; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.status-muted { color: var(--g-text-muted); font-size: 12px; }
.time-cell { font-size: 13px; color: var(--g-text-secondary); }

.action-buttons { display: flex; gap: 8px; align-items: center; }
.btn-with-shadow { box-shadow: 0 4px 6px -1px rgba(124, 58, 237, 0.2); }

.form-hint { font-size: 12px; color: var(--g-text-muted); margin-top: 4px; }

.dialog-footer { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.right-actions { display: flex; gap: 12px; }
</style>
