import axios from 'axios'
import router from '@/router'

// API 端口（npm Registry API）
const API_PORT = import.meta.env.VITE_API_PORT || '4874'

// 获取 API 基础 URL
const getApiBaseUrl = () => {
  const apiUrl = import.meta.env.VITE_API_URL
  if (apiUrl) return apiUrl
  
  // 开发环境：使用 API 端口
  const host = window.location.hostname || 'localhost'
  return `http://${host}:${API_PORT}`
}

const api = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: 30000,
})

// Web UI API（管理后台 API，走 4873 端口）
const webApi = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  timeout: 30000,
})

// Request interceptor - 注入 JWT token
// Request interceptor - 注入 JWT token
const requestInterceptor = (config: any) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}

const responseInterceptor = (error: any) => {
  if (error.response?.status === 401) {
    // 清除本地状态
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    // 跳转登录页
    const currentPath = window.location.pathname
    if (currentPath !== '/login') {
      router.push({ path: '/login', query: { redirect: currentPath } })
    }
  }
  return Promise.reject(error)
}

api.interceptors.request.use(requestInterceptor, (error) => Promise.reject(error))
api.interceptors.response.use((response) => response, responseInterceptor)

webApi.interceptors.request.use(requestInterceptor, (error) => Promise.reject(error))
webApi.interceptors.response.use((response) => response, responseInterceptor)

// Package APIs (npm Registry API - 端口 4874)
export const packageApi = {
  // Get package metadata
  getPackage(name: string) {
    return api.get(`/${encodeURIComponent(name)}`)
  },

  // Get all cached packages (Web API - 端口 4873)
  getPackages() {
    return webApi.get('/-/api/packages')
  },

  // Search packages (Web API - 端口 4873)
  search(query: string) {
    return webApi.get('/-/api/search', { params: { q: query } })
  },

  // Delete package
  deletePackage(name: string) {
    return api.delete(`/${encodeURIComponent(name)}`)
  },
}

// Auth APIs (Web API - 端口 4873)
export const authApi = {
  // Login (npm 兼容 API - 端口 4874)
  login(username: string, password: string) {
    return api.put('/-/user/org.couchdb.user:' + encodeURIComponent(username), {
      name: username,
      password: password,
    })
  },

  // Get current user
  getCurrentUser() {
    return webApi.get('/-/api/user')
  },

  // Logout
  logout() {
    return webApi.delete('/-/api/session')
  },
}

// Admin APIs (Web API - 端口 4873)
export const adminApi = {
  // Get all users
  getUsers() {
    return webApi.get('/-/api/admin/users')
  },

  // Create user
  createUser(user: { name: string; password: string; email: string; role?: string }) {
    return webApi.post('/-/api/admin/users', user)
  },

  // Update user
  updateUser(name: string, data: { email?: string; password?: string; role?: string }) {
    return webApi.put(`/-/api/admin/users/${encodeURIComponent(name)}`, data)
  },

  // Delete user
  deleteUser(name: string) {
    return webApi.delete(`/-/api/admin/users/${encodeURIComponent(name)}`)
  },

  // Get stats
  getStats() {
    return webApi.get('/-/api/stats')
  },

  // Get system info
  getSystemInfo() {
    return webApi.get('/-/api/admin/system')
  },

  // Get config
  getConfig() {
    return webApi.get('/-/api/admin/config')
  },

  // Update config
  updateConfig(data: object) {
    return webApi.put('/-/api/admin/config', data)
  },

  // Get webhooks
  getWebhooks() {
    return webApi.get('/-/api/admin/webhooks')
  },

  // Create webhook
  createWebhook(data: { name: string; url: string; secret?: string; events?: string; enabled?: boolean }) {
    return webApi.post('/-/api/admin/webhooks', data)
  },

  // Update webhook
  updateWebhook(id: number, data: { name?: string; url?: string; secret?: string; events?: string; enabled?: boolean }) {
    return webApi.put(`/-/api/admin/webhooks/${id}`, data)
  },

  // Delete webhook
  deleteWebhook(id: number) {
    return webApi.delete(`/-/api/admin/webhooks/${id}`)
  },

  // Test webhook
  testWebhook(id: number) {
    return webApi.post(`/-/api/admin/webhooks/${id}/test`)
  },

  // Get audit logs
  getAuditLogs(page = 1, limit = 20) {
    return webApi.get('/-/api/admin/audit-logs', { params: { page, limit } })
  },

  // Backup & Restore
  getBackupInfo() {
    return webApi.get('/-/api/admin/backup/info')
  },

  downloadBackup() {
    return webApi.get('/-/api/admin/backup/download', {
      responseType: 'blob',
    })
  },

  restoreBackup(formData: FormData) {
    return webApi.post('/-/api/admin/backup/restore', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  listBackups() {
    return webApi.get('/-/api/admin/backup/list')
  },

  // Garbage Collection
  getGCStats() {
    return webApi.get('/-/api/admin/gc/stats')
  },

  analyzeGC(params?: { days?: number; minVersions?: number; includeDeprecated?: boolean }) {
    return webApi.get('/-/api/admin/gc/analyze', { params })
  },

  runGC(data: {
    dryRun: boolean
    maxInactiveDays?: number
    minVersionsToKeep?: number
    includeDeprecated?: boolean
  }) {
    return webApi.post('/-/api/admin/gc/run', data)
  },

  deprecatePackage(name: string, data: { version?: string; reason: string }) {
    return webApi.post(`/-/api/admin/packages/${encodeURIComponent(name)}/deprecate`, data)
  },

  undeprecatePackage(name: string, version?: string) {
    const params = version ? { version } : {}
    return webApi.delete(`/-/api/admin/packages/${encodeURIComponent(name)}/deprecate`, { params })
  },
}

// Token APIs (CI/CD tokens - Web API 端口 4873)
export const tokenApi = {
  // List tokens
  list() {
    return webApi.get('/-/npm/v1/tokens')
  },

  // Create token
  create(data: { name: string; readonly?: boolean; days?: number }) {
    return webApi.post('/-/npm/v1/tokens', data)
  },

  // Delete token
  delete(id: number) {
    return webApi.delete(`/-/npm/v1/tokens/token/${id}`)
  },
}

// Package Owner APIs (Web API 端口 4873)
export const ownerApi = {
  // List package owners (admin)
  list(packageName: string) {
    return webApi.get(`/-/api/admin/packages/${encodeURIComponent(packageName)}/owners`)
  },

  // Add package owner (admin)
  add(packageName: string, username: string) {
    return webApi.post(`/-/api/admin/packages/${encodeURIComponent(packageName)}/owners`, { name: username })
  },

  // Remove package owner (admin)
  remove(packageName: string, username: string) {
    return webApi.delete(`/-/api/admin/packages/${encodeURIComponent(packageName)}/owners/${encodeURIComponent(username)}`)
  },
}

export default api