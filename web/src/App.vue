<template>
  <div class="app-container">
    <el-container>
      <!-- Header -->
      <el-header class="app-header">
        <div class="header-content">
          <router-link to="/" class="logo">
            <span class="logo-icon">🍇</span>
            <span class="logo-text">Grape</span>
          </router-link>
          
          <el-menu mode="horizontal" :ellipsis="false" class="nav-menu">
            <el-menu-item index="packages">
              <router-link to="/packages">包列表</router-link>
            </el-menu-item>
            <el-menu-item index="guide">
              <router-link to="/guide">使用指南</router-link>
            </el-menu-item>
            <el-menu-item index="admin">
              <router-link to="/admin">管理后台</router-link>
            </el-menu-item>
          </el-menu>

          <div class="header-right">
            <template v-if="userStore.isLoggedIn">
              <el-dropdown>
                <span class="user-dropdown">
                  <el-avatar :size="32" class="user-avatar">
                    {{ userStore.username?.charAt(0).toUpperCase() }}
                  </el-avatar>
                  <span class="username">{{ userStore.username }}</span>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
            <template v-else>
              <el-button type="primary" @click="$router.push('/login')">登录</el-button>
            </template>
          </div>
        </div>
      </el-header>

      <!-- Main Content -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>

      <!-- Footer -->
      <el-footer class="app-footer">
        <p>🍇 Grape - 轻盈如风的企业级私有 npm 仓库</p>
        <p>Powered by Go + Vue 3</p>
      </el-footer>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const handleLogout = () => {
  userStore.logout()
  window.location.href = '/'
}
</script>

<style scoped>
.app-header {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  height: 60px;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  max-width: 1400px;
  margin: 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  font-size: 20px;
  font-weight: 600;
  color: var(--grape-primary);
}

.logo-icon {
  font-size: 28px;
}

.nav-menu {
  border-bottom: none;
  flex: 1;
  justify-content: center;
}

.nav-menu .el-menu-item {
  padding: 0;
}

.nav-menu .el-menu-item a {
  display: block;
  padding: 0 20px;
  text-decoration: none;
  color: inherit;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.user-avatar {
  background-color: var(--grape-primary);
  color: white;
}

.username {
  color: var(--grape-text);
}

.app-footer {
  background: white;
  border-top: 1px solid #eee;
  text-align: center;
  padding: 20px;
  color: #666;
  font-size: 14px;
}

.app-footer p {
  margin: 4px 0;
}

/* Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>