import { request } from './request'

/** 后端原始响应类型（snake_case，字节单位） */
export interface OverviewRaw {
  cpu: {
    usage_percent: number
    cores: number
  }
  memory: {
    total: number
    used: number
    free: number
    usage_percent: number
  }
  disk: {
    total: number
    used: number
    free: number
    usage_percent: number
  }
  network: {
    bytes_sent_per_sec: number
    bytes_recv_per_sec: number
  }
  system: {
    hostname: string
    os: string
    kernel: string
    uptime: number
    panel_version: string
    public_ip: string
  }
}

/** 前端组件使用的归一化类型 */
export interface OverviewData {
  cpu: { usage: number; cores: number }
  memory: { usedGB: number; totalGB: number; usage: number }
  disk: { usedGB: number; totalGB: number; usage: number }
  network: { uploadKbps: number; downloadKbps: number }
  uptime: number
  hostname: string
  os: string
  panelVersion: string
}

/** 将后端原始数据归一化为前端使用的格式 */
export function normalizeOverview(raw: OverviewRaw): OverviewData {
  return {
    cpu: { usage: raw.cpu?.usage_percent ?? 0, cores: raw.cpu?.cores ?? 0 },
    memory: {
      usedGB: (raw.memory?.used ?? 0) / 1024 ** 3,
      totalGB: (raw.memory?.total ?? 0) / 1024 ** 3,
      usage: raw.memory?.usage_percent ?? 0
    },
    disk: {
      usedGB: (raw.disk?.used ?? 0) / 1024 ** 3,
      totalGB: (raw.disk?.total ?? 0) / 1024 ** 3,
      usage: raw.disk?.usage_percent ?? 0
    },
    network: {
      uploadKbps: (raw.network?.bytes_sent_per_sec ?? 0) / 1024,
      downloadKbps: (raw.network?.bytes_recv_per_sec ?? 0) / 1024
    },
    uptime: raw.system?.uptime ?? 0,
    hostname: raw.system?.hostname ?? '',
    os: raw.system?.os ?? '',
    panelVersion: raw.system?.panel_version ?? ''
  }
}

export interface MetricPoint {
  timestamp: number
  cpu: number
  memory: number
}

export interface MetricsData {
  points: MetricPoint[]
}

export interface AuditLogItem {
  id: number
  user: string
  action: string
  target: string
  result: 'success' | 'failed'
  timestamp: string
}

export function getOverview() {
  return request<OverviewRaw>({
    url: '/dashboard/overview',
    method: 'get'
  })
}

export function getMetrics(range: string = '24h') {
  return request<MetricsData>({
    url: '/dashboard/metrics',
    method: 'get',
    params: { range }
  })
}

export function getAuditLogs(page = 1, pageSize = 20) {
  return request<{ items: AuditLogItem[]; total: number }>({
    url: '/dashboard/audit-logs',
    method: 'get',
    params: { page, pageSize }
  })
}
