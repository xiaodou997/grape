<template>
  <div class="gc-page fade-in">
    <!-- Top Stats -->
    <div class="gc-stats-grid">
      <div class="gc-stat-box blue">
        <span class="gs-label">{{ t('gc.totalPackages') }}</span>
        <span class="gs-value">{{ stats.totalPackages }}</span>
      </div>
      <div class="gc-stat-box green">
        <span class="gs-label">{{ t('gc.totalVersions') }}</span>
        <span class="gs-value">{{ stats.totalVersions }}</span>
      </div>
      <div class="gc-stat-box orange">
        <span class="gs-label">Storage</span>
        <span class="gs-value">{{ formatSize(stats.totalSize) }}</span>
      </div>
      <div class="gc-stat-box purple">
        <span class="gs-label">{{ t('gc.deprecatedPackages') }}</span>
        <span class="gs-value">{{ stats.deprecatedPackages }}</span>
      </div>
    </div>

    <!-- Core Actions Row -->
    <div class="gc-actions-row">
      <!-- Analysis Card -->
      <section class="gc-main-card">
        <div class="gc-card-header">
          <div class="header-with-icon">
            <div class="icon-circle p-blue"><el-icon><Monitor /></el-icon></div>
            <h3>{{ t('gc.analyze') }}</h3>
          </div>
          <el-button type="primary" @click="analyzeGC" :loading="analyzing" size="small">{{ t('gc.analyze') }}</el-button>
        </div>
        
        <div class="card-body-flex">
          <el-form :model="policy" label-position="top" class="gc-form">
            <div class="form-row">
              <el-form-item :label="t('gc.maxInactiveDays')">
                <el-input-number v-model="policy.maxInactiveDays" :min="30" :max="365" class="w-full" />
              </el-form-item>
              <el-form-item :label="t('gc.minVersionsToKeep')">
                <el-input-number v-model="policy.minVersionsToKeep" :min="1" :max="100" class="w-full" />
              </el-form-item>
            </div>
            <el-form-item :label="t('gc.includeDeprecated')">
              <el-switch v-model="policy.includeDeprecated" />
            </el-form-item>
          </el-form>

          <div class="candidate-list">
            <el-table v-if="candidates.length > 0" :data="candidates" max-height="200" class="modern-table small">
              <el-table-column prop="packageName" :label="t('table.package')" min-width="140" />
              <el-table-column prop="size" :label="t('common.name')" width="100">
                <template #default="scope">{{ formatSize(scope.row.size) }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-else :image-size="40" :description="t('gc.dryRunHint')" />
          </div>
        </div>
      </section>

      <!-- Execution Card -->
      <section class="gc-main-card highlight-border">
        <div class="gc-card-header">
          <div class="header-with-icon">
            <div class="icon-circle p-red"><el-icon><Delete /></el-icon></div>
            <h3>{{ t('gc.runGC') }}</h3>
          </div>
        </div>
        
        <div class="card-body-flex">
          <div class="gc-execution-content">
            <el-alert type="warning" :closable="false" show-icon class="mb-24">
              <template #title>{{ t('gc.dryRunHint') }}</template>
            </el-alert>

            <div class="dry-run-toggle">
              <span>{{ t('gc.dryRun') }}</span>
              <el-switch v-model="dryRun" />
            </div>

            <el-button type="danger" :loading="running" @click="runGC" class="w-full gc-run-btn">
              <el-icon><Delete /></el-icon>
              {{ dryRun ? t('gc.dryRun') : t('gc.runGC') }}
            </el-button>

            <!-- Result Panel -->
            <div v-if="result" class="gc-result-panel animate-slide-up">
              <div class="res-item">
                <span class="res-label">{{ t('gc.packagesToDelete') }}</span>
                <span class="res-value">{{ result.deletedCount }}</span>
              </div>
              <div class="res-item">
                <span class="res-label">{{ t('gc.spaceToFree') }}</span>
                <span class="res-value">{{ formatSize(result.deletedSize) }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Instructions -->
    <section class="gc-instructions-section">
      <h3 class="section-title-v2">{{ t('gc.strategyTitle') }}</h3>
      <el-collapse class="modern-collapse">
        <el-collapse-item :title="t('gc.whatIsGC')" name="1">
          <p class="instruction-text">{{ t('gc.gcDescription') }}</p>
        </el-collapse-item>
        <el-collapse-item :title="t('gc.runGC')" name="4">
          <p class="instruction-text mb-12">Automate via Cron Job:</p>
          <pre class="code-block"><code v-pre># Run dry-run every 1st of month
0 3 1 * * curl -X POST http://localhost:4873/-/api/admin/gc/run \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"dryRun":true}'</code></pre>
        </el-collapse-item>
      </el-collapse>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Monitor } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

