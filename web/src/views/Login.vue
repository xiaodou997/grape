<template>
  <div class="login-container">
    <div class="login-bg">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <div class="login-content animate-slide-up">
      <div class="login-card glass-panel">
        <div class="login-header">
          <div class="logo-wrapper">
            <span class="logo-emoji">🍇</span>
          </div>
          <h1 class="login-title">{{ t('login.title') }}</h1>
          <p class="login-subtitle">{{ t('login.subtitle') }}</p>
        </div>

        <el-form 
          ref="formRef" 
          :model="form" 
          :rules="rules" 
          @submit.prevent="handleLogin"
          label-position="top"
          class="modern-form"
        >
          <el-form-item :label="t('login.username')" prop="username">
            <el-input
              v-model="form.username"
              :placeholder="t('login.usernameRequired')"
              size="large"
              :prefix-icon="User"
            />
          </el-form-item>

          <el-form-item :label="t('login.password')" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              :placeholder="t('login.passwordRequired')"
              size="large"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <div class="form-options">
            <el-checkbox v-model="rememberMe">{{ t('login.rememberMe') }}</el-checkbox>
            <el-link type="primary" :underline="false" size="small">{{ t('login.forgotPassword') }}</el-link>
          </div>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-submit-btn"
              native-type="submit"
            >
              {{ t('login.submit') }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <p>{{ t('login.welcomeBack') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { FormInstance, FormRules } from 'element-plus'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(true)

const form = reactive({
  username: '',
  password: '',
})

const rules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }],
}))

const handleLogin = async () => {
  const valid = await formRef.value?.validate()
  if (!valid) return

  loading.value = true
  try {
    const success = await userStore.login(form.username, form.password)
    if (success) {
      ElMessage.success(t('login.success'))
      const redirect = getSafeRedirect(route.query.redirect as string)
      router.push(redirect)
    } else {
      ElMessage.error(t('login.error'))
    }
  } catch {
    ElMessage.error(t('login.errorRetry'))
  } finally {
    loading.value = false
  }
}

function getSafeRedirect(redirect: string | undefined): string {
  if (!redirect) return '/'
  if (redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }
  return '/'
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background-color: #f8fafc;
}

.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

.blob {
  position: absolute;
  filter: blur(80px);
  border-radius: 50%;
  opacity: 0.6;
}

.blob-1 {
  width: 500px;
  height: 500px;
  background: rgba(124, 58, 237, 0.2);
  top: -100px;
  left: -100px;
}

.blob-2 {
  width: 400px;
  height: 400px;
  background: rgba(16, 185, 129, 0.15);
  bottom: -50px;
  right: -50px;
}

.blob-3 {
  width: 300px;
  height: 300px;
  background: rgba(59, 130, 246, 0.1);
  top: 40%;
  left: 30%;
}

.login-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 460px;
  padding: 20px;
}

.login-card {
  padding: 48px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  box-shadow: var(--shadow-lg) !important;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo-wrapper {
  width: 64px;
  height: 64px;
  background: white;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  box-shadow: var(--shadow-md);
  font-size: 32px;
}

.login-title {
  font-size: 28px;
  font-weight: 800;
  color: var(--g-text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.5px;
}

.login-subtitle {
  font-size: 15px;
  color: var(--g-text-secondary);
}

.modern-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--g-text-primary);
  font-size: 13px;
  margin-bottom: 8px;
}

.modern-form :deep(.el-input__wrapper) {
  background-color: rgba(255, 255, 255, 0.5);
  box-shadow: 0 0 0 1px var(--g-border) inset;
  transition: all 0.2s ease;
}

.modern-form :deep(.el-input__wrapper.is-focus) {
  background-color: white;
  box-shadow: 0 0 0 1px var(--g-brand) inset, 0 0 0 4px var(--g-brand-light) !important;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.login-submit-btn {
  width: 100%;
  height: 48px !important;
  font-size: 16px !important;
  font-weight: 700 !important;
  box-shadow: 0 10px 15px -3px rgba(124, 58, 237, 0.3);
}

.login-footer {
  margin-top: 32px;
  text-align: center;
  font-size: 14px;
  color: var(--g-text-muted);
}
</style>
