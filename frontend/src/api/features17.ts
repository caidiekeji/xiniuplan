import { request } from './request'
import i18n from '@/locales'

// roadmap 17 项功能 API 汇总：拨测 / 守护 / Git部署 / Compose / 镜像构建 /
// AI诊断 / 防篡改 / 回收站 / 面板升级 / 多节点

async function http<T = any>(config: Parameters<typeof request>[0]): Promise<T> {
  const res = await request<T>(config)
  return res.data
}

/* ==================== 网站拨测（#6） ==================== */
export interface Probe {
  id: number
  name: string
  type: 'http' | 'tcp'
  target: string
  interval_sec: number
  keyword: string
  enabled: boolean
  last_status: 'up' | 'down' | 'unknown'
  last_code: number
  last_latency_ms: number
  last_checked: string | null
  created_at: string
}
export interface ProbeResult {
  id: number
  probe_id: number
  status: 'up' | 'down'
  code: number
  latency_ms: number
  message: string
  checked_at: string
}
export const listProbes = () => http<{ items: Probe[] }>({ url: '/probes', method: 'get' })
export const createProbe = (data: Partial<Probe>) => http<Probe>({ url: '/probes', method: 'post', data })
export const updateProbe = (id: number, data: Partial<Probe>) =>
  http({ url: `/probes/${id}`, method: 'put', data })
export const deleteProbe = (id: number) => http({ url: `/probes/${id}`, method: 'delete' })
export const checkProbeNow = (id: number) => http({ url: `/probes/${id}/check`, method: 'post' })
export const listProbeResults = (id: number, limit = 100) =>
  http<{ items: ProbeResult[] }>({ url: `/probes/${id}/results`, method: 'get', params: { limit } })

/* ==================== 进程守护（#11） ==================== */
export interface DaemonProcess {
  id: number
  name: string
  command: string
  work_dir: string
  enabled: boolean
  auto_start: boolean
  pid: number
  restarts: number
  last_started: string | null
  last_exit_code: number
  created_at: string
}
export const listDaemons = () => http<{ items: DaemonProcess[] }>({ url: '/daemon', method: 'get' })
export const createDaemon = (data: Partial<DaemonProcess>) =>
  http<DaemonProcess>({ url: '/daemon', method: 'post', data })
export const updateDaemon = (id: number, data: Partial<DaemonProcess>) =>
  http({ url: `/daemon/${id}`, method: 'put', data })
export const deleteDaemon = (id: number) => http({ url: `/daemon/${id}`, method: 'delete' })
export const startDaemon = (id: number) => http<DaemonProcess>({ url: `/daemon/${id}/start`, method: 'post' })
export const stopDaemon = (id: number) => http({ url: `/daemon/${id}/stop`, method: 'post' })
export const getDaemonLogs = (id: number, lines = 200) =>
  http<{ content: string }>({ url: `/daemon/${id}/logs`, method: 'get', params: { lines } })

/* ==================== Git 推送部署（#9） ==================== */
export interface DeployApp {
  id: number
  name: string
  repo_url: string
  branch: string
  secret: string
  build_type: 'static' | 'node' | 'dockerfile'
  out_dir: string
  dest_path: string
  container_port: number
  enabled: boolean
  last_deploy_at: string | null
  last_status: 'success' | 'failed' | 'deploying' | ''
  created_at: string
}
export interface DeployLog {
  id: number
  app_id: number
  status: 'success' | 'failed'
  output: string
  created_at: string
}
export const listDeployApps = () => http<{ items: DeployApp[] }>({ url: '/deploy/apps', method: 'get' })
export const createDeployApp = (data: Partial<DeployApp>) =>
  http<DeployApp>({ url: '/deploy/apps', method: 'post', data })
export const updateDeployApp = (id: number, data: Partial<DeployApp>) =>
  http({ url: `/deploy/apps/${id}`, method: 'put', data })
export const deleteDeployApp = (id: number) => http({ url: `/deploy/apps/${id}`, method: 'delete' })
export const runDeploy = (id: number) => http<{ output: string }>({ url: `/deploy/apps/${id}/run`, method: 'post' })
export const listDeployLogs = (id: number, limit = 20) =>
  http<{ items: DeployLog[] }>({ url: `/deploy/apps/${id}/logs`, method: 'get', params: { limit } })

/* ==================== Compose 编排（#7） ==================== */
export interface ComposeStack {
  name: string
  running: number
  total: number
  file_path: string
  content?: string
}
export const listComposeStacks = () => http<{ items: ComposeStack[] }>({ url: '/compose', method: 'get' })
export const getComposeStack = (name: string) => http<ComposeStack>({ url: `/compose/${name}`, method: 'get' })
export const saveComposeStack = (name: string, content: string) =>
  http({ url: '/compose', method: 'post', data: { name, content } })
