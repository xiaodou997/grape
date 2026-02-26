import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:4873',
  timeout: 30000,
})

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

// Package APIs
export const packageApi = {
  // Get package metadata
  getPackage(name: string) {
    return api.get(`/${name}`)
  },

  // Get all cached packages
  getPackages() {
    return api.get('/-/api/packages')
  },

  // Search packages
  search(query: string) {
    return api.get('/-/api/search', { params: { q: query } })
  },

  // Delete package
  deletePackage(name: string) {
    return api.delete(`/${name}`)
  },
}

// Auth APIs
export const authApi = {
  // Login
  login(username: string, password: string) {
    return api.put('/-/user/org.couchdb.user:' + username, {
      name: username,
      password: password,
    })
  },

  // Get current user
  getCurrentUser() {
    return api.get('/-/api/user')
  },

  // Logout
  logout() {
    return api.delete('/-/api/session')
  },
}

// Admin APIs
export const adminApi = {
  // Get all users
  getUsers() {
    return api.get('/-/api/admin/users')
  },

  // Create user
  createUser(user: { name: string; password: string; email: string }) {
    return api.post('/-/api/admin/users', user)
  },

  // Delete user
  deleteUser(name: string) {
    return api.delete(`/-/api/admin/users/${name}`)
  },

  // Get audit logs
  getAuditLogs(page = 1, limit = 20) {
    return api.get('/-/api/admin/audit-logs', { params: { page, limit } })
  },
}

export default api
