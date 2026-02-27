<template>
  <div class="gc-page">
    <!-- 统计概览 -->
    <el-row :gutter="20">
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="总包数" :value="stats.totalPackages" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="总版本数" :value="stats.totalVersions" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="总大小" :value="formatSize(stats.totalSize)" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="旧包（>180天）" :value="stats.oldPackages" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="旧包大小" :value="formatSize(stats.oldPackagesSize)" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="已废弃" :value="stats.deprecatedPackages" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 分析和清理 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 分析候选包 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>清理分析</span>
              <el-button type="primary" @click="analyzeGC" :loading="analyzing">
                分析候选包
              </el-button>
            </div>
          </template>

          <el-form :model="policy" label-width="120px" size="small">
            <el-form-item label="未访问天数">
              <el-input-number v-model="policy.maxInactiveDays" :min="30" :max="365" />
              <span class="form-hint">超过此天数未访问的包将被标记</span>
            </el-form-item>
            <el-form-item label="保留最小版本">
              <el-input-number v-model="policy.minVersionsToKeep" :min="1" :max="100" />
              <span class="form-hint">每个包最少保留的版本数</span>
            </el-form-item>
            <el-form-item label="包含已废弃">
              <el-switch v-model="policy.includeDeprecated" />
              <span class="form-hint">是否将已废弃的包纳入清理范围</span>
            </el-form-item>
          </el-form>

          <!-- 候选列表 -->
          <el-table
            v-if="candidates.length > 0"
            :data="candidates"
            style="width: 100%; margin-top: 16px"
            max-height="300"
          >
            <el-table-column prop="packageName" label="包名" width="200" />
            <el-table-column prop="lastAccessed" label="最后访问" width="150" />
            <el-table-column prop="accessCount" label="访问次数" width="100" />
            <el-table-column prop="size" label="大小" width="100">
              <template #default="scope">
                {{ formatSize(scope.row.size) }}
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" />
          </el-table>

          <el-empty v-else-if="!analyzing" description="点击分析查看可清理的包" />
        </el-card>
      </el-col>

      <!-- 执行清理 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span>执行清理</span>
          </template>

          <el-alert
            type="warning"
            :closable="false"
            style="margin-bottom: 16px"
          >
            <template #title>
              ⚠️ 清理操作不可逆，建议先使用「预览模式」查看将要删除的内容
            </template>
          </el-alert>

          <el-form label-width="120px">
            <el-form-item label="预览模式">
              <el-switch v-model="dryRun" />
              <span class="form-hint">开启后只显示将要删除的内容，不实际删除</span>
            </el-form-item>
          </el-form>

          <el-space>
            <el-button
              type="primary"
              :loading="running"
              @click="runGC"
            >
              <el-icon><Delete /></el-icon>
              {{ dryRun ? '预览清理' : '执行清理' }}
            </el-button>
          </el-space>

          <!-- 执行结果 -->
          <el-collapse v-if="result" style="margin-top: 16px">
            <el-collapse-item title="清理结果" name="result">
              <p><strong>删除包数:</strong> {{ result.deletedCount }}</p>
              <p><strong>释放空间:</strong> {{ formatSize(result.deletedSize) }}</p>
              <p v-if="result.deleted && result.deleted.length > 0"><strong>已删除:</strong></p>
              <ul v-if="result.deleted && result.deleted.length > 0">
                <li v-for="pkg in result.deleted" :key="pkg">{{ pkg }}</li>
              </ul>
              <p v-if="result.errors && result.errors.length > 0"><strong>错误:</strong></p>
              <ul v-if="result.errors && result.errors.length > 0">
                <li v-for="err in result.errors" :key="err" class="error-text">{{ err }}</li>
              </ul>
            </el-collapse-item>
          </el-collapse>
        </el-card>
      </el-col>
    </el-row>

    <!-- 使用说明 -->
    <el-card shadow="hover" style="margin-top: 20px">
      <template #header>
        <span>清理策略说明</span>
      </template>
      <el-collapse>
        <el-collapse-item title="什么是包清理 (GC)?" name="1">
          <p>随着时间推移，私有仓库会积累大量不再使用的包版本。包清理机制可以：</p>
          <ul>
            <li>识别长期未访问的包</li>
            <li>标记或删除不再需要的版本</li>
            <li>释放存储空间</li>
            <li>保持仓库整洁</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="清理规则" name="2">
          <ul>
            <li><strong>未访问天数:</strong> 超过指定天数未被下载/访问的包会被标记</li>
            <li><strong>保留最小版本:</strong> 每个包至少保留指定数量的版本</li>
            <li><strong>已废弃包:</strong> 默认不清理已废弃的包（可配置）</li>
            <li><strong>管理员权限:</strong> 只有管理员可以执行清理操作</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="推荐做法" name="3">
          <ul>
            <li>首先使用「预览模式」查看将要删除的内容</li>
            <li>对重要包使用「废弃标记」而非删除</li>
            <li>定期执行清理，建议每季度一次</li>
            <li>清理前先创建数据备份</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="定时清理" name="4">
          <p>可以配置定时任务自动执行清理：</p>
          <pre class="code-block"><code v-pre># 每月1号凌晨3点执行预览清理
