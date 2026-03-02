<template>
  <div class="admin-container fade-in">
    <!-- Sidebar Wrapper -->
    <aside class="admin-sidebar" :class="{ 'is-collapsed': isCollapse }">
      <div class="sidebar-sticky-wrapper">
        <div class="sidebar-inner">
          <div class="sidebar-header">
            <div class="admin-badge">
              <el-icon><Setting /></el-icon>
            </div>
            <span v-if="!isCollapse" class="admin-title">Workspace</span>
          </div>

          <nav class="admin-nav">
            <!-- Main Section -->
            <div class="nav-section">
              <div v-if="!isCollapse" class="nav-section-title">{{ t('admin.dashboard') }}</div>
              <router-link to="/admin/dashboard" class="nav-item">
                <el-icon><Odometer /></el-icon>
                <span v-if="!isCollapse">{{ t('admin.dashboard') }}</span>
              </router-link>
            </div>

            <!-- Users & Auth -->
            <div class="nav-section">
              <div v-if="!isCollapse" class="nav-section-title">{{ t('admin.menu.users') }}</div>
              <router-link to="/admin/users" class="nav-item">
                <el-icon><User /></el-icon>
                <span v-if="!isCollapse">{{ t('nav.users') }}</span>
              </router-link>
              <router-link to="/admin/tokens" class="nav-item">
                <el-icon><Key /></el-icon>
                <span v-if="!isCollapse">{{ t('nav.tokens') }}</span>
              </router-link>
            </div>

            <!-- Operations -->
            <div class="nav-section">
              <div v-if="!isCollapse" class="nav-section-title">{{ t('admin.menu.operations') }}</div>
              <router-link to="/admin/backup" class="nav-item">
                <el-icon><Download /></el-icon>
                <span v-if="!isCollapse">{{ t('nav.backup') }}</span>
              </router-link>
              <router-link to="/admin/gc" class="nav-item">
                <el-icon><Delete /></el-icon>
                <span v-if="!isCollapse">{{ t('nav.gc') }}</span>
              </router-link>
            </div>

            <!-- System -->
            <div class="nav-section">
              <div v-if="!isCollapse" class="nav-section-title">{{ t('admin.menu.config') }}</div>
              <router-link to="/admin/settings" class="nav-item">
                <el-icon><Tools /></el-icon>
                <span v-if="!isCollapse">{{ t('admin.basicSettings') }}</span>
              </router-link>
              <router-link to="/admin/upstreams" class="nav-item">
                <el-icon><Link /></el-icon>
                <span v-if="!isCollapse">{{ t('admin.upstreamConfig') }}</span>
              </router-link>
            </div>

            <!-- Monitoring -->
            <div class="nav-section">
              <div v-if="!isCollapse" class="nav-section-title">{{ t('admin.menu.monitoring') }}</div>
              <router-link to="/admin/webhooks" class="nav-item">
                <el-icon><Connection /></el-icon>
                <span v-if="!isCollapse">{{ t('nav.webhooks') }}</span>
              </router-link>
              <router-link to="/admin/audit-logs" class="nav-item">
                <el-icon><Document /></el-icon>
                <span v-if="!isCollapse">{{ t('admin.auditLogs') }}</span>
              </router-link>
            </div>
          </nav>

          <div class="sidebar-footer">
            <button class="collapse-btn" @click="isCollapse = !isCollapse">
              <el-icon>
                <Fold v-if="!isCollapse" />
                <Expand v-else />
              </el-icon>
            </button>
          </div>
        </div>
      </div>
    </aside>

    <!-- Content Area -->
    <main class="admin-main">
      <div class="admin-content-card">
        <div class="card-inner-header">
          <h2 class="card-inner-title">{{ activeTitle }}</h2>
        </div>
        <div class="card-inner-content">
          <router-view v-slot="{ Component }">
            <transition name="admin-fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Setting, Odometer, User, Key, Download, Delete,
  Tools, Link, Connection, Document, Fold, Expand
} from '@element-plus/icons-vue'

const { t } = useI18n()
const route = useRoute()
const isCollapse = ref(false)

const activeTitle = computed(() => {
  const path = route.path
  if (path.includes('dashboard')) return t('admin.dashboard')
  if (path.includes('users')) return t('nav.users')
  if (path.includes('tokens')) return t('nav.tokens')
  if (path.includes('backup')) return t('nav.backup')
  if (path.includes('gc')) return t('nav.gc')
  if (path.includes('settings')) return t('admin.basicSettings')
  if (path.includes('upstreams')) return t('admin.upstreamConfig')
  if (path.includes('webhooks')) return t('nav.webhooks')
  if (path.includes('audit-logs')) return t('admin.auditLogs')
  return t('nav.admin')
})
</script>

<style scoped>
.admin-container {
  display: flex;
  background-color: var(--g-bg);
  min-height: calc(100vh - 64px);
  padding: 32px;
  gap: 32px;
  align-items: stretch; /* 关键：强制左右对齐 */
}

.admin-sidebar {
  width: 240px;
  flex-shrink: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.admin-sidebar.is-collapsed {
  width: 72px;
}

.sidebar-sticky-wrapper {
  position: sticky;
  top: 96px; /* 64px header + 32px padding */
}

.sidebar-inner {
  background: white;
  border-radius: var(--radius-xl);
  border: 1px solid var(--g-border);
  display: flex;
  flex-direction: column;
  height: calc(100vh - 128px); /* 100vh - 64 header - 32*2 padding */
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.sidebar-header {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--g-border);
}

.admin-badge {
  width: 32px;
  height: 32px;
  background: var(--g-brand-light);
  color: var(--g-brand);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.admin-title {
  font-weight: 700;
  font-size: 15px;
  color: var(--g-text-primary);
  letter-spacing: -0.3px;
}

.admin-nav {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: 16px;
}

.nav-section-title {
  padding: 0 12px 8px;
  font-size: 10px;
  font-weight: 700;
  color: var(--g-text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  color: var(--g-text-secondary);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  border-radius: 10px;
  margin-bottom: 2px;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: var(--g-bg);
  color: var(--g-brand);
}

.nav-item.router-link-active {
  background: var(--g-brand-light);
  color: var(--g-brand);
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--g-border);
}

.collapse-btn {
  width: 100%;
  padding: 8px;
  border: none;
  background: var(--g-bg);
  color: var(--g-text-secondary);
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.admin-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.admin-content-card {
  flex: 1;
  background: white;
  border-radius: var(--radius-xl);
  border: 1px solid var(--g-border);
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.card-inner-header {
  padding: 24px 32px;
  border-bottom: 1px solid var(--g-border);
  background: #fafafa;
}

.card-inner-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--g-text-primary);
  margin: 0;
  letter-spacing: -0.5px;
}

.card-inner-content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
}

/* Transitions */
.admin-fade-enter-active,
.admin-fade-leave-active {
  transition: all 0.2s ease;
}

.admin-fade-enter-from { opacity: 0; transform: translateY(4px); }
.admin-fade-leave-to { opacity: 0; transform: translateY(-4px); }

@media (max-width: 1024px) {
  .admin-container { flex-direction: column; padding: 16px; }
  .admin-sidebar { width: 100% !important; }
  .sidebar-sticky-wrapper { position: static; }
  .sidebar-inner { height: auto; }
  .admin-nav { display: flex; overflow-x: auto; padding: 12px; gap: 8px; }
  .nav-section { margin-bottom: 0; }
  .nav-section-title { display: none; }
}
</style>
