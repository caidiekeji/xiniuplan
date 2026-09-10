import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, logout as apiLogout, refresh as apiRefresh, getUserInfo as apiGetUserInfo, changePassword as apiChangePassword, type ChangePasswordParams } from '@/api/auth'

export interface UserInfo {
  id: number
  username: string
  email: string
  role: string
  avatar?: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('cloudpivot_token') || '')
  const userInfo = ref<UserInfo | null>(null)
  const mustChangePassword = ref<boolean>(localStorage.getItem('cloudpivot_must_change_pwd') === '1')

  const isLoggedIn = computed(() => !!token.value)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('cloudpivot_token', t)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('cloudpivot_token')
  }

  function setMustChangePassword(v: boolean) {
    mustChangePassword.value = v
    if (v) {
      localStorage.setItem('cloudpivot_must_change_pwd', '1')
    } else {
      localStorage.removeItem('cloudpivot_must_change_pwd')
    }
  }

  async function login(username: string, password: string) {
    const res = await apiLogin({ username, password })
    setToken(res.data.access_token)
    // 登录响应携带强制改密标志
    setMustChangePassword(!!res.data.require_password_change)
    if (res.data.user) {
      userInfo.value = res.data.user
    }
    return res
  }

  async function changePassword(params: ChangePasswordParams) {
    const res = await apiChangePassword(params)
    // 改密成功：后端已将旧 token 加入黑名单，清除本地状态回到登录页
    setMustChangePassword(false)
    clearToken()
    userInfo.value = null
    return res
  }

  async function logout() {
    try {
      await apiLogout()
    } finally {
      clearToken()
      userInfo.value = null
      setMustChangePassword(false)
    }
  }

  async function refresh() {
    const res = await apiRefresh()
    setToken(res.data.access_token)
    return res
  }

  async function fetchUserInfo() {
    const res = await apiGetUserInfo()
    // 以服务端状态为准（处理刷新页面后的标志同步）
    if (typeof (res.data as any)?.must_change_password === 'boolean') {
      setMustChangePassword((res.data as any).must_change_password)
    }
    userInfo.value = res.data
    return res
  }

  return {
    token,
    userInfo,
    mustChangePassword,
    isLoggedIn,
    login,
    logout,
    changePassword,
    refresh,
    fetchUserInfo,
    setToken,
    clearToken,
    setMustChangePassword
  }
})
