import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource: string
  ip: string
  user_agent: string
  result: 'success' | 'failed'
  created_at: string
}

export interface LoginLog {
  id: number
  username: string
  ip: string
  user_agent: string
  result: 'success' | 'failed'
  created_at: string
}

export interface AuditLogParams {
  user_id?: number
  action?: string
  result?: 'success' | 'failed'
  page: number
  pageSize: number
  start_time?: string
  end_time?: string
}

export interface AuditLogListData {
  items: AuditLog[]
  total: number
}

export interface LoginLogListData {
  items: LoginLog[]
  total: number
}

/* ==================== 操作日志 ==================== */

export function getAuditLogs(params: AuditLogParams) {
  return request<AuditLogListData>({
    url: '/audit/logs',
    method: 'get',
    params
  })
}

/* ==================== 登录日志 ==================== */

export function getLoginLogs(page = 1, pageSize = 20) {
  return request<LoginLogListData>({
    url: '/audit/login-logs',
    method: 'get',
    params: { page, pageSize }
  })
}

/* ==================== 导出 ==================== */

export function exportAuditLogs(start?: string, end?: string) {
  const params: Record<string, string> = {}
  if (start) params.start = start
  if (end) params.end = end
  return request<{ download_url: string; filename: string }>({
    url: '/audit/export',
    method: 'get',
    params
  })
}
