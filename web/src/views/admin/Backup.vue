<template>
  <div class="backup-page fade-in">
    <!-- Top Stats -->
    <div class="backup-stats-row">
      <div class="mini-stat-card">
        <span class="m-label">{{ t('gc.totalPackages') }}</span>
        <span class="m-value">{{ backupInfo.totalPackages }}</span>
      </div>
      <div class="mini-stat-card">
        <span class="m-label">{{ t('home.stats.storageSize') }}</span>
        <span class="m-value">{{ formatSize(backupInfo.storageSize) }}</span>
      </div>
      <div class="mini-stat-card">
        <span class="m-label">DB Size</span>
        <span class="m-value">{{ formatSize(backupInfo.databaseSize) }}</span>
      </div>
    </div>

    <!-- Core Actions: Create & Restore -->
    <div class="action-grid-v2">
      <div class="action-card-v2">
        <div class="card-v2-header">
          <div class="icon-circle p-blue"><el-icon><Download /></el-icon></div>
          <h3>{{ t('backup.createBackup') }}</h3>
        </div>
        <p class="card-v2-desc">{{ t('backup.backupInfo') }}</p>
        <div class="card-v2-body">
          <el-button type="primary" size="large" @click="createBackup" :loading="downloading" class="w-full">
            {{ t('backup.downloadBackup') }}
          </el-button>
        </div>
        <el-alert :title="t('backup.backupInfo')" type="info" :closable="false" show-icon />
      </div>

      <div class="action-card-v2">
        <div class="card-v2-header">
          <div class="icon-circle p-orange"><el-icon><Upload /></el-icon></div>
          <h3>{{ t('backup.restoreBackup') }}</h3>
        </div>
        <p class="card-v2-desc">{{ t('backup.restoreConfirm') }}</p>
        <div class="card-v2-body">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            accept=".tar.gz,.tgz"
            :on-change="handleFileChange"
            class="upload-full-width"
          >
            <template #trigger>
              <el-button type="default" size="large" class="w-full">{{ t('backup.selectFile') }}</el-button>
            </template>
          </el-upload>
          <el-button
            type="warning"
            size="large"
            class="w-full mt-12"
            :disabled="!selectedFile"
            :loading="restoring"
            @click="restoreBackup"
          >
            {{ t('backup.restoreBackup') }}
          </el-button>
        </div>
        <el-alert title="⚠️ Warning: This will overwrite data" type="warning" :closable="false" show-icon />
      </div>
    </div>

    <!-- History -->
    <section class="backup-history-section" v-if="backups.length > 0">
      <h3 class="section-title-v2">{{ t('backup.backupList') }}</h3>
      <el-table :data="backups" class="modern-table">
        <el-table-column prop="name" :label="t('common.name')" min-width="200" />
        <el-table-column prop="createdAt" :label="t('common.createdAt')" width="180" />
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
          <template #default="scope">
            <el-button text type="primary" size="small" @click="downloadAutoBackup(scope.row.name)">
              {{ t('common.download') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadInstance, UploadFile, UploadFiles } from 'element-plus'
import { Download, Upload } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

const { t } = useI18n()

interface BackupInfo {
  totalPackages: number
  storageSize: number
  databaseSize: number
  dataDir: string
}

interface Backup {
  name: string
  createdAt: string
  auto: boolean
}

const loading = ref(false)
const downloading = ref(false)
const restoring = ref(false)
const backupInfo = ref<BackupInfo>({ totalPackages: 0, storageSize: 0, databaseSize: 0, dataDir: '' })
const backups = ref<Backup[]>([])
const selectedFile = ref<File | null>(null)
const uploadRef = ref<UploadInstance>()

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const loadBackupInfo = async () => {
  loading.value = true
  try {
    const res = await adminApi.getBackupInfo()
    backupInfo.value = res.data
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

const createBackup = async () => {
  try {
    downloading.value = true
    const response = await adminApi.downloadBackup()
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `grape-backup-${new Date().getTime()}.tar.gz`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    ElMessage.success(t('common.success'))
  } catch {
    ElMessage.error(t('errors.saveFailed'))
  } finally {
    downloading.value = false
  }
}

const handleFileChange = (file: UploadFile) => { selectedFile.value = file.raw || null }

const restoreBackup = async () => {
  if (!selectedFile.value) return
  try {
    await ElMessageBox.confirm(t('backup.restoreConfirm'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
    restoring.value = true
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    await adminApi.restoreBackup(formData)
    ElMessage.success(t('common.success'))
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.response?.data?.error || 'Restore failed')
  } finally {
    restoring.value = false
  }
}

const downloadAutoBackup = (name: string) => { ElMessage.info('Developing...') }

onMounted(() => {
  loadBackupInfo()
})
</script>

<style scoped>
.backup-page { display: flex; flex-direction: column; gap: 32px; }
.backup-stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
.mini-stat-card { background: white; padding: 20px; border-radius: 16px; border: 1px solid var(--g-border); display: flex; flex-direction: column; gap: 4px; }
.m-label { font-size: 12px; color: var(--g-text-muted); font-weight: 600; text-transform: uppercase; }
.m-value { font-size: 20px; font-weight: 700; color: var(--g-text-primary); }

.action-grid-v2 { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; align-items: stretch; }
.action-card-v2 { background: white; border-radius: 20px; border: 1px solid var(--g-border); padding: 32px; display: flex; flex-direction: column; height: 100%; }
.card-v2-header { display: flex; align-items: center; gap: 16px; margin-bottom: 16px; }
.card-v2-header h3 { font-size: 18px; font-weight: 700; margin: 0; }
.icon-circle { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 20px; }
.p-blue { background: #eff6ff; color: #3b82f6; }
.p-orange { background: #fff7ed; color: #f97316; }
.card-v2-desc { font-size: 14px; color: var(--g-text-secondary); margin-bottom: 24px; flex: 1; }
.card-v2-body { margin-bottom: 24px; }

.w-full { width: 100%; }
.mt-12 { margin-top: 12px; }
.upload-full-width :deep(.el-upload) { width: 100%; }

.section-title-v2 { font-size: 18px; font-weight: 700; margin-bottom: 16px; }

@media (max-width: 768px) {
  .action-grid-v2 { grid-template-columns: 1fr; }
}
</style>
