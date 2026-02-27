<template>
  <div class="tokens-page">
    <div class="page-header">
      <h2>CI/CD Tokens</h2>
      <p class="description">管理用于自动化发布的持久化 Token。这些 Token 可用于 CI/CD 流水线，无需交互式登录。</p>
    </div>

    <div class="actions">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建 Token
      </el-button>
    </div>

    <!-- Token 列表 -->
    <el-table :data="tokens" v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.readonly ? 'info' : 'success'" size="small">
            {{ row.readonly ? '只读' : '可发布' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="过期时间" width="180">
        <template #default="{ row }">
          <span v-if="row.expiresAt">{{ formatDate(row.expiresAt) }}</span>
          <span v-else class="never-expires">永不过期</span>
        </template>
      </el-table-column>
      <el-table-column label="最后使用" width="180">
        <template #default="{ row }">
          <span v-if="row.lastUsed">{{ formatDate(row.lastUsed) }}</span>
          <span v-else class="never-used">从未使用</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="danger" size="small" text @click="confirmDelete(row)">
            <el-icon><Delete /></el-icon>
            撤销
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建 Token 对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建 Token" width="500px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="如: github-ci" />
          <div class="form-tip">用于识别 Token 用途的描述性名称</div>
        </el-form-item>
        <el-form-item label="权限类型">
          <el-radio-group v-model="createForm.readonly">
            <el-radio :value="false">可发布</el-radio>
            <el-radio :value="true">只读</el-radio>
          </el-radio-group>
          <div class="form-tip">
            只读 Token 只能下载包，不能发布或删除包
          </div>
        </el-form-item>
        <el-form-item label="有效期">
          <el-select v-model="createForm.days" placeholder="选择有效期" style="width: 100%">
            <el-option label="永不过期" :value="0" />
            <el-option label="7 天" :value="7" />
            <el-option label="30 天" :value="30" />
            <el-option label="90 天" :value="90" />
            <el-option label="365 天" :value="365" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createToken" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 显示新创建的 Token -->
    <el-dialog v-model="showTokenDialog" title="Token 已创建" width="550px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" show-icon style="margin-bottom: 20px">
        <template #title>
          <strong>请立即复制 Token</strong>
        </template>
        <p style="margin: 8px 0 0 0">Token 只会显示一次，关闭此对话框后将无法再次查看。</p>
      </el-alert>
      
      <div class="token-display">
        <code>{{ newToken }}</code>
        <el-button type="primary" size="small" @click="copyToken">
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </div>

      <div class="usage-example">
        <h4>使用示例</h4>
        <p>在 CI/CD 环境中配置：</p>
        <pre v-pre><code># GitHub Actions 示例
npm config set //your-registry.com/:_authToken ${{ secrets.GRAPE_TOKEN }}</code></pre>
      </div>

      <template #footer>
        <el-button type="primary" @click="showTokenDialog = false">我已复制，关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, DocumentCopy } from '@element-plus/icons-vue'
import { tokenApi } from '@/api'

interface Token {
  id: number
  name: string
  readonly: boolean
  expiresAt: string | null
  lastUsed: string | null
  createdAt: string
}

const tokens = ref<Token[]>([])
const loading = ref(false)
const createDialogVisible = ref(false)
const showTokenDialog = ref(false)
const creating = ref(false)
const newToken = ref('')

const createForm = ref({
  name: '',
  readonly: false,
  days: 0
})

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const loadTokens = async () => {
  loading.value = true
  try {
    const res = await tokenApi.list()
    tokens.value = res.data.objects || []
  } catch (error) {
    ElMessage.error('加载 Token 列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  createForm.value = { name: '', readonly: false, days: 0 }
  createDialogVisible.value = true
}

const createToken = async () => {
  if (!createForm.value.name.trim()) {
    ElMessage.warning('请输入 Token 名称')
    return
  }

  creating.value = true
  try {
    const res = await tokenApi.create({
      name: createForm.value.name,
      readonly: createForm.value.readonly,
      days: createForm.value.days || undefined
    })
    
    newToken.value = res.data.token
    createDialogVisible.value = false
    showTokenDialog.value = true
    loadTokens()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '创建 Token 失败')
  } finally {
    creating.value = false
  }
}

const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(newToken.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

const confirmDelete = (token: Token) => {
  ElMessageBox.confirm(
    `确定要撤销 Token "${token.name}" 吗？撤销后使用该 Token 的 CI/CD 流水线将无法继续工作。`,
    '撤销 Token',
    {
      confirmButtonText: '确定撤销',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    deleteToken(token.id)
  }).catch(() => {})
}

const deleteToken = async (id: number) => {
  try {
    await tokenApi.delete(id)
    ElMessage.success('Token 已撤销')
    loadTokens()
  } catch (error) {
    ElMessage.error('撤销 Token 失败')
  }
}

onMounted(() => {
  loadTokens()
})
</script>

<style scoped>
.tokens-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #303133;
}

.page-header .description {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.actions {
  margin-bottom: 16px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.token-display {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 16px;
}

.token-display code {
  flex: 1;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  word-break: break-all;
  color: #409eff;
}

.usage-example {
  background: #fafafa;
  padding: 12px;
  border-radius: 4px;
}

.usage-example h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #606266;
}

.usage-example p {
  margin: 0 0 8px 0;
  font-size: 13px;
  color: #909399;
}

.usage-example pre {
  margin: 0;
  padding: 8px;
  background: #282c34;
  border-radius: 4px;
  overflow-x: auto;
}

.usage-example code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #abb2bf;
}

.never-expires {
  color: #67c23a;
}

.never-used {
  color: #909399;
}
</style>
