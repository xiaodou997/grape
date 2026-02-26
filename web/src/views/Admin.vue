<template>
  <div class="admin-page fade-in">
    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="概览" name="overview">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="总包数" :value="stats.totalPackages" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="私有包" :value="stats.privatePackages" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="用户数" :value="stats.users" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <el-statistic title="今日下载" :value="stats.todayDownloads" />
            </el-card>
          </el-col>
        </el-row>

        <el-card class="recent-activity" style="margin-top: 20px">
          <template #header>
            <span>最近活动</span>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="activity in recentActivities"
              :key="activity.id"
              :timestamp="activity.time"
              placement="top"
            >
              <el-card>
                <h4>{{ activity.action }}</h4>
                <p>{{ activity.description }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="用户管理" name="users">
        <router-view />
      </el-tab-pane>

      <el-tab-pane label="系统设置" name="settings">
        <router-view />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const activeTab = ref('overview')

const stats = reactive({
  totalPackages: 125,
  privatePackages: 23,
  users: 8,
  todayDownloads: 1567,
})

const recentActivities = ref([
  { id: 1, action: '包发布', description: 'admin 发布了 @company/utils@1.2.0', time: '2024-01-15 10:30' },
  { id: 2, action: '用户登录', description: 'developer 登录了系统', time: '2024-01-15 09:15' },
  { id: 3, action: '包下载', description: 'lodash 被下载了 15 次', time: '2024-01-14 18:00' },
])
</script>

<style scoped>
.admin-page {
  max-width: 1200px;
  margin: 0 auto;
}

.stat-card {
  text-align: center;
}

.recent-activity {
  max-height: 500px;
  overflow-y: auto;
}
</style>
