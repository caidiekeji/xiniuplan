import { request } from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginData {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
  require_password_change?: boolean
  user?: {
    id: number
    username: string
    email: string
    role: string
  }
}

export interface ChangePasswordParams {
  old_password: string
  new_password: string
}

export function login(data: LoginParams) {
  return request<LoginData>({
    url: '/auth/login',
    method: 'post',
    data
  })
}

export function changePassword(data: ChangePasswordParams) {
  return request<{ updated: boolean }>({
    url: '/auth/change-password',
    method: 'put',
    data
  })
}

export function forgotPassword(data: { username: string; email: string }) {
  return request<{ sent: boolean }>({
    url: '/auth/forgot-password',
    method: 'post',
    data
  })
}

export function resetPassword(data: { token: string; new_password: string }) {
  return request<{ reset: boolean }>({
    url: '/auth/reset-password',
    method: 'post',
    data
  })
}

export function logout() {
  return request({
    url: '/auth/logout',
    method: 'post'
  })
}

export function refresh() {
  return request<{ access_token: string }>({
    url: '/auth/refresh',
    method: 'post'
  })
}

export function getUserInfo() {
  return request({
    url: '/auth/info',
    method: 'get'
  })
}
