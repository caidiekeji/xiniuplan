import { request, default as service } from './request'

/* ==================== 类型定义 ==================== */

export type BackupType = 'site' | 'database' | 'system'
export type BackupStatus = 'completed' | 'running' | 'failed'

export interface Backup {
  id: number
  type: BackupType
  target: string
  file_path?: string
  file_name?: string
  size: number
  formatted_size?: string
  status: BackupStatus
  remark?: string
  error_msg?: string
  created_at: string
}

export interface BackupListParams {
  type?: BackupType
  page: number
  pageSize: number
  keyword?: string
}

export interface BackupListData {
  items: Backup[]
  total: number
}

export interface CreateBackupData {
  type: BackupType
  target: string
  remark?: string
}

/* ==================== API 方法 ==================== */

export function getBackups(params: BackupListParams) {
  return request<BackupListData>({
    url: '/backups',
    method: 'get',
    params
  })
}

export function createBackup(data: CreateBackupData) {
  return request<Backup>({
    url: '/backups',
    method: 'post',
    data
  })
}

export function getBackup(id: number | string) {
  return request<Backup>({
    url: `/backups/${id}`,
    method: 'get'
  })
}

export function deleteBackup(id: number | string) {
  return request({
    url: `/backups/${id}`,
    method: 'delete'
  })
}

export function restoreBackup(id: number | string) {
  return request({
    url: `/backups/${id}/restore`,
    method: 'post'
  })
}

export function downloadBackup(id: number | string) {
  return service.get(`/backups/${id}/download`, {
    responseType: 'blob'
  })
}
