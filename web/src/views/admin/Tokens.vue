<template>
  <div class="tokens-page fade-in">
    <div class="page-header-modern">
      <div class="header-info">
        <h3>{{ t('tokens.title') }}</h3>
        <p class="subtitle">{{ t('webhooks.description') }}</p>
      </div>
      <el-button type="primary" @click="showCreateDialog" class="btn-with-shadow">
        <el-icon><Plus /></el-icon>
        {{ t('tokens.createToken') }}
      </el-button>
    </div>

    <div class="table-container">
      <el-table :data="tokens" v-loading="loading" class="modern-table">
        <el-table-column prop="name" :label="t('tokens.tokenName')" min-width="150">
          <template #default="{ row }">
            <span class="token-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.type')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.readonly ? 'info' : 'success'" size="small" effect="light" round>
              {{ row.readonly ? t('tokens.readonly') : t('nav.packages') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('tokens.expiresAt')" width="180">
          <template #default="{ row }">
            <span v-if="row.expiresAt" class="time-cell">{{ formatDate(row.expiresAt) }}</span>
            <span v-else class="status-success">{{ t('tokens.neverExpires') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('tokens.lastUsed')" width="180">
          <template #default="{ row }">
            <span v-if="row.lastUsed" class="time-cell">{{ formatDate(row.lastUsed) }}</span>
            <span v-else class="status-muted">{{ t('tokens.never') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" text size="small" @click="confirmDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Create Token Dialog -->
    <el-dialog v-model="createDialogVisible" :title="t('tokens.createToken')" width="480px" class="modern-dialog">
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="t('tokens.tokenName')" required>
          <el-input v-model="createForm.name" placeholder="e.g. production-ci" />
        </el-form-item>
        <el-form-item :label="t('common.type')">
          <el-radio-group v-model="createForm.readonly">
            <el-radio-button :value="false">{{ t('nav.packages') }}</el-radio-button>
            <el-radio-button :value="true">{{ t('tokens.readonly') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('tokens.expiresAt')">
          <el-select v-model="createForm.days" style="width: 100%">
            <el-option :label="t('tokens.neverExpires')" :value="0" />
            <el-option :label="`7 ${t('tokens.days')}`" :value="7" />
            <el-option :label="`30 ${t('tokens.days')}`" :value="30" />
            <el-option :label="`90 ${t('tokens.days')}`" :value="90" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="createToken" :loading="creating">{{ t('common.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- New Token Reveal -->
    <el-dialog v-model="showTokenDialog" :title="t('tokens.tokenCreated')" width="500px" :close-on-click-modal="false" class="modern-dialog">
      <el-alert :title="t('tokens.tokenWarning')" type="warning" :closable="false" show-icon />
      
      <div class="token-reveal-card">
        <div class="token-value-box">
          <code>{{ newToken }}</code>
        </div>
        <el-button type="primary" @click="copyToken" class="copy-btn">
          <el-icon><DocumentCopy /></el-icon>
          {{ t('tokens.copyToken') }}
        </el-button>
      </div>

      <template #footer>
        <el-button type="primary" @click="showTokenDialog = false" style="width: 100%">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, DocumentCopy } from '@element-plus/icons-vue'
import { tokenApi } from '@/api'

const { t } = useI18n()

interface Token {
  id: number
  name: string
  readonly: boolean
  expiresAt: string | null
  lastUsed: string | null
  createdAt: string
}

const tokens = ref<Token[]>([])
const loading = ref(false)
const createDialogVisible = ref(false)
const showTokenDialog = ref(false)
const creating = ref(false)
const newToken = ref('')

const createForm = ref({
  name: '',
  readonly: false,
  days: 0
})

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

const loadTokens = async () => {
  loading.value = true
  try {
    const res = await tokenApi.list()
    tokens.value = res.data.objects || []
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  createForm.value = { name: '', readonly: false, days: 0 }
  createDialogVisible.value = true
}

const createToken = async () => {
  if (!createForm.value.name.trim()) return
  creating.value = true
  try {
    const res = await tokenApi.create({
      name: createForm.value.name,
      readonly: createForm.value.readonly,
      days: createForm.value.days || undefined
    })
    newToken.value = res.data.token
    createDialogVisible.value = false
    showTokenDialog.value = true
    loadTokens()
  } catch {
    ElMessage.error(t('errors.saveFailed'))
  } finally {
    creating.value = false
  }
}

const copyToken = async () => {
  const text = newToken.value
  try {
    // 优先使用新 API
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else {
      // 回退到传统方案
      const textArea = document.createElement("textarea")
      textArea.value = text
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    }
    ElMessage.success(t('tokens.tokenCopied'))
  } catch {
    ElMessage.error('Copy failed')
  }
}

const confirmDelete = (token: Token) => {
  ElMessageBox.confirm(`${t('tokens.deleteTokenConfirm')} "${token.name}"?`, t('common.warning'), {
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel'),
    type: 'warning'
  }).then(async () => {
    await tokenApi.delete(token.id)
    ElMessage.success(t('tokens.tokenDeleted'))
    loadTokens()
  })
}

onMounted(loadTokens)
</script>

<style scoped>
.tokens-page { padding: 0; }
.page-header-modern { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.subtitle { font-size: 14px; color: var(--g-text-secondary); margin-top: 4px; }
.token-name { font-weight: 600; color: var(--g-text-primary); }
.time-cell { font-size: 13px; color: var(--g-text-secondary); }
.status-success { color: var(--g-success); font-weight: 500; font-size: 13px; }
.status-muted { color: var(--g-text-muted); font-size: 13px; }

.token-reveal-card {
  margin-top: 24px;
  background: var(--g-bg);
  padding: 20px;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.token-value-box {
  background: #0f172a;
  padding: 16px;
  border-radius: 12px;
  overflow-x: auto;
}

.token-value-box code {
  color: #38bdf8;
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
}

.copy-btn { width: 100%; height: 44px !important; }
.btn-with-shadow { box-shadow: 0 4px 6px -1px rgba(124, 58, 237, 0.2); }
</style>
