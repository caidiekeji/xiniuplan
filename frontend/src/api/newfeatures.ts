import { request } from './request'

// 自动解包 ApiResponse.data
async function http<T = any>(config: Parameters<typeof request>[0]): Promise<T> {
  const res = await request<T>(config)
  return res.data
}

/* ==================== 新增功能 API 汇总 ==================== */

/* ---------- 安全评分（D-007） ---------- */
export interface ScoreCheck {
  key: string
  label: string
  score: number
  max: number
  status: 'good' | 'warn' | 'bad'
  advice: string
}
export interface SecurityScore {
  total: number
  grade: string
  checks: ScoreCheck[]
}
export const getSecurityScore = () =>
  http<SecurityScore>({ url: '/dashboard/security-score', method: 'get' })

/* ---------- 系统更新（SY-005） ---------- */
export interface PackageUpdate {
  name: string
  current?: string
  new: string
  source: string
}
export const checkSystemUpdates = () =>
  http<{ package_manager: string; items: PackageUpdate[]; total: number }>({ url: '/system/updates', method: 'get' })
export const installSystemUpdates = (confirm: boolean) =>
  http<{ installed: boolean; output: string }>({ url: '/system/updates/install', method: 'post', data: { confirm } })

/* ---------- NTP（SY-006） ---------- */
export interface NTPStatus {
  ntp_enabled: boolean
  synced: boolean
  timezone: string
  local_time: string
}
export const getNTPStatus = () => http<NTPStatus>({ url: '/system/ntp', method: 'get' })
export const setNTPStatus = (enabled: boolean) =>
  http({ url: '/system/ntp', method: 'put', data: { enabled } })

/* ---------- 远程备份存储（BK-003） ---------- */
export interface RemoteStorage {
  id: number
  name: string
  type: 's3' | 'webdav' | 'sftp' | 'panel'
  config: Record<string, string>
  enabled: boolean
  created_at: string
}
export const listRemoteStorages = () => http<RemoteStorage[]>({ url: '/remote-storages', method: 'get' })
export const createRemoteStorage = (data: { name: string; type: string; config: Record<string, string>; enabled: boolean }) =>
  http<RemoteStorage>({ url: '/remote-storages', method: 'post', data })
export const updateRemoteStorage = (id: number, data: Record<string, any>) =>
  http<RemoteStorage>({ url: `/remote-storages/${id}`, method: 'put', data })
export const deleteRemoteStorage = (id: number) =>
  http({ url: `/remote-storages/${id}`, method: 'delete' })
export const testRemoteStorage = (id: number) =>
  http({ url: `/remote-storages/${id}/test`, method: 'post' })
export const pushBackup = (backupId: number, remoteId: number) =>
  http({ url: `/backups/${backupId}/push`, method: 'post', data: { remote_id: remoteId } })

/* ---------- 端口转发（S-004，走 firewall rules type=forward） ---------- */
export const createForwardRule = (data: {
  type: 'forward'
  direction: string
  protocol: 'tcp' | 'udp'
  port: number
  dest_ip: string
  dest_port: number
  action: 'accept'
  remark?: string
}) => http({ url: '/firewall/rules', method: 'post', data })

/* ---------- 在线会话（U-006） ---------- */
export interface OnlineSession {
  token_hash: string
  user_id: number
  username: string
  ip: string
  user_agent: string
  token_type: string
  login_at: string
  last_active: string
  expires_at: string
}
export const listSessions = () => http<OnlineSession[]>({ url: '/auth/sessions', method: 'get' })
export const revokeSession = (hash: string) =>
  http({ url: `/auth/sessions/${hash}`, method: 'delete' })

/* ---------- API Key（U-007） ---------- */
export interface APIKey {
  id: number
  name: string
  key_id: string
  secret?: string
  user_id: number
  status: string
  last_used_at?: string
  expires_at?: string
  created_at: string
}
export const listAPIKeys = () => http<APIKey[]>({ url: '/auth/apikeys', method: 'get' })
export const createAPIKey = (data: { name: string; permissions?: string; expires_in_days?: number }) =>
  http<APIKey>({ url: '/auth/apikeys', method: 'post', data })
export const updateAPIKey = (id: number, data: { status?: string; permissions?: string }) =>
  http<APIKey>({ url: `/auth/apikeys/${id}`, method: 'put', data })
export const deleteAPIKey = (id: number) => http({ url: `/auth/apikeys/${id}`, method: 'delete' })

/* ---------- 容器增强（C-004/006/008/009） ---------- */
export interface DockerNetwork {
  id: string
  name: string
  driver: string
  scope: string
  internal: boolean
  subnet: string
  gateway: string
  containers: string[]
  built_in: boolean
}
export interface DockerVolume {
  name: string
  driver: string
  mountpoint: string
  size: number
  created_at: string
  in_use: boolean
  used_by: string
}
export const listNetworks = () => http<{ items: DockerNetwork[] }>({ url: '/networks', method: 'get' })
export const createNetwork = (data: { name: string; driver?: string; subnet?: string; gateway?: string }) =>
  http({ url: '/networks', method: 'post', data })
