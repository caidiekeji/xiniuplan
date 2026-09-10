import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import i18n from '@/locales'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Login.vue'),
    meta: { public: true, title: 'auth.login' }
  },
  {
    path: '/change-password',
    name: 'ChangePassword',
    component: () => import('@/views/login/ChangePassword.vue'),
    meta: { title: '修改密码' }
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Dashboard.vue'),
        meta: { title: 'menu.dashboard', icon: 'dashboard' }
      },
      {
        path: 'website',
        name: 'Website',
        component: () => import('@/views/website/index.vue'),
        meta: { title: 'menu.website', icon: 'website' }
      },
      {
        path: 'database',
        name: 'Database',
        component: () => import('@/views/database/index.vue'),
        meta: { title: 'menu.database', icon: 'database' }
      },
      {
        path: 'container',
        name: 'Container',
        component: () => import('@/views/container/index.vue'),
        meta: { title: 'menu.container', icon: 'container' }
      },
      {
        path: 'appstore',
        name: 'AppStore',
        component: () => import('@/views/appstore/index.vue'),
        meta: { title: 'menu.appstore', icon: 'appstore' }
      },
      {
        path: 'plugins',
        name: 'Plugins',
        component: () => import('@/views/plugins/index.vue'),
        meta: { title: 'menu.plugins', icon: 'plugins' }
      },
      {
        path: 'file',
        name: 'File',
        component: () => import('@/views/file/index.vue'),
        meta: { title: 'menu.file', icon: 'file' }
      },
      {
        path: 'terminal',
        name: 'Terminal',
        component: () => import('@/views/terminal/index.vue'),
        meta: { title: 'menu.terminal', icon: 'terminal' }
      },
      {
        path: 'cron',
        name: 'Cron',
        component: () => import('@/views/cron/index.vue'),
        meta: { title: 'menu.cron', icon: 'cron' }
      },
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => import('@/views/monitor/index.vue'),
        meta: { title: 'menu.monitor', icon: 'monitor' }
      },
      {
        path: 'backup',
        name: 'Backup',
        component: () => import('@/views/backup/index.vue'),
        meta: { title: 'menu.backup', icon: 'backup' }
      },
      {
        path: 'firewall',
        name: 'Firewall',
        component: () => import('@/views/firewall/index.vue'),
        meta: { title: 'menu.firewall', icon: 'firewall' }
      },
      {
        path: 'ssh',
        name: 'SSH',
        component: () => import('@/views/ssh/index.vue'),
        meta: { title: 'menu.ssh', icon: 'ssh' }
      },
      {
        path: 'alert',
        name: 'Alert',
        component: () => import('@/views/alert/index.vue'),
        meta: { title: 'menu.alert', icon: 'alert' }
      },
      {
        path: 'ssl',
        name: 'SSL',
        component: () => import('@/views/ssl/index.vue'),
        meta: { title: 'menu.ssl', icon: 'ssl' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/index.vue'),
        meta: { title: 'menu.users', icon: 'users' }
      },
      {
        path: 'audit',
        name: 'Audit',
        component: () => import('@/views/audit/index.vue'),
        meta: { title: 'menu.audit', icon: 'audit' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/index.vue'),
        meta: { title: 'menu.settings', icon: 'settings' }
      }
    ]
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('@/views/login/ResetPassword.vue'),
    meta: { public: true, title: '重置密码' }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  // 支持安全入口：main.ts 引导时确定前缀（如 /admin-xyz），本模块在 base 设置后才被动态导入
  history: createWebHistory((window as any).__PANEL_BASE__ || '/'),
  routes
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (!to.meta.public && !authStore.token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/login' && authStore.token) {
    // 强制改密期间访问登录页 → 引导到改密页
    if (authStore.mustChangePassword) {
      return { path: '/change-password' }
    }
    return { path: '/dashboard' }
  }
  // 强制改密：除登录页/改密页外一律拦截
  if (authStore.token && authStore.mustChangePassword && to.path !== '/change-password') {
    return { path: '/change-password' }
  }
  // 未处于强制改密状态却访问改密页 → 回首页
  if (to.path === '/change-password' && !authStore.mustChangePassword) {
    return { path: '/dashboard' }
  }
  const titleKey = to.meta.title as string | undefined
  const pageTitle = titleKey ? i18n.global.t(titleKey) : ''
  const brandName = i18n.global.t('auth.brandName')
  document.title = pageTitle ? `${pageTitle} · ${brandName}` : brandName
})

export default router
