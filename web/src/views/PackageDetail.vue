<template>
  <div class="package-detail-page fade-in" v-loading="loading">
    <template v-if="packageData">
      <!-- Header -->
      <div class="package-header-section">
        <div class="package-title">
          <h1>{{ packageData.name }}</h1>
          <el-tag v-if="packageData.private" type="success">私有包</el-tag>
          <el-tag v-else type="info">缓存包</el-tag>
        </div>
        <p class="package-description">{{ packageData.description || '暂无描述' }}</p>
      </div>

      <!-- Version Selector -->
      <el-card class="version-section">
        <template #header>
          <div class="section-header">
            <span>版本选择</span>
            <el-select v-model="selectedVersion" placeholder="选择版本" style="width: 200px">
              <el-option
                v-for="(_v, version) in packageData.versions"
                :key="version"
                :label="String(version)"
                :value="String(version)"
              />
            </el-select>
          </div>
        </template>

        <div class="install-command">
          <h4>安装命令</h4>
          <div class="code-block">
            <code>npm install {{ packageData.name }}@{{ selectedVersion || 'latest' }}</code>
            <el-button text @click="copyInstallCommand">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
        </div>
      </el-card>

      <!-- Version Info -->
      <el-card v-if="currentVersion" class="info-section">
        <template #header>
          <span>版本信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="版本号">{{ selectedVersion }}</el-descriptions-item>
          <el-descriptions-item label="许可证">{{ currentVersion.license || 'MIT' }}</el-descriptions-item>
          <el-descriptions-item label="主页">{{ currentVersion.homepage || '-' }}</el-descriptions-item>
          <el-descriptions-item label="仓库">
            <a v-if="currentVersion.repository?.url" :href="currentVersion.repository.url" target="_blank">
              {{ currentVersion.repository.url }}
            </a>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="发布者">{{ packageData._npmUser?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布时间">{{ formatTime(packageData.time?.[selectedVersion || '']) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Dependencies -->
      <el-card v-if="currentVersion?.dependencies" class="deps-section">
        <template #header>
          <span>依赖项</span>
        </template>
        <div class="deps-list">
          <el-tag
            v-for="(_version, dep) in currentVersion.dependencies"
            :key="String(dep)"
            class="dep-tag"
            @click="goToPackage(String(dep))"
          >
            {{ dep }}
          </el-tag>
        </div>
      </el-card>

      <!-- Readme -->
      <el-card class="readme-section">
        <template #header>
          <span>README</span>
        </template>
        <div class="readme-content" v-html="readmeHtml"></div>
      </el-card>

      <!-- Admin Actions -->
      <el-card v-if="userStore.isLoggedIn" class="actions-section">
        <template #header>
          <span>管理操作</span>
        </template>
        
        <!-- Owner 管理（仅 admin） -->
        <div v-if="userStore.role === 'admin'" class="owner-management">
          <h4>包所有者</h4>
          <div class="owners-list" v-loading="ownersLoading">
            <el-tag 
              v-for="owner in owners" 
              :key="owner.username"
              closable
              @close="removeOwner(owner.username)"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ owner.username }}
            </el-tag>
            <el-tag v-if="owners.length === 0" type="info">无限制（任何 developer 可发布）</el-tag>
          </div>
          <div class="add-owner" style="margin-top: 12px;">
            <el-input 
              v-model="newOwnerUsername" 
              placeholder="输入用户名" 
              style="width: 200px; margin-right: 8px;"
              @keyup.enter="addOwner"
            />
            <el-button type="primary" @click="addOwner" :loading="addingOwner">添加所有者</el-button>
          </div>
        </div>
        
        <el-divider v-if="userStore.role === 'admin'" />
        
        <el-button type="danger" @click="handleDelete">
          删除此包
        </el-button>
      </el-card>
    </template>

    <el-empty v-else-if="!loading" description="包不存在" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { packageApi, ownerApi } from '@/api'
import { useUserStore } from '@/stores/user'

// 配置 marked：开启 gfm（GitHub Flavored Markdown）和换行支持
marked.setOptions({ gfm: true, breaks: true })

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const packageData = ref<any>(null)
const selectedVersion = ref('')
const readmeHtml = ref('')

// Owner 管理
const owners = ref<{ username: string; email: string }[]>([])
const ownersLoading = ref(false)
const newOwnerUsername = ref('')
const addingOwner = ref(false)

const currentVersion = computed(() => {
  if (!packageData.value?.versions || !selectedVersion.value) return null
  return packageData.value.versions[selectedVersion.value]
})

