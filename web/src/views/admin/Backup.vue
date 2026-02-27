<template>
  <div class="backup-page">
    <el-row :gutter="20">
      <!-- 备份信息 -->
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>数据概览</span>
              <el-button type="primary" @click="loadBackupInfo">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <el-row :gutter="20" v-loading="loading">
            <el-col :span="6">
              <el-statistic title="总包数" :value="backupInfo.totalPackages" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="存储占用" :value="formatSize(backupInfo.storageSize)" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="数据库大小" :value="formatSize(backupInfo.databaseSize)" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="数据目录" :value="backupInfo.dataDir || '-'" />
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 创建备份 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span>创建备份</span>
          </template>
          <div class="backup-section">
            <el-icon class="backup-icon" :size="48"><Download /></el-icon>
            <p>将当前数据（数据库 + 包文件）打包下载</p>
            <el-button type="primary" size="large" @click="createBackup" :loading="downloading">
              <el-icon><Download /></el-icon>
              下载备份文件
            </el-button>
          </div>
          <el-alert
            type="info"
            :closable="false"
            style="margin-top: 16px"
          >
            <template #title>
              备份内容包括：SQLite 数据库、所有包元数据和 tarball 文件
            </template>
          </el-alert>
        </el-card>
      </el-col>

      <!-- 恢复备份 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span>恢复备份</span>
          </template>
          <div class="backup-section">
            <el-icon class="backup-icon" :size="48"><Upload /></el-icon>
            <p>从备份文件恢复数据</p>
            <el-upload
              ref="uploadRef"
              :auto-upload="false"
              :show-file-list="true"
              :limit="1"
              accept=".tar.gz,.tgz"
              :on-change="handleFileChange"
              :on-exceed="handleExceed"
            >
              <template #trigger>
                <el-button type="success" size="large">
                  <el-icon><Upload /></el-icon>
                  选择备份文件
                </el-button>
              </template>
            </el-upload>
            <el-button
              type="warning"
              size="large"
              style="margin-top: 12px"
              :disabled="!selectedFile"
              :loading="restoring"
              @click="restoreBackup"
            >
              <el-icon><RefreshRight /></el-icon>
              恢复数据
            </el-button>
          </div>
          <el-alert
            type="warning"
            :closable="false"
            style="margin-top: 16px"
          >
            <template #title>
              ⚠️ 恢复操作会覆盖当前数据，请谨慎操作！恢复后需要重启服务。
            </template>
          </el-alert>
        </el-card>
      </el-col>
    </el-row>

    <!-- 自动备份列表 -->
    <el-card shadow="hover" style="margin-top: 20px" v-if="backups.length > 0">
      <template #header>
        <span>自动备份记录</span>
      </template>
      <el-table :data="backups" style="width: 100%">
        <el-table-column prop="name" label="备份名称" />
        <el-table-column prop="createdAt" label="创建时间" />
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button type="primary" size="small" @click="downloadAutoBackup(scope.row.name)">
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 使用说明 -->
    <el-card shadow="hover" style="margin-top: 20px">
      <template #header>
        <span>使用说明</span>
      </template>
      <el-collapse>
        <el-collapse-item title="定时备份建议" name="1">
          <p>建议在服务器上配置定时任务（cron）自动备份：</p>
          <pre class="code-block"><code v-pre># 每天凌晨 3 点自动备份
0 3 * * * /path/to/grape backup -o /backup/grape-$(date +\%Y\%m\%d).tar.gz</code></pre>
        </el-collapse-item>
        <el-collapse-item title="命令行备份" name="2">
          <p>也可以使用命令行工具进行备份：</p>
          <pre class="code-block"><code v-pre># 创建备份
grape backup -o backup.tar.gz

# 查看备份内容
grape list -i backup.tar.gz

# 恢复备份
grape restore -i backup.tar.gz --force</code></pre>
        </el-collapse-item>
        <el-collapse-item title="恢复后注意事项" name="3">
          <ul>
            <li>恢复完成后需要重启 Grape 服务</li>
            <li>如果配置文件有变化，需要手动更新配置</li>
            <li>恢复前会自动备份当前数据到 <code>.restore-backup-时间戳</code> 目录</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadInstance, UploadFile, UploadFiles } from 'element-plus'
import { Download, Upload, Refresh, RefreshRight } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

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
const backupInfo = ref<BackupInfo>({
  totalPackages: 0,
  storageSize: 0,
  databaseSize: 0,
  dataDir: '',
})
const backups = ref<Backup[]>([])
const selectedFile = ref<File | null>(null)
const uploadRef = ref<UploadInstance>()

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const loadBackupInfo = async () => {
  loading.value = true
  try {
    const res = await adminApi.getBackupInfo()
    backupInfo.value = res.data
  } catch {
    ElMessage.error('加载备份信息失败')
  } finally {
    loading.value = false
  }
}

const loadBackups = async () => {
  try {
    const res = await adminApi.listBackups()
    backups.value = res.data.backups || []
  } catch {
    // 忽略错误
  }
}

const createBackup = async () => {
  try {
    downloading.value = true
    const response = await adminApi.downloadBackup()
    
    // 创建下载链接
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `grape-backup-${new Date().toISOString().slice(0, 19).replace(/[:-]/g, '')}.tar.gz`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('备份下载成功')
  } catch {
    ElMessage.error('备份失败')
  } finally {
    downloading.value = false
  }
}

const handleFileChange = (file: UploadFile, _uploadFiles: UploadFiles) => {
  selectedFile.value = file.raw || null
}

const handleExceed = () => {
  ElMessage.warning('只能选择一个文件')
}

const restoreBackup = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择备份文件')
    return
  }

  try {
    await ElMessageBox.confirm(
      '恢复操作会覆盖当前数据，确定要继续吗？',
      '警告',
      {
        confirmButtonText: '确定恢复',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    restoring.value = true
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const res = await adminApi.restoreBackup(formData)
    
    ElMessage.success(res.data.message || '恢复成功')
    
    // 清除文件选择
    selectedFile.value = null
    uploadRef.value?.clearFiles()
    
    // 提示重启
    if (res.data.restart) {
      ElMessageBox.alert(
        '数据已恢复，请重启 Grape 服务以应用更改。',
        '需要重启',
        {
          confirmButtonText: '知道了',
          type: 'info',
        }
      )
    }
  } catch (error: unknown) {
    if (error !== 'cancel') {
      const err = error as { response?: { data?: { error?: string } } }
      ElMessage.error(err.response?.data?.error || '恢复失败')
    }
  } finally {
    restoring.value = false
  }
}

const downloadAutoBackup = async (name: string) => {
  ElMessage.info('此功能暂未实现')
}

onMounted(() => {
  loadBackupInfo()
  loadBackups()
})
</script>

<style scoped>
.backup-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.backup-section {
  text-align: center;
  padding: 20px 0;
}

.backup-icon {
  color: var(--el-color-primary);
  margin-bottom: 12px;
}

.code-block {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

.code-block code {
  font-family: monospace;
  font-size: 13px;
}

:deep(.el-upload) {
  display: inline-block;
}
</style>
