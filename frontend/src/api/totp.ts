import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface TOTPStatus {
  enabled: boolean
  bound_at?: string
}

export interface TOTPBindData {
  secret: string
  otpauth_url: string
}

export interface TOTPVerifyData {
  recovery_codes: string[]
}

export interface TOTPVerifyParams {
  /** 绑定时服务端返回的 secret（后端 Verify 用它校验验证码） */
  secret: string
  code: string
}

export interface TOTPDisableParams {
  code: string
}

export interface TOTPLoginParams {
  temp_token: string
  /** 后端 dto 绑定的字段名是 totp_code（required） */
  totp_code: string
}

export interface TOTPLoginData {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
  user?: {
    id: number
    username: string
    email: string
    role: string
  }
}

/* ==================== TOTP 接口 ==================== */

export function getTOTPStatus() {
  return request<TOTPStatus>({
    url: '/auth/totp/status',
    method: 'get'
  })
}

export function bindTOTP() {
  return request<TOTPBindData>({
    url: '/auth/totp/bind',
    method: 'post'
  })
}

export function verifyTOTP(data: TOTPVerifyParams) {
  return request<TOTPVerifyData>({
    url: '/auth/totp/verify',
    method: 'post',
    data
  })
}

export function disableTOTP(data: TOTPDisableParams) {
  return request({
    url: '/auth/totp/disable',
    method: 'post',
    data
  })
}

export function loginTOTP(data: TOTPLoginParams) {
  return request<TOTPLoginData>({
    url: '/auth/totp/login',
    method: 'post',
    data,
    // temp_token 通过 Bearer 头传递（后端 JWTAuth 校验 type=totp_temp）
    headers: { Authorization: `Bearer ${data.temp_token}` }
  })
}
