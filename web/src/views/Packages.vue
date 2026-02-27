<template>
  <div class="packages-page fade-in">
    <!-- Search Bar -->
    <div class="search-section">
      <el-input
        v-model="searchQuery"
        placeholder="搜索包..."
        size="large"
        clearable
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>

    <!-- Package List -->
    <div class="packages-list" v-loading="loading">
      <el-row :gutter="20">
        <el-col :span="8" v-for="pkg in paginatedPackages" :key="pkg.name">
          <el-card class="package-card" shadow="hover" @click="goToPackage(pkg.name)">
            <div class="package-header">
              <h3 class="package-name">{{ pkg.name }}</h3>
              <el-tag v-if="pkg.private" type="success" size="small">私有</el-tag>
              <el-tag v-else type="info" size="small">缓存</el-tag>
            </div>
            <p class="package-description">{{ pkg.description || '暂无描述' }}</p>
            <div class="package-meta">
              <span class="version">
                <el-icon><Document /></el-icon>
                {{ pkg.version || 'latest' }}
              </span>
              <span class="time">
                <el-icon><Clock /></el-icon>
                {{ formatTime(pkg.updatedAt) }}
              </span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="filteredPackages.length === 0 && !loading" description="暂无包" />

      <!-- Pagination -->
      <div class="pagination-section" v-if="totalPackages > pageSize">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="totalPackages"
          layout="total, prev, pager, next, jumper"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Document, Clock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { packageApi } from '@/api'
import type { Package } from '@/api/types'

const router = useRouter()
const loading = ref(false)
const searchQuery = ref('')
const packages = ref<Package[]>([])

// 分页
const currentPage = ref(1)
const pageSize = ref(30)  // 每页显示 30 个包

const filteredPackages = computed(() => {
  if (!searchQuery.value) return packages.value
  const query = searchQuery.value.toLowerCase()
  return packages.value.filter(pkg => 
    pkg.name.toLowerCase().includes(query) ||
    (pkg.description && pkg.description.toLowerCase().includes(query))
  )
})

const totalPackages = computed(() => filteredPackages.value.length)

const paginatedPackages = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredPackages.value.slice(start, end)
})

const handleSearch = () => {
  // 搜索时重置到第一页
  currentPage.value = 1
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  // 滚动到顶部
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const goToPackage = (name: string) => {
  router.push(`/package/${encodeURIComponent(name)}`)
}

const formatTime = (time: string | undefined) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleDateString('zh-CN')
}

const loadPackages = async () => {
  loading.value = true
  try {
    const res = await packageApi.getPackages()
    packages.value = res.data.packages || []
  } catch {
    ElMessage.error('加载包列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPackages()
})
</script>

<style scoped>
.packages-page {
  max-width: 1200px;
  margin: 0 auto;
}

.search-section {
  margin-bottom: 24px;
}

.packages-list {
  min-height: 400px;
}

.package-card {
  cursor: pointer;
  margin-bottom: 20px;
}

.package-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.package-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--grape-primary);
  margin: 0;
  word-break: break-all;
}

.package-description {
  color: #666;
  font-size: 14px;
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.package-meta {
  display: flex;
  gap: 16px;
  color: #999;
  font-size: 13px;
}

.package-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pagination-section {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}
</style>
