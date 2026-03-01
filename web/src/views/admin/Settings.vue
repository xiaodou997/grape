<template>
  <div class="settings-page">
    <el-form label-width="140px" v-loading="loading">
      <!-- 认证配置 -->
      <el-card shadow="hover" class="setting-section">
        <template #header>
          <div class="section-header">
            <el-icon><Lock /></el-icon>
            <span>{{ $t('settings.auth') }}</span>
          </div>
        </template>
        <el-form-item :label="$t('settings.jwtSecret')">
          <el-input
            v-model="form.auth.jwtSecret"
            type="password"
            show-password
            style="max-width: 400px"
            :placeholder="$t('settings.jwtSecretPlaceholder')"
          />
          <div class="form-hint">{{ $t('settings.jwtSecretHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('settings.jwtExpiry')">
          <el-input-number v-model="form.auth.jwtExpiry" :min="1" :max="8760" />
          <span class="unit">{{ $t('common.hours') }}</span>
        </el-form-item>
        <el-form-item :label="$t('settings.allowRegistration')">
          <el-switch v-model="form.auth.allowRegistration" />
          <div class="form-hint">{{ $t('settings.allowRegistrationHint') }}</div>
        </el-form-item>
      </el-card>

      <!-- 日志配置 -->
      <el-card shadow="hover" class="setting-section">
        <template #header>
          <div class="section-header">
            <el-icon><Document /></el-icon>
            <span>{{ $t('settings.log') }}</span>
          </div>
        </template>
        <el-form-item :label="$t('settings.logLevel')">
          <el-select v-model="form.log.level" style="width: 200px">
            <el-option label="debug" value="debug" />
            <el-option label="info" value="info" />
            <el-option label="warn" value="warn" />
            <el-option label="error" value="error" />
          </el-select>
        </el-form-item>
      </el-card>

      <!-- 保存按钮 -->
      <div class="form-actions">
        <el-button type="primary" @click="saveConfig" :loading="saving">
          {{ $t('common.save') }}
        </el-button>
        <el-button @click="loadConfig">{{ $t('common.reset') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Lock, Document } from '@element-plus/icons-vue'
import { adminApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)

const form = reactive({
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
  loading.value = true
  try {
    const res = await adminApi.getConfig()
    const data = res.data
    if (data.auth) {
      form.auth.jwtSecret = ''
      form.auth.jwtExpiry = data.auth.jwtExpiry || 24
      form.auth.allowRegistration = !!data.auth.allowRegistration
    }
    if (data.log) {
      form.log.level = data.log.level || 'info'
    }
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const payload: any = {
      auth: {
        jwtExpiry: form.auth.jwtExpiry,
        allowRegistration: form.auth.allowRegistration,
      },
      log: {
        level: form.log.level,
      },
    }
    if (form.auth.jwtSecret) {
      payload.auth.jwtSecret = form.auth.jwtSecret
    }
    await adminApi.updateConfig(payload)
    ElMessage.success(t('common.saveSuccess'))
    form.auth.jwtSecret = ''
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || t('errors.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.settings-page {
  padding: 0;
}

.setting-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

.unit {
  margin-left: 8px;
  color: #606266;
}

.form-actions {
  margin-top: 30px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: center;
}
</style>