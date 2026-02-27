import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页' },
  },
  {
    path: '/packages',
    name: 'Packages',
    component: () => import('@/views/Packages.vue'),
    meta: { title: '包列表' },
  },
  {
    path: '/guide',
    name: 'Guide',
    component: () => import('@/views/Guide.vue'),
    meta: { title: '使用指南' },
  },
  {
    path: '/package/:name',
    name: 'PackageDetail',
    component: () => import('@/views/PackageDetail.vue'),
    meta: { title: '包详情' },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/Admin.vue'),
    meta: { title: '管理后台', requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'tokens',
        name: 'AdminTokens',
        component: () => import('@/views/admin/Tokens.vue'),
        meta: { title: 'Token 管理' },
      },
      {
        path: 'backup',
        name: 'AdminBackup',
        component: () => import('@/views/admin/Backup.vue'),
        meta: { title: '备份恢复' },
      },
      {
        path: 'gc',
        name: 'AdminGC',
        component: () => import('@/views/admin/GC.vue'),
        meta: { title: '包清理' },
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/Settings.vue'),
        meta: { title: '系统设置' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 检查 JWT token 是否有效
function isTokenValid(token: string | null): boolean {
  if (!token) return false
  const parts = token.split('.')
  if (parts.length !== 3) return false
  const payloadStr = parts[1]
  if (!payloadStr) return false
  try {
    const payload = JSON.parse(atob(payloadStr))
    return payload.exp * 1000 > Date.now()
  } catch {
    return false
  }
}

// Navigation guard
router.beforeEach((to, _from, next) => {
  // Set page title
  document.title = `${to.meta.title || 'Grape'} - Grape`
  
  const token = localStorage.getItem('token')
  const validToken = isTokenValid(token)
  
  // 如果 token 无效，清除本地存储
  if (token && !validToken) {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }
  
  // 需要认证但 token 无效
  if (to.meta.requiresAuth && !validToken) {
    // 获取安全的重定向路径，防止开放重定向攻击
    const path = to.fullPath
    let safeRedirect = '/'
    if (path && path.startsWith('/') && !path.startsWith('//')) {
      safeRedirect = path
    }
    next({ name: 'Login', query: { redirect: safeRedirect } })
    return
  }
  
  // 需要 admin 角色检查
  if (to.meta.requiresAdmin) {
    const userStore = useUserStore()
    if (userStore.role !== 'admin') {
      next({ name: 'Home' })
      return
    }
  }
  
  next()
})

export default router