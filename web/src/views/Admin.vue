<template>
  <div class="admin-page fade-in">
    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="概览" name="overview">
        <el-row :gutter="20" v-loading="loading">
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="总包数" :value="stats.totalPackages" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="存储占用" :value="stats.storageSize" suffix="MB" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="用户数" :value="stats.users" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="服务状态" value="运行中" />
            </el-card>
          </el-col>
        </el-row>

        <el-card class="quick-actions" style="margin-top: 20px">
          <template #header>
            <span>快速操作</span>
          </template>
          <el-space wrap>
            <el-button type="primary" @click="$router.push('/packages')">
              浏览包
            </el-button>
            <el-button @click="activeTab = 'users'">
              用户管理
            </el-button>
            <el-button @click="activeTab = 'tokens'">
              Token 管理
            </el-button>
            <el-button @click="activeTab = 'backup'">
              备份恢复
            </el-button>
            <el-button @click="activeTab = 'gc'">
              包清理
            </el-button>
            <el-button @click="copyRegistryUrl">
              复制 registry 地址
            </el-button>
          </el-space>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="用户管理" name="users">
        <UsersManager />
      </el-tab-pane>

      <el-tab-pane label="Token 管理" name="tokens">
        <TokenManager />
      </el-tab-pane>

      <el-tab-pane label="备份恢复" name="backup">
        <BackupManager />
      </el-tab-pane>

      <el-tab-pane label="包清理" name="gc">
        <GCManager />
      </el-tab-pane>

      <el-tab-pane label="系统管理" name="settings">
        <SystemSettings />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'
import UsersManager from './admin/Users.vue'
import TokenManager from './admin/Tokens.vue'
import BackupManager from './admin/Backup.vue'
import GCManager from './admin/GC.vue'
import SystemSettings from './admin/Settings.vue'

const activeTab = ref('overview')
const loading = ref(false)

const stats = reactive({
  totalPackages: 0,
  storageSize: 0,
  users: 0,
})

const loadStats = async () => {
  loading.value = true
  try {
    const [statsRes, usersRes] = await Promise.all([
      adminApi.getStats(),
      adminApi.getUsers(),
    ])

    stats.totalPackages = statsRes.data.totalPackages || 0
    stats.storageSize = statsRes.data.storageSize || 0
    stats.users = (usersRes.data.users || []).length
  } catch {
    // 忽略加载错误
  } finally {
    loading.value = false
  }
}

const copyRegistryUrl = () => {
  const url = `${window.location.protocol}//${window.location.host}`
  navigator.clipboard.writeText(url)
  ElMessage.success(`已复制: ${url}`)
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.admin-page {
  max-width: 1200px;
  margin: 0 auto;
}

.stat-card {
  text-align: center;
}

.server-info {
  max-height: 500px;
  overflow-y: auto;
}

.quick-actions {
  padding: 10px 0;
}
</style>
