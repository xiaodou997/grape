<template>
  <div class="webhooks-page">
    <div class="page-header">
      <div>
        <h2>{{ $t('webhooks.title') }}</h2>
        <p class="description">{{ $t('webhooks.description') }}</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        {{ $t('webhooks.createWebhook') }}
      </el-button>
    </div>

    <el-table :data="webhooks" stripe v-loading="loading">
      <el-table-column prop="name" :label="$t('common.name')" width="150" />
      <el-table-column prop="url" :label="$t('webhooks.webhookUrl')" min-width="250">
        <template #default="{ row }">
          <el-link type="primary" :href="row.url" target="_blank">{{ row.url }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="events" :label="$t('webhooks.events')" width="180">
        <template #default="{ row }">
          <el-tag v-for="event in formatEvents(row.events)" :key="event" size="small" class="event-tag">
            {{ event }}
          </el-tag>
          <span v-if="!row.events">{{ $t('webhooks.allEvents') }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="enabled" :label="$t('common.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">
            {{ row.enabled ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastDeliveryAt" :label="$t('webhooks.lastDelivery')" width="160">
        <template #default="{ row }">
          {{ row.lastDeliveryAt ? formatTime(row.lastDeliveryAt) : '-' }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.actions')" width="200" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="handleEdit(row)">
            {{ $t('common.edit') }}
          </el-button>
          <el-button text type="success" size="small" @click="handleTest(row)" :loading="row.testing">
            {{ $t('webhooks.test') }}
          </el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && webhooks.length === 0" :description="$t('webhooks.noWebhooks')" />

    <!-- Webhook 对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingWebhook ? $t('webhooks.editWebhook') : $t('webhooks.createWebhook')"
      width="500px"
    >
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item :label="$t('common.name')" prop="name" required>
          <el-input v-model="form.name" :placeholder="$t('webhooks.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('webhooks.webhookUrl')" prop="url" required>
          <el-input v-model="form.url" placeholder="https://example.com/webhook" />
        </el-form-item>
        <el-form-item :label="$t('webhooks.secret')">
          <el-input
            v-model="form.secret"
            type="password"
            show-password
            :placeholder="$t('webhooks.secretPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="$t('webhooks.events')">
          <el-select
            v-model="form.eventsArr"
            multiple
            :placeholder="$t('webhooks.eventsPlaceholder')"
            style="width: 100%"
          >
            <el-option label="package:published" value="package:published" />
            <el-option label="package:unpublished" value="package:unpublished" />
            <el-option label="user:created" value="user:created" />
            <el-option label="user:deleted" value="user:deleted" />
          </el-select>
          <div class="form-hint">{{ $t('webhooks.eventsHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('common.status')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveWebhook" :loading="saving">
          {{ $t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '@/api'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingWebhook = ref<any>(null)
const formRef = ref<FormInstance>()

const webhooks = ref<any[]>([])

const form = reactive({
  name: '',
  url: '',
  secret: '',
  eventsArr: [] as string[],
  enabled: true,
})

const rules: FormRules = {
  name: [{ required: true, message: t('webhooks.nameRequired'), trigger: 'blur' }],
  url: [
    { required: true, message: t('webhooks.urlRequired'), trigger: 'blur' },
    { type: 'url', message: t('webhooks.urlInvalid'), trigger: 'blur' },
  ],
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
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

const resetForm = () => {
  form.name = ''
  form.url = ''
  form.secret = ''
  form.eventsArr = []
  form.enabled = true
  editingWebhook.value = null
}

const openCreateDialog = () => {
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  editingWebhook.value = row
  form.name = row.name
  form.url = row.url
  form.secret = ''
  form.eventsArr = row.events ? row.events.split(',').filter(Boolean) : []
  form.enabled = row.enabled
  dialogVisible.value = true
}

const saveWebhook = async () => {
  const valid = await formRef.value?.validate()
  if (!valid) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      url: form.url,
      secret: form.secret,
      events: form.eventsArr.join(','),
      enabled: form.enabled,
    }
    if (editingWebhook.value) {
      await adminApi.updateWebhook(editingWebhook.value.id, payload)
      ElMessage.success(t('webhooks.webhookUpdated'))
    } else {
      await adminApi.createWebhook(payload)
      ElMessage.success(t('webhooks.webhookCreated'))
    }
    dialogVisible.value = false
    resetForm()
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
    await ElMessageBox.confirm(
      t('webhooks.deleteConfirm', { name: row.name }),
      t('common.warning'),
      {
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )
    await adminApi.deleteWebhook(row.id)
    ElMessage.success(t('webhooks.webhookDeleted'))
    loadWebhooks()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || t('errors.deleteFailed'))
    }
  }
}

onMounted(() => {
  loadWebhooks()
})
</script>

<style scoped>
.webhooks-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
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

.event-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