const loadPackage = async () => {
  const name = route.params.name as string
  if (!name) return

  loading.value = true
  try {
    const res = await packageApi.getPackage(decodeURIComponent(name))
    packageData.value = res.data
    
    const latestVersion = packageData.value['dist-tags']?.latest
    if (latestVersion) {
      selectedVersion.value = String(latestVersion)
    } else if (packageData.value.versions) {
      const versions = Object.keys(packageData.value.versions)
      if (versions.length > 0) {
        selectedVersion.value = String(versions[versions.length - 1] || '')
      }
    }

    const rawReadme = packageData.value.readme || ''
    readmeHtml.value = rawReadme
      ? (marked.parse(rawReadme) as string)
      : '<p style="color:#999">暂无 README</p>'
    
    // 加载 owner 列表
    loadOwners()
  } catch (error) {
    // 加载失败，错误已通过 ElMessage 通知用户
    ElMessage.error('加载包失败')
  } finally {
    loading.value = false
  }
}

const copyInstallCommand = () => {
  const cmd = `npm install ${packageData.value.name}@${selectedVersion.value || 'latest'}`
  navigator.clipboard.writeText(cmd)
  ElMessage.success('已复制到剪贴板')
}

const goToPackage = (name: string) => {
  router.push(`/package/${encodeURIComponent(name)}`)
}

const formatTime = (time: string | undefined): string => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定要删除此包吗？此操作不可恢复。', '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await packageApi.deletePackage(packageData.value.name)
    ElMessage.success('删除成功')
    router.push('/packages')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// Owner 管理
const loadOwners = async () => {
  if (userStore.role !== 'admin' || !packageData.value) return
  
  ownersLoading.value = true
  try {
    const res = await ownerApi.list(packageData.value.name)
    owners.value = res.data.owners || []
  } catch {
    // 忽略错误
  } finally {
    ownersLoading.value = false
  }
}

const addOwner = async () => {
  if (!newOwnerUsername.value.trim()) return
  
  addingOwner.value = true
  try {
    await ownerApi.add(packageData.value.name, newOwnerUsername.value.trim())
    ElMessage.success('已添加所有者')
    newOwnerUsername.value = ''
    loadOwners()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '添加失败')
  } finally {
    addingOwner.value = false
  }
}

const removeOwner = async (username: string) => {
  try {
    await ElMessageBox.confirm(`确定要移除 ${username} 的所有者权限吗？`, '确认', {
      confirmButtonText: '移除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await ownerApi.remove(packageData.value.name, username)
    ElMessage.success('已移除所有者')
    loadOwners()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '移除失败')
    }
  }
}

onMounted(() => {
  loadPackage()
})

watch(() => route.params.name, () => {
  loadPackage()
})
</script>

<style scoped>
.package-detail-page {
  max-width: 900px;
  margin: 0 auto;
}

.package-header-section {
  margin-bottom: 24px;
}

.package-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.package-title h1 {
  font-size: 28px;
  color: var(--grape-primary);
  margin: 0;
}

.owner-management {
  margin-bottom: 16px;
}

.owner-management h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #606266;
}

.owners-list {
  min-height: 32px;
}

.package-description {
  color: #666;
  margin-top: 8px;
  font-size: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-section,
.info-section,
.deps-section,
.readme-section,
.actions-section {
  margin-bottom: 20px;
}

.install-command h4 {
  margin-bottom: 8px;
}

.install-command .code-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.deps-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.dep-tag {
  cursor: pointer;
}

.dep-tag:hover {
  background-color: var(--grape-primary);
  color: white;
}

.readme-content {
  max-height: 600px;
  overflow-y: auto;
  line-height: 1.7;
}

.readme-content :deep(pre) {
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 12px 0;
}

.readme-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
  font-size: 13px;
}

.readme-content :deep(code) {
  background-color: #f0f0f0;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Fira Code', 'Courier New', monospace;
  font-size: 13px;
  color: #c7254e;
}

.readme-content :deep(h1) {
  font-size: 22px;
  margin-top: 24px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
}

.readme-content :deep(h2) {
  font-size: 18px;
  margin-top: 20px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid #eee;
}

.readme-content :deep(h3) {
  font-size: 16px;
  margin-top: 16px;
  margin-bottom: 8px;
}

.readme-content :deep(p) {
  margin-bottom: 12px;
}

.readme-content :deep(ul),
.readme-content :deep(ol) {
  padding-left: 24px;
  margin-bottom: 12px;
}

.readme-content :deep(li) {
  margin-bottom: 4px;
}

.readme-content :deep(blockquote) {
  border-left: 4px solid #ddd;
  padding-left: 16px;
  margin: 12px 0;
  color: #666;
}

.readme-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin-bottom: 16px;
}

.readme-content :deep(th),
.readme-content :deep(td) {
  border: 1px solid #ddd;
  padding: 8px 12px;
  text-align: left;
}

.readme-content :deep(th) {
  background-color: #f5f5f5;
  font-weight: 600;
}

.readme-content :deep(a) {
  color: var(--grape-primary, #409eff);
  text-decoration: none;
}

.readme-content :deep(a:hover) {
  text-decoration: underline;
}

.readme-content :deep(img) {
  max-width: 100%;
  border-radius: 4px;
}
</style>