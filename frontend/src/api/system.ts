import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface SystemSettings {
  panel_port: number
  domain: string
  language: 'zh-CN' | 'en-US'
  theme: 'dark' | 'light' | 'auto'
  security_entry: string
  session_timeout: number
}

export interface SmtpConfig {
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_from: string
}

export interface VersionInfo {
  panel_version: string
  go_version: string
  build_time: string
  uptime: number
  latest_version?: string
  has_update?: boolean
}

export interface SystemInfo {
  hostname: string
  os: string
  kernel: string
  arch: string
  cpu_model: string
  cpu_cores: number
  memory_total: number
  swap_total: number
  uptime: number
  load_avg: number[]
}

export type CleanType = 'cache' | 'docker' | 'logs'

export interface CleanResult {
  type: CleanType
  freed_bytes: number
  cleaned_items: number
}

/* ==================== 面板设置 ==================== */

export function getSettings() {
  return request<SystemSettings>({
    url: '/system/settings',
    method: 'get'
  })
}

export function updateSettings(settings: Partial<SystemSettings>) {
  return request<SystemSettings>({
    url: '/system/settings',
    method: 'put',
    data: settings
  })
}

/* ==================== 邮件配置 ==================== */

export function getSmtpConfig() {
  return request<SmtpConfig>({
    url: '/system/smtp',
    method: 'get'
  })
}

export function updateSmtpConfig(config: SmtpConfig) {
  return request<SmtpConfig>({
    url: '/system/smtp',
    method: 'put',
    data: config
  })
}

export function sendTestEmail(to: string) {
  return request({
    url: '/system/smtp/test',
    method: 'post',
    data: { to }
  })
}

/* ==================== 版本信息 ==================== */

export function getVersion() {
  return request<VersionInfo>({
    url: '/system/version',
    method: 'get'
  })
}

export function checkUpdate() {
  return request<VersionInfo>({
    url: '/system/version/check',
    method: 'post'
  })
}

/* ==================== 系统信息 ==================== */

export function getSystemInfo() {
  return request<SystemInfo>({
    url: '/system/info',
    method: 'get'
  })
}

/* ==================== 清理工具 ==================== */

export function cleanSystem(cleanType: CleanType) {
  return request<CleanResult>({
    url: '/system/clean',
    method: 'post',
    // 后端 CleanRequest 绑定的字段名是 clean_type（oneof=cache docker logs）
    data: { clean_type: cleanType }
  })
}
