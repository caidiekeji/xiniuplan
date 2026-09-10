import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface ProcessInfo {
  pid: number
  name: string
  cpu: number
  memory: number
  memory_rss: number
  status: 'running' | 'sleeping' | 'stopped' | 'zombie'
  user?: string
}

export interface DiskMount {
  mount_point: string
  total: number
  used: number
  available: number
  usage_percent: number
  filesystem?: string
}

export interface NetworkStats {
  total_connections: number
  established: number
  listening: number
  time_wait: number
  close_wait: number
  other: number
  upload_speed: number
  download_speed: number
}

export interface NetworkConnection {
  protocol: string
  local_address: string
  remote_address: string
  state: string
  pid?: number
}

export interface MonitorHistoryPoint {
  timestamp: number
  cpu: number
  memory: number
}

export interface MonitorHistoryData {
  points: MonitorHistoryPoint[]
}

/* ==================== 进程管理 ==================== */

export function getProcesses(top?: number, sort?: 'cpu' | 'memory') {
  const params: Record<string, any> = {}
  if (top !== undefined) params.top = top
  if (sort) params.sort = sort
  return request<ProcessInfo[]>({
    url: '/monitor/processes',
    method: 'get',
    params
  })
}

export function killProcess(pid: number) {
  return request({
    url: `/monitor/processes/${pid}/kill`,
    method: 'post'
  })
}

/* ==================== 磁盘分析 ==================== */

export function getDiskUsage() {
  return request<DiskMount[]>({
    url: '/monitor/disk',
    method: 'get'
  })
}

/* ==================== 网络统计 ==================== */

export function getNetworkStats() {
  return request<NetworkStats>({
    url: '/monitor/network',
    method: 'get'
  })
}

/* ==================== 历史趋势 ==================== */

export function getMonitorHistory(type: string, range: string) {
  return request<MonitorHistoryData>({
    url: '/monitor/history',
    method: 'get',
    params: { type, range }
  })
}
