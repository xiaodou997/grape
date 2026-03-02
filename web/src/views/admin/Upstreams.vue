<template>
  <div class="upstreams-page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>{{ $t('settings.upstreams') }}</span>
          <el-button type="primary" @click="addUpstream">
            <el-icon><Plus /></el-icon>
            {{ $t('settings.addUpstream') }}
          </el-button>
        </div>
      </template>

      <el-table :data="upstreams" border v-loading="loading">
        <el-table-column :label="$t('common.name')" width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" size="small" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('settings.upstreamUrl')" min-width="300">
          <template #default="{ row }">
            <el-input v-model="row.url" size="small" placeholder="https://registry.npmjs.org" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('settings.upstreamScope')" width="150">
          <template #default="{ row }">
            <el-input v-model="row.scope" size="small" placeholder="@scope" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('settings.upstreamTimeout')" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.timeout" :min="1" :max="300" size="small" style="width: 90px" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.actions')" width="100" fixed="right">
          <template #default="{ $index }">
            <el-button text type="danger" size="small" @click="removeUpstream($index)">
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="upstreams.length === 0" :description="$t('settings.noUpstreams')" />

      <div class="form-actions">
        <el-button type="primary" @click="saveConfig" :loading="saving">
          {{ $t('common.save') }}
        </el-button>
        <el-button @click="loadConfig">{{ $t('common.reset') }}</el-button>
      </div>
    </el-card>

    <!-- 说明 -->
    <el-card shadow="hover" style="margin-top: 20px">
      <template #header>
        <span>{{ $t('common.tip') }}</span>
      </template>
      <el-collapse>
        <el-collapse-item :title="$t('settings.upstreamHelpTitle')" name="1">
          <ul class="help-list">
            <li><strong>{{ $t('settings.upstreamName') }}</strong>: {{ $t('settings.upstreamNameHelp') }}</li>
            <li><strong>{{ $t('settings.upstreamUrl') }}</strong>: {{ $t('settings.upstreamUrlHelp') }}</li>
            <li><strong>{{ $t('settings.upstreamScope') }}</strong>: {{ $t('settings.upstreamScopeHelp') }}</li>
            <li><strong>{{ $t('settings.upstreamTimeout') }}</strong>: {{ $t('settings.upstreamTimeoutHelp') }}</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item :title="$t('settings.upstreamExampleTitle')" name="2">
          <pre class="code-block"><code># {{ $t('settings.upstreamExample1') }}
名称: npm
URL: https://registry.npmjs.org
Scope: ({{ $t('common.empty') }})

# {{ $t('settings.upstreamExample2') }}
名称: taobao
URL: https://registry.npmmirror.com
Scope: ({{ $t('common.empty') }})

# {{ $t('settings.upstreamExample3') }}
名称: company
URL: https://npm.company.com
Scope: @company</code></pre>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '@/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)

interface Upstream {
  name: string
  url: string
  scope: string
  timeout: number
  enabled: boolean
}

const upstreams = ref<Upstream[]>([])

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await adminApi.getConfig()
    const data = res.data
    upstreams.value = (data.registry?.upstreams || []).map((u: any) => ({
      name: u.name || '',
      url: u.url || '',
      scope: u.scope || '',
      timeout: u.timeout || 30,
      enabled: u.enabled !== false,
    }))
  } catch {
    ElMessage.error(t('errors.loadFailed'))
  } finally {
    loading.value = false
  }
}

const addUpstream = () => {
  upstreams.value.push({ name: '', url: '', scope: '', timeout: 30, enabled: true })
}

const removeUpstream = (index: number) => {
  upstreams.value.splice(index, 1)
}

const saveConfig = async () => {
  saving.value = true
  try {
    const payload = {
      registry: {
        upstreams: upstreams.value,
      },
    }
    await adminApi.updateConfig(payload)
    ElMessage.success(t('common.saveSuccess'))
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
.upstreams-page {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-actions {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
  text-align: center;
}

.help-list {
  line-height: 2;
}

.help-list li {
  margin-bottom: 8px;
}

.code-block {
  background: #0f172a;
  color: #e2e8f0;
  padding: 20px;
  border-radius: 12px;
  overflow-x: auto;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.code-block code {
  color: #38bdf8;
}
</style>