const { t } = useI18n()
const loading = ref(false)
const analyzing = ref(false)
const running = ref(false)
const dryRun = ref(true)

const stats = ref<any>({ totalPackages: 0, totalVersions: 0, totalSize: 0, deprecatedPackages: 0 })
const policy = reactive({ maxInactiveDays: 180, minVersionsToKeep: 5, includeDeprecated: false })
const candidates = ref<any[]>([])
const result = ref<any>(null)

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + ['B', 'KB', 'MB', 'GB'][i]
}

const loadStats = async () => {
  loading.value = true
  try {
    const res = await adminApi.getGCStats()
    stats.value = res.data
  } catch { ElMessage.error(t('errors.loadFailed')) }
  finally { loading.value = false }
}

const analyzeGC = async () => {
  analyzing.value = true
  candidates.value = []
  try {
    const res = await adminApi.analyzeGC({
      days: policy.maxInactiveDays,
      minVersions: policy.minVersionsToKeep,
      includeDeprecated: policy.includeDeprecated,
    })
    candidates.value = res.data.candidates || []
  } catch { ElMessage.error('Analysis failed') }
  finally { analyzing.value = false }
}

const runGC = async () => {
  try {
    await ElMessageBox.confirm(dryRun.value ? 'Preview changes?' : 'CAUTION: This is permanent!', t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: dryRun.value ? 'info' : 'warning',
    })
    running.value = true
    result.value = null
    const res = await adminApi.runGC({
      dryRun: dryRun.value,
      maxInactiveDays: policy.maxInactiveDays,
      minVersionsToKeep: policy.minVersionsToKeep,
      includeDeprecated: policy.includeDeprecated,
    })
    result.value = res.data
    ElMessage.success(t('gc.gcComplete'))
    loadStats()
  } catch (error) { if (error !== 'cancel') ElMessage.error('Action failed') }
  finally { running.value = false }
}

onMounted(loadStats)
</script>

<style scoped>
.gc-page { display: flex; flex-direction: column; gap: 32px; }
.gc-stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.gc-stat-box { background: white; padding: 20px; border-radius: 16px; border: 1px solid var(--g-border); display: flex; flex-direction: column; gap: 4px; }
.gs-label { font-size: 12px; color: var(--g-text-muted); font-weight: 600; text-transform: uppercase; }
.gs-value { font-size: 20px; font-weight: 700; color: var(--g-text-primary); }

.gc-actions-row { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; align-items: stretch; }
.gc-main-card { background: white; border-radius: 20px; border: 1px solid var(--g-border); padding: 32px; display: flex; flex-direction: column; }
.highlight-border { border-color: var(--g-brand-light); box-shadow: var(--shadow-sm); }
.gc-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-with-icon { display: flex; align-items: center; gap: 12px; }
.header-with-icon h3 { font-size: 18px; font-weight: 700; margin: 0; }

.icon-circle { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.p-blue { background: #eff6ff; color: #3b82f6; }
.p-red { background: #fef2f2; color: #ef4444; }

.card-body-flex { flex: 1; display: flex; flex-direction: column; }
.gc-form { margin-bottom: 20px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.w-full { width: 100%; }
.mb-24 { margin-bottom: 24px; }
.mb-12 { margin-bottom: 12px; }

.candidate-list { flex: 1; }

.dry-run-toggle { display: flex; justify-content: space-between; align-items: center; padding: 16px; background: var(--g-bg); border-radius: 12px; margin-bottom: 20px; }
.dry-run-toggle span { font-weight: 600; font-size: 14px; }

.gc-run-btn { height: 48px !important; font-weight: 700 !important; }

.gc-result-panel { margin-top: 24px; padding: 20px; background: var(--g-brand-light); border-radius: 16px; display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.res-item { display: flex; flex-direction: column; gap: 4px; }
.res-label { font-size: 12px; color: var(--g-brand); font-weight: 600; }
.res-value { font-size: 18px; font-weight: 700; color: var(--g-brand); }

.section-title-v2 { font-size: 18px; font-weight: 700; margin-bottom: 16px; }
.instruction-text { font-size: 14px; color: var(--g-text-secondary); line-height: 1.6; }

@media (max-width: 1024px) {
  .gc-stats-grid { grid-template-columns: repeat(2, 1fr); }
  .gc-actions-row { grid-template-columns: 1fr; }
}
</style>
