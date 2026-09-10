import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface SSHConfig {
  port: number
  permit_root_login: boolean
  permit_password_login: boolean
  max_auth_tries: number
  client_alive_interval: number
  allow_users: string
}

export interface SSHKey {
  id: number
  type: string
  comment: string
  fingerprint: string
  created_at: string
}

export interface SSHLoginLog {
  time: string
  user: string
  ip: string
  status: 'success' | 'failed'
  method: string
}

export interface SSHLogListData {
  items: SSHLoginLog[]
  total: number
}

export interface Fail2banConfig {
  enabled: boolean
  max_retry: number
  ban_time: number
  find_time: number
}

export interface BannedIP {
  ip: string
  jail: string
  banned_at: string
  reason: string
}

export interface Fail2banStatus {
  enabled: boolean
  installed: boolean
  banned_ips: BannedIP[]
}

export interface AddSSHKeyData {
  /** 后端 dto 绑定的字段名是 public_key（required） */
  public_key: string
  comment?: string
}

/* ==================== SSH 配置 ==================== */

export function getSSHConfig() {
  return request<SSHConfig>({
    url: '/ssh/config',
    method: 'get'
  })
}

export function updateSSHConfig(data: SSHConfig) {
  return request<SSHConfig>({
    url: '/ssh/config',
    method: 'put',
    data
  })
}

/* ==================== 密钥管理 ==================== */

export function getSSHKeys() {
  return request<SSHKey[]>({
    url: '/ssh/keys',
    method: 'get'
  })
}

export function addSSHKey(data: AddSSHKeyData) {
  return request<SSHKey>({
    url: '/ssh/keys',
    method: 'post',
    data
  })
}

export function deleteSSHKey(fingerprint: string) {
  return request({
    url: `/ssh/keys/${fingerprint}`,
    method: 'delete'
  })
}

/* ==================== 登录日志 ==================== */

export function getSSHLogs(page = 1, pageSize = 20) {
  return request<SSHLogListData>({
    url: '/ssh/logs',
    method: 'get',
    params: { page, pageSize }
  })
}

/* ==================== Fail2ban ==================== */

export function getFail2banStatus() {
  return request<Fail2banStatus>({
    url: '/ssh/fail2ban/status',
    method: 'get'
  })
}

export function updateFail2banConfig(data: Fail2banConfig) {
  return request<Fail2banConfig>({
    url: '/ssh/fail2ban/config',
    method: 'put',
    data
  })
}

export function unbanIP(ip: string) {
  return request({
    url: '/ssh/fail2ban/unban',
    method: 'post',
    data: { ip }
  })
}

export function banIP(ip: string, reason?: string) {
  return request({
    url: '/ssh/fail2ban/ban',
    method: 'post',
    data: { ip, reason }
  })
}
