<template>
  <div class="login-page">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <span class="grape-icon">🍇</span>
          <h2>{{ t('login.title') }}</h2>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            :placeholder="t('login.username')"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="t('login.password')"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            native-type="submit"
          >
            {{ t('login.submit') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
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

// 获取安全的重定向路径，防止开放重定向攻击
function getSafeRedirect(redirect: string | undefined): string {
  if (!redirect) return '/'
  // 只允许以 / 开头且不以 // 开头的相对路径
  if (redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }
  return '/'
}
</script>

<style scoped>
.login-page {
  min-height: calc(100vh - 140px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f8f5ff 0%, #fff 100%);
}

.login-card {
  width: 400px;
}

.login-header {
  text-align: center;
}

.grape-icon {
  font-size: 48px;
  display: block;
  margin-bottom: 8px;
}

.login-header h2 {
  margin: 0;
  color: var(--grape-primary);
}
</style>
