<template>
  <div class="admin-layout">
    <!-- 左侧菜单 -->
    <aside class="admin-sidebar">
      <div class="sidebar-header">
        <el-icon :size="20"><Setting /></el-icon>
        <span>{{ $t('nav.admin') }}</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        router
        class="admin-menu"
        :collapse="isCollapse"
      >
        <!-- 仪表盘 -->
        <el-menu-item index="/admin/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>{{ $t('admin.dashboard') }}</template>
        </el-menu-item>

        <el-divider />

        <!-- 用户权限组 -->
        <div class="menu-group-title">{{ $t('admin.menu.users') }}</div>
        <el-menu-item index="/admin/users">
          <el-icon><User /></el-icon>
          <template #title>{{ $t('nav.users') }}</template>
        </el-menu-item>
        <el-menu-item index="/admin/tokens">
          <el-icon><Key /></el-icon>
          <template #title>{{ $t('nav.tokens') }}</template>
        </el-menu-item>

        <el-divider />

        <!-- 运维管理组 -->
        <div class="menu-group-title">{{ $t('admin.menu.operations') }}</div>
        <el-menu-item index="/admin/backup">
          <el-icon><Download /></el-icon>
          <template #title>{{ $t('nav.backup') }}</template>
        </el-menu-item>
        <el-menu-item index="/admin/gc">
          <el-icon><Delete /></el-icon>
          <template #title>{{ $t('nav.gc') }}</template>
        </el-menu-item>

        <el-divider />

        <!-- 系统配置组 -->
        <div class="menu-group-title">{{ $t('admin.menu.config') }}</div>
        <el-menu-item index="/admin/settings">
          <el-icon><Tools /></el-icon>
          <template #title>{{ $t('admin.basicSettings') }}</template>
        </el-menu-item>
        <el-menu-item index="/admin/upstreams">
          <el-icon><Link /></el-icon>
          <template #title>{{ $t('admin.upstreamConfig') }}</template>
        </el-menu-item>
        <el-menu-item index="/admin/system">
          <el-icon><InfoFilled /></el-icon>
          <template #title>{{ $t('admin.systemInfo') }}</template>
        </el-menu-item>

        <el-divider />

        <!-- 集成组 -->
        <div class="menu-group-title">{{ $t('admin.menu.integrations') }}</div>
        <el-menu-item index="/admin/webhooks">
          <el-icon><Connection /></el-icon>
          <template #title>{{ $t('nav.webhooks') }}</template>
        </el-menu-item>

        <el-divider />

        <!-- 监控组 -->
        <div class="menu-group-title">{{ $t('admin.menu.monitoring') }}</div>
        <el-menu-item index="/admin/audit-logs">
          <el-icon><Document /></el-icon>
          <template #title>{{ $t('admin.auditLogs') }}</template>
        </el-menu-item>
      </el-menu>

      <!-- 折叠按钮 -->
      <div class="sidebar-footer">
        <el-button text @click="isCollapse = !isCollapse">
          <el-icon>
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
        </el-button>
      </div>
    </aside>

    <!-- 右侧内容区 -->
    <main class="admin-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Setting,
  Odometer,
  User,
  Key,
  Download,
  Delete,
  Tools,
  Link,
  InfoFilled,
  Connection,
  Document,
  Fold,
  Expand,
} from '@element-plus/icons-vue'

const route = useRoute()
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: calc(100vh - 60px);
  overflow: hidden;
}

.admin-sidebar {
  width: 220px;
  min-width: 220px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
}

.admin-sidebar:has(.admin-menu.el-menu--collapse) {
  width: 64px;
  min-width: 64px;
}

.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  border-bottom: 1px solid #e4e7ed;
}

.admin-menu {
  flex: 1;
  border-right: none;
  padding: 8px 0;
  overflow-y: auto;
}

.menu-group-title {
  padding: 8px 20px;
  font-size: 12px;
  color: #909399;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

:deep(.el-menu-item) {
  height: 44px;
  line-height: 44px;
  margin: 2px 8px;
  border-radius: 6px;
}

:deep(.el-menu-item.is-active) {
  background: #ecf5ff;
}

:deep(.el-divider) {
  margin: 8px 16px;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid #e4e7ed;
  text-align: center;
}

.admin-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f5f7fa;
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>