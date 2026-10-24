import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import i18n from '@/locales'

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

const service: AxiosInstance = axios.create({
  // 安全入口模式下 __PANEL_BASE__ 为 /{entry}，否则为空字符串
  baseURL: `${(window as any).__PANEL_BASE__ || ''}/api/v1`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器：附加 token
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    // 显式传入的 Authorization（如 TOTP 二次验证的 temp_token）优先
    if (config.headers.Authorization) {
      return config
    }
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：统一处理
service.interceptors.response.use(
  (response: AxiosResponse) => {
    // blob 下载直接返回原始响应
    if (response.config.responseType === 'blob') {
      return response
    }
    const res = response.data as ApiResponse
    if (res.code !== 200) {
      if (res.code === 401) {
        const authStore = useAuthStore()
        authStore.clearToken()
        router.push('/login')
      }
      // 强制改密：拦截业务请求，引导到改密页
      if (res.code === 40302) {
        const authStore = useAuthStore()
        authStore.setMustChangePassword(true)
        if (router.currentRoute.value.path !== '/change-password') {
          router.push('/change-password')
        }
      }
      return Promise.reject(new Error(res.message || i18n.global.t('common.errors.requestFailed')))
    }
    return res as unknown as AxiosResponse
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.clearToken()
      router.push('/login')
    }
    // 强制改密（HTTP 403 + 业务码 40302）
    if (error.response?.status === 403 && error.response?.data?.code === 40302) {
      const authStore = useAuthStore()
      authStore.setMustChangePassword(true)
      if (router.currentRoute.value.path !== '/change-password') {
        router.push('/change-password')
      }
    }
    const message = error.response?.data?.message || error.message || i18n.global.t('common.errors.networkError')
    return Promise.reject(new Error(message))
  }
)

export function request<T = any>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.request(config)
}

export default service
