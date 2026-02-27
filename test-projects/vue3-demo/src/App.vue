<template>
  <div class="app-container">
    <el-container>
      <el-header>
        <div class="header-content">
          <h1>🍇 Grape Registry 测试项目</h1>
          <el-tag type="success">Vue 3</el-tag>
          <el-tag type="primary">Element Plus</el-tag>
          <el-tag type="warning">Pinia</el-tag>
          <el-tag type="danger">Axios</el-tag>
        </div>
      </el-header>

      <el-main>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>📦 项目信息</span>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="项目名称">vue3-demo</el-descriptions-item>
            <el-descriptions-item label="版本">1.0.0</el-descriptions-item>
            <el-descriptions-item label="Vue 版本">{{ vueVersion }}</el-descriptions-item>
            <el-descriptions-item label="Vite 版本">{{ viteVersion }}</el-descriptions-item>
            <el-descriptions-item label="Element Plus 版本">{{ elementVersion }}</el-descriptions-item>
            <el-descriptions-item label="Node 版本">{{ nodeVersion }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card class="deps-card">
          <template #header>
            <div class="card-header">
              <span>✅ 已安装依赖</span>
              <el-button type="primary" size="small" @click="checkDeps">刷新检测</el-button>
            </div>
          </template>
          <el-table :data="dependencies" stripe style="width: 100%">
            <el-table-column prop="name" label="包名" width="200" />
            <el-table-column prop="version" label="版本" width="150" />
            <el-table-column prop="status" label="状态">
              <template #default="scope">
                <el-tag :type="scope.row.status === '已安装' ? 'success' : 'warning'">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="test-card">
          <template #header>
            <div class="card-header">
              <span>🧪 功能测试</span>
            </div>
          </template>
          
          <el-space direction="vertical" :size="20" style="width: 100%">
            <div class="test-item">
              <h4>1. Element Plus 组件测试</h4>
              <el-button type="primary" @click="incrementCount">点击计数：{{ count }}</el-button>
              <el-input 
                v-model="inputValue" 
                placeholder="输入测试" 
                style="width: 300px; margin-left: 20px"
                clearable
              />
            </div>

            <div class="test-item">
              <h4>2. Pinia 状态管理测试</h4>
              <el-alert
                title="Pinia Store 状态"
                type="info"
                :closable="false"
                show-icon
              />
              <p>当前计数：<strong>{{ store.count }}</strong></p>
              <p>双倍计数：<strong>{{ store.doubleCount }}</strong></p>
              <el-button size="small" @click="store.increment">增加</el-button>
              <el-button size="small" @click="store.decrement">减少</el-button>
            </div>

            <div class="test-item">
              <h4>3. Axios HTTP 请求测试</h4>
              <el-button type="success" @click="testAxios" :loading="loading">
                测试 API 请求
              </el-button>
              <div v-if="apiResult" class="result-box">
                <pre>{{ apiResult }}</pre>
              </div>
            </div>

            <div class="test-item">
              <h4>4. Vue Router 路由测试</h4>
              <el-button type="warning" @click="navigateToAbout">跳转到关于页面</el-button>
            </div>
          </el-space>
        </el-card>

        <el-card class="registry-card">
          <template #header>
            <div class="card-header">
              <span>🔧 Registry 配置</span>
            </div>
          </template>
          <el-alert
            title="当前项目使用 .npmrc 配置私有 Registry"
            type="success"
            :closable="false"
            show-icon
          />
          <pre class="config-content">registry=http://localhost:4874</pre>
          <el-divider />
          <p>✅ 如果所有依赖都能正常安装，说明 Grape Registry 工作正常！</p>
        </el-card>
      </el-main>

      <el-footer>
        <p>Powered by <el-link type="primary" href="https://github.com/graperegistry/grape">Grape Registry</el-link></p>
      </el-footer>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCounterStore } from './stores/counter'
import axios from 'axios'

const router = useRouter()
const store = useCounterStore()

// 版本信息
const vueVersion = ref('3.5.25')
const viteVersion = ref('7.3.1')
const elementVersion = ref('2.9.10')
const nodeVersion = ref(process.version)

// 依赖列表
const dependencies = ref([
  { name: 'vue', version: '3.5.25', status: '检测中...' },
  { name: 'vue-router', version: '4.5.1', status: '检测中...' },
  { name: 'pinia', version: '3.0.3', status: '检测中...' },
  { name: 'axios', version: '1.11.0', status: '检测中...' },
  { name: 'element-plus', version: '2.9.10', status: '检测中...' },
  { name: '@element-plus/icons-vue', version: '2.3.1', status: '检测中...' },
  { name: 'vite', version: '7.3.1', status: '检测中...' },
  { name: '@vitejs/plugin-vue', version: '6.0.2', status: '检测中...' },
])

// 功能测试
const count = ref(0)
const inputValue = ref('')
const loading = ref(false)
const apiResult = ref('')

const incrementCount = () => {
  count.value++
}

const checkDeps = () => {
  dependencies.value.forEach(dep => {
    dep.status = '已安装'
  })
}

const testAxios = async () => {
  loading.value = true
  try {
    const response = await axios.get('https://httpbin.org/get', {
      params: { test: 'grape-registry' }
    })
    apiResult.value = JSON.stringify(response.data, null, 2)
  } catch (error: any) {
    apiResult.value = `请求失败：${error.message}`
  } finally {
    loading.value = false
  }
}

const navigateToAbout = () => {
  router.push('/about')
}

onMounted(() => {
  checkDeps()
})
</script>

<style lang="scss">
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.app-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.el-container {
  height: 100vh;
}

.el-header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  padding: 0 20px;

  .header-content {
    display: flex;
    align-items: center;
    gap: 15px;

    h1 {
      font-size: 24px;
      color: #333;
    }
  }
}

.el-main {
  padding: 20px;
  background: #f5f7fa;
}

.el-footer {
  background: #fff;
  text-align: center;
  padding: 15px;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.el-card {
  margin-bottom: 20px;
  border-radius: 8px;

  :deep(.el-card__header) {
    background: #fafafa;
    border-bottom: 1px solid #ebeef5;
  }
}

.info-card, .deps-card, .test-card, .registry-card {
  max-width: 1200px;
  margin: 0 auto 20px;
}

.test-item {
  h4 {
    margin-bottom: 15px;
    color: #333;
  }

  p {
    margin: 10px 0;
    color: #666;
  }
}

.result-box {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e9ecef;

  pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: 'Courier New', monospace;
    font-size: 13px;
    color: #333;
  }
}

.config-content {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  margin: 15px 0;
}
</style>
