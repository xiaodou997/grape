<template>
  <div class="settings-page">
    <el-tabs v-model="activeTab" type="card">
      <!-- Tab 1: 系统信息 -->
      <el-tab-pane label="系统信息" name="system">
        <div v-loading="sysLoading">
          <el-descriptions :column="2" border class="sys-desc">
            <el-descriptions-item label="版本号">{{ sysInfo.version }}</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ sysInfo.uptime }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">{{ formatTime(sysInfo.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="监听地址">{{ sysInfo.host }}</el-descriptions-item>
            <el-descriptions-item label="存储路径">{{ sysInfo.storagePath }}</el-descriptions-item>
            <el-descriptions-item label="数据库路径">{{ sysInfo.databasePath }}</el-descriptions-item>
          </el-descriptions>

          <el-divider>上游源列表</el-divider>
          <el-table :data="sysInfo.upstreams" stripe>
            <el-table-column prop="name" label="名称" width="120" />
            <el-table-column prop="url" label="URL" />
            <el-table-column prop="scope" label="Scope" width="120">
              <template #default="{ row }">{{ row.scope || '默认' }}</template>
            </el-table-column>
            <el-table-column prop="enabled" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- Tab 2: 配置管理 -->
      <el-tab-pane label="配置管理" name="config">
        <div v-loading="cfgLoading">
          <el-form label-width="120px">
            <!-- 上游源管理 -->
            <el-divider content-position="left">上游源管理</el-divider>
            <el-table :data="cfgForm.upstreams" border>
              <el-table-column label="名称" width="120">
                <template #default="{ row }">
                  <el-input v-model="row.name" size="small" />
                </template>
              </el-table-column>
              <el-table-column label="URL">
                <template #default="{ row }">
                  <el-input v-model="row.url" size="small" />
                </template>
              </el-table-column>
              <el-table-column label="Scope" width="120">
                <template #default="{ row }">
                  <el-input v-model="row.scope" size="small" placeholder="留空=默认" />
                </template>
              </el-table-column>
              <el-table-column label="超时(秒)" width="100">
                <template #default="{ row }">
                  <el-input-number v-model="row.timeout" :min="1" :max="300" size="small" style="width: 80px" />
                </template>
              </el-table-column>
              <el-table-column label="启用" width="70">
                <template #default="{ row }">
                  <el-switch v-model="row.enabled" size="small" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="70">
                <template #default="{ $index }">
                  <el-button text type="danger" size="small" @click="removeUpstream($index)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-button style="margin-top: 8px" @click="addUpstream">
              + 添加上游源
            </el-button>

            <!-- 认证配置 -->
            <el-divider content-position="left">认证配置</el-divider>
            <el-form-item label="JWT 密钥">
              <el-input v-model="cfgForm.auth.jwtSecret" type="password" show-password style="max-width: 400px" placeholder="留空则不修改" />
            </el-form-item>
            <el-form-item label="JWT 过期时间">
              <el-input-number v-model="cfgForm.auth.jwtExpiry" :min="1" :max="8760" />
              <span style="margin-left: 8px; color: #909399;">小时</span>
            </el-form-item>
            <el-form-item label="允许自助注册">
              <el-switch v-model="cfgForm.auth.allowRegistration" />
              <span style="margin-left: 8px; color: #909399; font-size: 12px;">关闭后用户只能由管理员创建</span>
            </el-form-item>

            <!-- 日志配置 -->
            <el-divider content-position="left">日志配置</el-divider>
            <el-form-item label="日志级别">
              <el-select v-model="cfgForm.log.level" style="width: 160px">
                <el-option label="debug" value="debug" />
                <el-option label="info" value="info" />
                <el-option label="warn" value="warn" />
                <el-option label="error" value="error" />
              </el-select>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="saveConfig" :loading="cfgSaving">
                保存并热加载
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <!-- Tab 3: Webhook 管理 -->
      <el-tab-pane label="Webhook 管理" name="webhooks">
        <div class="webhook-header">
          <el-button type="primary" @click="openCreateWebhook">
            <el-icon><Plus /></el-icon>
            添加 Webhook
          </el-button>
        </div>

        <el-table :data="webhooks" stripe v-loading="webhookLoading">
          <el-table-column prop="name" label="名称" width="140" />
          <el-table-column prop="url" label="URL" />
          <el-table-column prop="events" label="事件" width="200">
            <template #default="{ row }">
              {{ row.events || '全部事件' }}
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastDeliveryAt" label="最后投递" width="160">
            <template #default="{ row }">
              {{ row.lastDeliveryAt ? formatTime(row.lastDeliveryAt) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button text type="primary" size="small" @click="handleEditWebhook(row)">编辑</el-button>
              <el-button text type="success" size="small" @click="handleTestWebhook(row)">测试</el-button>
              <el-button text type="danger" size="small" @click="handleDeleteWebhook(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!webhookLoading && webhooks.length === 0" description="暂无 Webhook" />

        <!-- Webhook 对话框 -->
        <el-dialog v-model="showWebhookDialog" :title="editingWebhook ? '编辑 Webhook' : '添加 Webhook'" width="480px">
          <el-form :model="webhookForm" label-width="100px">
            <el-form-item label="名称" required>
              <el-input v-model="webhookForm.name" />
            </el-form-item>
            <el-form-item label="URL" required>
              <el-input v-model="webhookForm.url" placeholder="https://example.com/webhook" />
            </el-form-item>
            <el-form-item label="Secret">
              <el-input v-model="webhookForm.secret" type="password" show-password placeholder="可选，用于签名验证" />
            </el-form-item>
            <el-form-item label="事件">
              <el-select v-model="webhookForm.eventsArr" multiple placeholder="留空=订阅全部" style="width: 100%">
                <el-option label="package:published" value="package:published" />
                <el-option label="package:unpublished" value="package:unpublished" />
                <el-option label="user:created" value="user:created" />
                <el-option label="user:deleted" value="user:deleted" />
              </el-select>
            </el-form-item>
            <el-form-item label="启用">
              <el-switch v-model="webhookForm.enabled" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showWebhookDialog = false">取消</el-button>
            <el-button type="primary" @click="saveWebhook" :loading="webhookSaving">保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <!-- Tab 4: 审计日志 -->
      <el-tab-pane label="审计日志" name="audit">
        <el-table :data="auditLogs" stripe v-loading="auditLoading">
          <el-table-column prop="action" label="操作" width="180">
            <template #default="{ row }">
              <el-tag :type="auditActionType(row.action)" size="small">{{ row.action }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="username" label="操作用户" width="130" />
          <el-table-column prop="ip" label="IP 地址" width="140" />
          <el-table-column prop="detail" label="详情" />
          <el-table-column prop="createdAt" label="时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.createdAt) }}
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!auditLoading && auditLogs.length === 0" description="暂无审计日志" />

        <div class="pagination" v-if="auditTotal > 0">
          <el-pagination
            v-model:current-page="auditPage"
            :page-size="auditLimit"
            :total="auditTotal"
            layout="prev, pager, next, total"
            @current-change="loadAuditLogs"
          />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

const activeTab = ref('system')

// ---- 系统信息 ----
const sysLoading = ref(false)
const sysInfo = reactive({
  version: '',
  startTime: '',
  uptime: '',
  storagePath: '',
  databasePath: '',
  host: '',
  upstreams: [] as any[],
})

const loadSystemInfo = async () => {
  sysLoading.value = true
  try {
    const res = await adminApi.getSystemInfo()
    Object.assign(sysInfo, res.data)
  } catch {
    ElMessage.error('加载系统信息失败')
  } finally {
    sysLoading.value = false
  }
}

// ---- 配置管理 ----
const cfgLoading = ref(false)
const cfgSaving = ref(false)
const cfgForm = reactive({
  upstreams: [] as Array<{ name: string; url: string; scope: string; timeout: number; enabled: boolean }>,
  auth: {
    jwtSecret: '',
    jwtExpiry: 24,
    allowRegistration: false,
  },
  log: {
    level: 'info',
  },
})

const loadConfig = async () => {
  cfgLoading.value = true
  try {
    const res = await adminApi.getConfig()
    const data = res.data
    cfgForm.upstreams = (data.registry?.upstreams || []).map((u: any) => ({
      name: u.name || '',
      url: u.url || '',
      scope: u.scope || '',
      timeout: u.timeout || 30,
      enabled: u.enabled !== false,
    }))
    if (data.auth) {
      cfgForm.auth.jwtSecret = ''
      cfgForm.auth.jwtExpiry = data.auth.jwtExpiry || 24
      cfgForm.auth.allowRegistration = !!data.auth.allowRegistration
    }
    if (data.log) {
      cfgForm.log.level = data.log.level || 'info'
    }
  } catch {
    ElMessage.error('加载配置失败')
  } finally {
    cfgLoading.value = false
  }
}

const addUpstream = () => {
  cfgForm.upstreams.push({ name: '', url: '', scope: '', timeout: 30, enabled: true })
}

const removeUpstream = (index: number) => {
  cfgForm.upstreams.splice(index, 1)
}

const saveConfig = async () => {
  cfgSaving.value = true
  try {
    const payload: any = {
      registry: {
        upstreams: cfgForm.upstreams,
      },
      auth: {
        jwtExpiry: cfgForm.auth.jwtExpiry,
        allowRegistration: cfgForm.auth.allowRegistration,
      },
      log: {
        level: cfgForm.log.level,
      },
    }
    if (cfgForm.auth.jwtSecret) {
      payload.auth.jwtSecret = cfgForm.auth.jwtSecret
    }
    await adminApi.updateConfig(payload)
    ElMessage.success('配置已保存，即时生效')
    cfgForm.auth.jwtSecret = ''
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    cfgSaving.value = false
  }
}

// ---- Webhook 管理 ----
const webhookLoading = ref(false)
const webhookSaving = ref(false)
const showWebhookDialog = ref(false)
const editingWebhook = ref<any>(null)
const webhooks = ref<any[]>([])

const webhookForm = reactive({
  name: '',
  url: '',
  secret: '',
  eventsArr: [] as string[],
  enabled: true,
})

const loadWebhooks = async () => {
  webhookLoading.value = true
  try {
    const res = await adminApi.getWebhooks()
    webhooks.value = res.data.webhooks || []
  } catch {
    ElMessage.error('加载 Webhook 列表失败')
  } finally {
    webhookLoading.value = false
  }
}

const openCreateWebhook = () => {
  editingWebhook.value = null
  resetWebhookForm()
  showWebhookDialog.value = true
}

const handleEditWebhook = (row: any) => {
  editingWebhook.value = row
  webhookForm.name = row.name
  webhookForm.url = row.url
  webhookForm.secret = ''
  webhookForm.eventsArr = row.events ? row.events.split(',').filter(Boolean) : []
  webhookForm.enabled = row.enabled
  showWebhookDialog.value = true
}

const saveWebhook = async () => {
  if (!webhookForm.name || !webhookForm.url) {
    ElMessage.warning('名称和 URL 不能为空')
    return
  }
  webhookSaving.value = true
  try {
    const payload = {
      name: webhookForm.name,
      url: webhookForm.url,
      secret: webhookForm.secret,
      events: webhookForm.eventsArr.join(','),
      enabled: webhookForm.enabled,
    }
    if (editingWebhook.value) {
      await adminApi.updateWebhook(editingWebhook.value.id, payload)
      ElMessage.success('更新成功')
    } else {
      await adminApi.createWebhook(payload)
      ElMessage.success('创建成功')
    }
    showWebhookDialog.value = false
    resetWebhookForm()
    loadWebhooks()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    webhookSaving.value = false
  }
}

const handleTestWebhook = async (row: any) => {
  try {
    await adminApi.testWebhook(row.id)
    ElMessage.success('测试事件已发送')
  } catch {
    ElMessage.error('发送失败')
  }
}

const handleDeleteWebhook = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定删除 Webhook "${row.name}"？`, '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await adminApi.deleteWebhook(row.id)
    ElMessage.success('删除成功')
    loadWebhooks()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const resetWebhookForm = () => {
  webhookForm.name = ''
  webhookForm.url = ''
  webhookForm.secret = ''
  webhookForm.eventsArr = []
  webhookForm.enabled = true
}

// ---- 审计日志 ----
const auditLoading = ref(false)
const auditLogs = ref<any[]>([])
const auditTotal = ref(0)
const auditPage = ref(1)
const auditLimit = 20

const loadAuditLogs = async (page = auditPage.value) => {
  auditLoading.value = true
  try {
    const res = await adminApi.getAuditLogs(page, auditLimit)
    auditLogs.value = res.data.logs || []
    auditTotal.value = res.data.total || 0
    auditPage.value = page
  } catch {
    ElMessage.error('加载审计日志失败')
  } finally {
    auditLoading.value = false
  }
}

const auditActionType = (action: string): 'danger' | 'warning' | 'success' | 'primary' | 'info' => {
  if (action === 'login') return 'success'
  if (action === 'package_publish') return 'primary'
  if (action === 'package_unpublish') return 'danger'
  if (action === 'user_create') return 'warning'
  if (action === 'user_delete') return 'danger'
  if (action === 'config_update') return 'warning'
  return 'info'
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

onMounted(() => {
  loadSystemInfo()
  loadConfig()
  loadWebhooks()
  loadAuditLogs()
})

watch(activeTab, (tab) => {
  if (tab === 'audit') loadAuditLogs()
  if (tab === 'system') loadSystemInfo()
})
</script>

<style scoped>
.settings-page {
  padding: 20px 0;
}

.sys-desc {
  margin-bottom: 20px;
}

.webhook-header {
  margin-bottom: 16px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
