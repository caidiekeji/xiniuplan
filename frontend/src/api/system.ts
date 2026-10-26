import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface SystemSettings {
  panel_port: number
  domain: string
  // 前端支持三语言（zh-CN / en-US / ja-JP），设置页此前只暴露了前两种
  language: 'zh-CN' | 'en-US' | 'ja-JP'
  theme: 'dark' | 'light' | 'auto'
  security_entry: string
  session_timeout: number
  // Docker 镜像拉取地址（加速源 + 私有仓库凭据）
  docker_registry_mirrors: string
  docker_registry_private: string
  docker_registry_user: string
  docker_registry_pass: string
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

/**
 * 后端 settings 表的实际键名（backend/internal/model/setting.go）
 * 与前端表单字段并非一一对应——域名的库键是 panel_domain，
 * 数值型设置在库里是字符串。此处集中做键名与类型适配，
 * 避免页面提交时写出后端不认识的键（那会新建垃圾设置项）。
 */
interface RawSettings {
  panel_port?: string | number
  panel_domain?: string
  domain?: string
  language?: string
  theme?: string
  security_entry?: string
  session_timeout?: string | number
  docker_registry_mirrors?: string
  docker_registry_private?: string
  docker_registry_user?: string
  docker_registry_pass?: string
}

function toSettings(raw: RawSettings): SystemSettings {
  return {
    panel_port: Number(raw.panel_port ?? 8888),
    domain: raw.panel_domain ?? raw.domain ?? '',
    language: (raw.language ?? 'zh-CN') as SystemSettings['language'],
    theme: (raw.theme ?? 'dark') as SystemSettings['theme'],
    security_entry: raw.security_entry ?? '',
    session_timeout: Number(raw.session_timeout ?? 120),
    docker_registry_mirrors: raw.docker_registry_mirrors ?? '',
    docker_registry_private: raw.docker_registry_private ?? '',
    docker_registry_user: raw.docker_registry_user ?? '',
    docker_registry_pass: raw.docker_registry_pass ?? ''
  }
}

function toSettingsPayload(settings: Partial<SystemSettings>) {
  const payload: Record<string, string> = {}
  if (settings.panel_port !== undefined) payload.panel_port = String(settings.panel_port)
  if (settings.domain !== undefined) payload.panel_domain = settings.domain
  if (settings.language !== undefined) payload.language = settings.language
  if (settings.theme !== undefined) payload.theme = settings.theme
  if (settings.security_entry !== undefined) payload.security_entry = settings.security_entry
  if (settings.session_timeout !== undefined) payload.session_timeout = String(settings.session_timeout)
  if (settings.docker_registry_mirrors !== undefined) payload.docker_registry_mirrors = settings.docker_registry_mirrors
  if (settings.docker_registry_private !== undefined) payload.docker_registry_private = settings.docker_registry_private
  if (settings.docker_registry_user !== undefined) payload.docker_registry_user = settings.docker_registry_user
  if (settings.docker_registry_pass !== undefined) payload.docker_registry_pass = settings.docker_registry_pass
  return payload
}

export async function getSettings() {
  // 后端返回 { settings: {...} }（dto.SettingsResponse），实际值在 settings 下
  const res = await request<{ settings?: RawSettings }>({
    url: '/system/settings',
    method: 'get'
  })
  return { ...res, data: toSettings(res.data?.settings || {}) }
}

export function updateSettings(settings: Partial<SystemSettings>) {
  return request<{ updated: boolean; port_changed?: boolean; port_message?: string }>({
    url: '/system/settings',
    method: 'put',
    // 后端 dto.UpdateSettingsRequest 要求包一层 settings（binding:"required"），
    // 直接提交 map 会 400 —— 设置保存按钮此前一直是坏的（2026-10-23 闭环审计）
    data: { settings: toSettingsPayload(settings) }
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