export const deleteComposeStack = (name: string) =>
  http<{ output: string }>({ url: `/compose/${name}`, method: 'delete' })
export const composeAction = (name: string, action: 'up' | 'down' | 'start' | 'stop' | 'restart') =>
  http<{ output: string }>({ url: `/compose/${name}/${action}`, method: 'post' })

/* ==================== 镜像在线构建（#8，流式日志） ==================== */
export async function buildImage(
  tag: string,
  dockerfile: string,
  onLine: (line: string) => void,
  signal?: AbortSignal
): Promise<void> {
  const base = (window as any).__PANEL_BASE__ || ''
  const authStore = (await import('@/stores/auth')).useAuthStore()
  const resp = await fetch(`${base}/api/v1/images/build`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authStore.token}`
    },
    body: JSON.stringify({ tag, dockerfile }),
    signal
  })
  if (!resp.ok || !resp.body) {
    throw new Error(i18n.global.t('common.errors.buildRequestFailed', { status: resp.status }))
  }
  const reader = resp.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    const lines = buf.split('\n')
    buf = lines.pop() ?? ''
    for (const line of lines) {
      if (line.trim()) onLine(line)
    }
  }
  if (buf.trim()) onLine(buf)
}

/* ==================== AI 运维助手（#14） ==================== */
export interface DiagnoseResult {
  analysis: string
  model: string
  latency_ms: number
}
export const diagnoseAI = (data: { type: string; content: string; question?: string }) =>
  http<DiagnoseResult>({ url: '/ai/diagnose', method: 'post', data })

/* ==================== 网站防篡改（#15） ==================== */
export interface TamperChange {
  path: string
  status: 'modified' | 'added' | 'deleted'
  old_hash?: string
  new_hash?: string
}
export const getTamperStatus = (siteId: number) =>
  http<{ has_baseline: boolean; files: number }>({ url: `/sites/${siteId}/integrity/status`, method: 'get' })
export const buildTamperBaseline = (siteId: number) =>
  http<{ files: number; message: string }>({ url: `/sites/${siteId}/integrity/baseline`, method: 'post' })
export const checkTamper = (siteId: number, notify = false) =>
  http<{ changes: TamperChange[]; count: number }>({
    url: `/sites/${siteId}/integrity/check`,
    method: 'get',
    params: { notify: notify ? '1' : '0' }
  })

/* ==================== 文件回收站（#5） ==================== */
export interface RecycleItem {
  id: number
  original_path: string
  recycle_path: string
  size: number
  is_dir: boolean
  deleted_at: string
}
export const listRecycle = () => http<{ items: RecycleItem[] }>({ url: '/files/recycle', method: 'get' })
export const restoreRecycle = (id: number) => http({ url: '/files/recycle/restore', method: 'post', data: { id } })
export const purgeRecycle = (id: number) => http({ url: '/files/recycle', method: 'delete', params: { id } })

/* ==================== 面板一键自升级（#17） ==================== */
export interface UpgradeResult {
  downloaded_bytes: number
  sha256: string
  message: string
}
export const upgradePanel = () => http<UpgradeResult>({ url: '/system/upgrade', method: 'post' })

/* ==================== 多节点管理（#13） ==================== */
export interface NodeItem {
  id: number
  name: string
  base_url: string
  api_key: string
  hostname: string
  version: string
  enabled: boolean
  last_status: 'online' | 'offline' | ''
  cpu_percent: number
  mem_percent: number
  disk_percent: number
  last_error: string
  last_seen_at: string | null
  created_at: string
}
export const listNodes = () => http<{ items: NodeItem[] }>({ url: '/nodes', method: 'get' })
export const addNode = (data: { name: string; base_url: string; api_key: string }) =>
  http<NodeItem>({ url: '/nodes', method: 'post', data })
export const deleteNode = (id: number) => http({ url: `/nodes/${id}`, method: 'delete' })
export const refreshNode = (id: number) => http<NodeItem>({ url: `/nodes/${id}/refresh`, method: 'post' })
export const enableNode = (id: number) => http({ url: `/nodes/${id}/enable`, method: 'post' })
export const disableNode = (id: number) => http({ url: `/nodes/${id}/disable`, method: 'post' })

/* ==================== WAF 站点防护（#10，走站点更新接口） ==================== */
export const updateSiteWAF = (siteId: number, data: { waf_enabled: boolean; cc_rpm?: number; ua_blacklist?: string }) =>
  http({ url: `/sites/${siteId}`, method: 'put', data })
