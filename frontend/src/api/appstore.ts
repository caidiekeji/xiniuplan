import { request } from './request'

/* ==================== 类型定义 ==================== */

export type AppCategory =
  | 'website'
  | 'devtools'
  | 'database'
  | 'middleware'
  | 'monitor'
  | 'office'
  | 'storage'
  | 'network'

export type AppStatus = 'installed' | 'installing' | 'failed' | 'not_installed'

export type AppSourceType = 'official' | 'community' | 'private'

export interface EnvParam {
  key: string
  label: string
  type: 'string' | 'password' | 'number' | 'select'
  default?: string | number
  required?: boolean
  description?: string
  options?: { label: string; value: string }[]
  placeholder?: string
}

export interface ResourceRequirement {
  memory?: string
  disk?: string
  cpu?: string
}

export interface App {
  key: string
  name: string
  version: string
  category: AppCategory
  description: string
  icon?: string
  author?: string
  homepage?: string
  resources?: ResourceRequirement
  env_params: EnvParam[]
  ports?: number[]
  volumes?: string[]
  tags?: string[]
  source_id?: string
  source_name?: string
}

export interface AppListData {
  items: App[]
  total: number
}

export interface InstalledApp {
  key: string
  name: string
  version: string
  category: AppCategory
  status: AppStatus
  container_id?: string
  access_url?: string
  initial_password?: string
  installed_at: string
  env?: Record<string, string>
}

export interface InstallResult {
  success: boolean
  container_id?: string
  access_url?: string
  initial_password?: string
  message?: string
}

export interface AppStatusResult {
  key: string
  status: AppStatus
  container_id?: string
  access_url?: string
  progress?: number
  logs?: string
}

export interface AppSource {
  id: string
  name: string
  type: AppSourceType
  url?: string
  path?: string
  enabled: boolean
  last_sync_at?: string
  created_at: string
}

export interface AppSourceInput {
  name: string
  type: AppSourceType
  url?: string
  path?: string
  enabled?: boolean
}

export interface SyncResult {
  added: number
  updated: number
  deleted: number
}

/* ==================== 应用商店 ==================== */

export function getApps(category?: AppCategory, keyword?: string, sourceId?: string) {
  return request<AppListData>({
    url: '/apps',
    method: 'get',
    params: { category, keyword, source_id: sourceId }
  })
}

export function getApp(key: string) {
  return request<App>({
    url: `/apps/${key}`,
    method: 'get'
  })
}

export function getInstalledApps() {
  return request<InstalledApp[]>({
    url: '/apps/installed',
    method: 'get'
  })
}

export function installApp(key: string, envParams: Record<string, string>, sourceId?: string) {
  return request<InstallResult>({
    url: `/apps/${key}/install`,
    method: 'post',
    data: { env_params: envParams, source_id: sourceId }
  })
}

export function uninstallApp(key: string) {
  return request({
    url: `/apps/${key}/uninstall`,
    method: 'post'
  })
}

export function getAppStatus(key: string) {
  return request<AppStatusResult>({
    url: `/apps/${key}/status`,
    method: 'get'
  })
}

/* ==================== 源管理 ==================== */

export function getAppSources() {
  return request<AppSource[]>({
    url: '/app-sources',
    method: 'get'
  })
}

export function createAppSource(data: AppSourceInput) {
  return request<AppSource>({
    url: '/app-sources',
    method: 'post',
    data
  })
}

export function updateAppSource(id: string, data: AppSourceInput) {
  return request<AppSource>({
    url: `/app-sources/${id}`,
    method: 'put',
    data
  })
}

export function deleteAppSource(id: string) {
  return request({
    url: `/app-sources/${id}`,
    method: 'delete'
  })
}

export function toggleAppSource(id: string) {
  return request<AppSource>({
    url: `/app-sources/${id}/toggle`,
    method: 'post'
  })
}

export function syncAppSource(id: string) {
  return request<SyncResult>({
    url: `/app-sources/${id}/sync`,
    method: 'post'
  })
}