export const removeNetwork = (id: string) => http({ url: `/networks/${id}`, method: 'delete' })
export const listVolumes = () => http<{ items: DockerVolume[] }>({ url: '/volumes', method: 'get' })
export const createVolume = (data: { name: string; driver?: string }) => http({ url: '/volumes', method: 'post', data })
export const removeVolume = (name: string, force?: boolean) =>
  http({ url: `/volumes/${name}`, method: 'delete', params: { force: force ? 'true' : 'false' } })
export const removeImage = (id: string, force?: boolean) =>
  http({ url: `/images/${id}`, method: 'delete', params: { force: force ? 'true' : 'false' } })

/* ---------- 进程终止（M-002） ---------- */
export const killProcess = (pid: number, signal?: number) =>
  http({ url: '/monitor/processes/kill', method: 'post', data: { pid, signal } })

/* ---------- 数据库参数与状态（DB-002/004） ---------- */
export interface VarEntry { name: string; value: string }
export interface DBStatus {
  current_qps: number
  threads_connected: number
  max_connections: number
  slow_queries: number
  uptime: number
  extra: VarEntry[]
  pg?: {
    num_backends: number
    xact_committed: number
    xact_rolled_back: number
    blocks_read: number
    blocks_hit: number
    cache_hit_ratio: string
    deadlocks: number
    temp_files: number
  }
}
export const listDBVariables = (id: number, search?: string) =>
  http<VarEntry[]>({ url: `/databases/${id}/variables`, method: 'get', params: { search } })
export const updateDBVariable = (id: number, name: string, value: string) =>
  http({ url: `/databases/${id}/variables`, method: 'put', data: { name, value } })
export const getDBStatus = (id: number) => http<DBStatus>({ url: `/databases/${id}/status`, method: 'get' })

/* ---------- 网站增强（W-006~009、W-018） ---------- */
export interface SiteSnippet { name: string; content: string }
export const listRewriteTemplates = () =>
  http<{ name: string; content: string }[]>({ url: '/sites/rewrite-templates', method: 'get' })
export const listSnippets = (siteId: number) =>
  http<SiteSnippet[]>({ url: `/sites/${siteId}/snippets`, method: 'get' })
export const setSnippet = (siteId: number, data: Record<string, any>) =>
  http({ url: `/sites/${siteId}/snippets`, method: 'post', data })
export const deleteSnippet = (siteId: number, name: string) =>
  http({ url: `/sites/${siteId}/snippets/${name}`, method: 'delete' })
export const addBasicAuthUser = (siteId: number, username: string, password: string) =>
  http({ url: `/sites/${siteId}/basic-auth`, method: 'post', data: { username, password } })

export interface TrafficSummary {
  range: string
  total_pv: number
  total_uv: number
  unique_ips: number
  top_urls: { key: string; count: number }[]
  top_ips: { key: string; count: number }[]
  status_2xx: number
  status_4xx: number
  status_5xx: number
  buckets: { time: string; pv: number; uv: number }[]
}
export const getSiteTraffic = (siteId: number, range = '24h') =>
  http<TrafficSummary>({ url: `/sites/${siteId}/traffic`, method: 'get', params: { range } })

/* ---------- 异常登录（S-012） ---------- */
export interface AnomalyEvent {
  type: 'brute_force_ip' | 'brute_force_user' | 'new_ip_login'
  subject: string
  count: number
  first_seen: string
  last_seen: string
  detail: string
}
export const listAnomalies = () =>
  http<{ items: AnomalyEvent[]; total: number }>({ url: '/audit/anomalies', method: 'get' })

/* ---------- 终端增强（T-004/006） ---------- */
export interface QuickCommand { label: string; command: string; desc: string }
export const listQuickCommands = () => http<QuickCommand[]>({ url: '/terminal/quick-commands', method: 'get' })
export interface TerminalRecord { name: string; size: number; time: string }
export const listTerminalRecords = () => http<TerminalRecord[]>({ url: '/terminal/records', method: 'get' })
export const getTerminalRecord = (name: string) =>
  http<{ name: string; content: string }>({ url: `/terminal/records/${name}`, method: 'get' })
export const deleteTerminalRecord = (name: string) =>
  http({ url: `/terminal/records/${name}`, method: 'delete' })

/* ---------- 文件权限（F-005） ---------- */
export const setFilePermission = (data: {
  path: string
  mode?: string
  owner?: string
  group?: string
  recursive?: boolean
}) => http({ url: '/files/permission', method: 'post', data })

/* ---------- 应用升级（A-005） ---------- */
export const upgradeApp = (key: string) =>
  http<{ key: string; version: string; status: string }>({ url: `/apps/${key}/upgrade`, method: 'post' })
