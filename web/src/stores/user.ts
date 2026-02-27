import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref<string | null>(localStorage.getItem('username'))
  const role = ref<string | null>(localStorage.getItem('role'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'admin')

  async function login(name: string, password: string) {
    try {
      const res = await authApi.login(name, password)
      if (res.data.ok) {
        token.value = res.data.token
        username.value = name
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('username', name)
        
        // 获取用户角色
        try {
          const userRes = await authApi.getCurrentUser()
          if (userRes.data.role) {
            role.value = userRes.data.role
            localStorage.setItem('role', userRes.data.role)
          }
        } catch {
          // 忽略获取用户信息失败
        }
        
        return true
      }
      return false
    } catch {
      return false
    }
  }

  function logout() {
    token.value = null
    username.value = null
    role.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }

  // 初始化时从 localStorage 恢复角色
  function init() {
    const storedRole = localStorage.getItem('role')
    if (storedRole) {
      role.value = storedRole
    }
  }

  init()

  return {
    token,
    username,
    role,
    isLoggedIn,
    isAdmin,
    login,
    logout,
  }
})
