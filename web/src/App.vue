<template>
  <div class="app-container">
    <el-container>
      <!-- Modern Sticky Header -->
      <el-header class="app-header glass-panel">
        <div class="header-content">
          <router-link to="/" class="brand-logo">
            <div class="logo-circle">🍇</div>
            <span class="logo-text">Grape</span>
          </router-link>

          <nav class="main-nav">
            <router-link to="/packages" class="nav-link">{{ t('nav.packages') }}</router-link>
            <router-link to="/guide" class="nav-link">{{ t('nav.guide') }}</router-link>
            <router-link to="/admin" class="nav-link">{{ t('nav.admin') }}</router-link>
          </nav>

          <div class="header-actions">
            <LanguageSwitch />
            
            <div class="auth-area">
              <template v-if="userStore.isLoggedIn">
                <el-dropdown trigger="click">
                  <div class="user-pill">
                    <el-avatar :size="28" class="user-avatar">
                      {{ userStore.username?.charAt(0).toUpperCase() }}
                    </el-avatar>
                    <span class="username">{{ userStore.username }}</span>
                    <el-icon><ArrowDown /></el-icon>
                  </div>
                  <template #dropdown>
                    <el-dropdown-menu class="user-menu-dropdown">
                      <el-dropdown-item @click="showChangePassword = true">
                        <el-icon><Lock /></el-icon> {{ t('nav.changePassword') }}
                      </el-dropdown-item>
                      <el-dropdown-item divided @click="handleLogout" class="logout-item">
                        <el-icon><SwitchButton /></el-icon> {{ t('nav.logout') }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
              <template v-else>
                <el-button type="primary" plain @click="$router.push('/login')">
                  {{ t('nav.login') }}
                </el-button>
              </template>
            </div>
          </div>
        </div>
      </el-header>

      <!-- Change Password Dialog -->
      <el-dialog
        v-model="showChangePassword"
        :title="t('users.changePassword')"
        width="420px"
        class="modern-dialog"
      >
        <el-form
          ref="pwdFormRef"
          :model="pwdForm"
          :rules="pwdRules"
          label-position="top"
        >
          <el-form-item :label="t('users.oldPassword')" prop="oldPassword">
            <el-input
              v-model="pwdForm.oldPassword"
              type="password"
              show-password
              :placeholder="t('users.oldPasswordRequired')"
            />
          </el-form-item>
          <el-form-item :label="t('users.newPassword')" prop="newPassword">
            <el-input
              v-model="pwdForm.newPassword"
              type="password"
              show-password
              :placeholder="t('users.passwordMinLength')"
            />
          </el-form-item>
          <el-form-item
            :label="t('users.confirmPassword')"
            prop="confirmPassword"
          >
            <el-input
              v-model="pwdForm.confirmPassword"
              type="password"
              show-password
              :placeholder="t('users.confirmPasswordRequired')"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="showChangePassword = false">{{ t('common.cancel') }}</el-button>
            <el-button
              type="primary"
              @click="handleChangePassword"
              :loading="pwdLoading"
            >{{ t('common.save') }}</el-button>
          </div>
        </template>
      </el-dialog>

      <!-- Main Content -->
      <el-main class="main-content-wrapper">
        <router-view v-slot="{ Component }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>

      <!-- Footer -->
      <footer class="modern-footer">
        <div class="footer-content">
          <div class="footer-logo">🍇 Grape Registry</div>
          <p class="footer-tagline">{{ t('nav.footerTagline') }}</p>
          <div class="footer-meta">
            <span>v0.1.0</span>
            <span class="dot">·</span>
            <span>Powered by Go & Vue 3</span>
          </div>
        </div>
      </footer>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ArrowDown, Lock, SwitchButton } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { adminApi } from '@/api'
import LanguageSwitch from '@/components/LanguageSwitch.vue'

const { t } = useI18n()
const userStore = useUserStore()

const handleLogout = () => {
  userStore.logout()
  window.location.href = '/'
}

// Change password logic remains the same
const showChangePassword = ref(false)
const pwdLoading = ref(false)
const pwdFormRef = ref<FormInstance>()
const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const pwdRules: FormRules = {
  oldPassword: [{ required: true, message: t('users.oldPasswordRequired'), trigger: 'blur' }],
  newPassword: [
    { required: true, message: t('users.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('users.passwordMinLength'), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: t('users.confirmPasswordRequired'), trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: Function) => {
        if (value !== pwdForm.newPassword) {
          callback(new Error(t('users.passwordMismatch')))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const handleChangePassword = async () => {
  const valid = await pwdFormRef.value?.validate()
  if (!valid) return

  pwdLoading.value = true
  try {
    await adminApi.updateUser(userStore.username!, { password: pwdForm.newPassword })
    ElMessage.success(t('users.passwordChanged'))
    showChangePassword.value = false
    pwdFormRef.value?.resetFields()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || t('users.passwordChangeFailed'))
  } finally {
    pwdLoading.value = false
  }
}
</script>

<style scoped>
.app-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  height: 64px !important;
  display: flex;
  align-items: center;
  padding: 0 40px;
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--g-border) !important;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  transition: transform 0.2s ease;
}

.brand-logo:hover {
  transform: scale(1.02);
}

.logo-circle {
  width: 36px;
  height: 36px;
  background: var(--g-brand-light);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  box-shadow: inset 0 0 0 1px rgba(124, 58, 237, 0.1);
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--g-brand), #9333ea);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

.main-nav {
  display: flex;
  gap: 8px;
}

.nav-link {
  padding: 8px 16px;
  text-decoration: none;
  color: var(--g-text-secondary);
  font-weight: 500;
  font-size: 14px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.nav-link:hover {
  color: var(--g-brand);
  background: var(--g-brand-light);
}

.nav-link.router-link-active {
  color: var(--g-brand);
  background: var(--g-brand-light);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-pill {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px 6px 6px;
  background: var(--g-bg);
  border-radius: 30px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.user-pill:hover {
  border-color: var(--g-brand);
  background: white;
  box-shadow: var(--shadow-sm);
}

.user-avatar {
  background-color: var(--g-brand);
  font-weight: 600;
  border: 2px solid white;
}

.username {
  font-size: 14px;
  font-weight: 500;
  color: var(--g-text-primary);
}

.user-menu-dropdown {
  padding: 8px;
  border-radius: 12px !important;
}

.logout-item {
  color: var(--g-error) !important;
}

.main-content-wrapper {
  padding: 0 !important;
  background-color: var(--g-bg);
  min-height: calc(100vh - 64px - 200px);
}

.modern-footer {
  background: white;
  padding: 60px 40px;
  border-top: 1px solid var(--g-border);
  text-align: center;
}

.footer-content {
  max-width: 600px;
  margin: 0 auto;
}

.footer-logo {
  font-size: 18px;
  font-weight: 700;
  color: var(--g-text-primary);
  margin-bottom: 12px;
}

.footer-tagline {
  color: var(--g-text-secondary);
  font-size: 14px;
  margin-bottom: 24px;
}

.footer-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--g-text-muted);
  font-size: 12px;
}

.dot {
  font-weight: bold;
}

/* Page Transitions */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1), 
              transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