0 3 1 * * curl -X POST http://localhost:4873/-/api/admin/gc/run \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dryRun":true,"maxInactiveDays":180}'</code></pre>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { adminApi } from '@/api'

interface GCStats {
  totalPackages: number
  totalVersions: number
  totalSize: number
  oldPackages: number
  oldPackagesSize: number
  deprecatedPackages: number
}

interface GCCandidate {
  packageName: string
  lastAccessed: string
  accessCount: number
  size: number
  isDeprecated: boolean
  reason: string
}

interface GCResult {
  ok: boolean
  dryRun: boolean
  deletedCount: number
  deletedSize: number
  deleted: string[]
  errors: string[]
}

const loading = ref(false)
const analyzing = ref(false)
const running = ref(false)
const dryRun = ref(true)

const stats = ref<GCStats>({
  totalPackages: 0,
  totalVersions: 0,
  totalSize: 0,
  oldPackages: 0,
  oldPackagesSize: 0,
  deprecatedPackages: 0,
})

const policy = reactive({
  maxInactiveDays: 180,
  minVersionsToKeep: 5,
  includeDeprecated: false,
})

const candidates = ref<GCCandidate[]>([])
const result = ref<GCResult | null>(null)

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const loadStats = async () => {
  loading.value = true
  try {
    const res = await adminApi.getGCStats()
    stats.value = res.data
  } catch {
    ElMessage.error('加载统计信息失败')
  } finally {
    loading.value = false
  }
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
    if (candidates.value.length === 0) {
      ElMessage.success('没有找到需要清理的包')
    }
  } catch {
    ElMessage.error('分析失败')
  } finally {
    analyzing.value = false
  }
}

const runGC = async () => {
  const action = dryRun.value ? '预览' : '执行'
  
  try {
    await ElMessageBox.confirm(
      dryRun.value
        ? '确定要预览清理结果吗？不会实际删除数据。'
        : '确定要执行清理吗？此操作不可逆！',
      `${action}清理`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: dryRun.value ? 'info' : 'warning',
      }
    )

    running.value = true
    result.value = null

    const res = await adminApi.runGC({
      dryRun: dryRun.value,
      maxInactiveDays: policy.maxInactiveDays,
      minVersionsToKeep: policy.minVersionsToKeep,
      includeDeprecated: policy.includeDeprecated,
    })

    result.value = res.data
    ElMessage.success(res.data.dryRun ? '预览完成' : '清理完成')
    
    // 刷新统计
    loadStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  } finally {
    running.value = false
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.gc-page {
  max-width: 1400px;
  margin: 0 auto;
}

.stat-card {
  text-align: center;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-hint {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}

.error-text {
  color: #f56c6c;
}

.code-block {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

.code-block code {
  font-family: monospace;
  font-size: 13px;
}
</style>
