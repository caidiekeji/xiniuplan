import { request } from './request'
import { useAuthStore } from '@/stores/auth'

/* ==================== 类型定义 ==================== */

export type ContainerStatus = 'running' | 'stopped' | 'exited' | 'created' | 'paused' | 'restarting'

export interface Container {
  id: string
  name: string
  image: string
  /** Docker 原始状态描述，如 "Up 24 hours"，仅用于展示 */
  status: string
  /** 归一化状态：running / exited / paused / created / restarting，用于状态判定 */
  state: ContainerStatus
  /** 端口映射：后端为逗号分隔的字符串，如 "127.0.0.1:15432->5432/tcp, 80/tcp"；未发布端口时为空串 */
  ports: string
  /** 创建时间（Unix 秒） */
  created: number
  labels?: Record<string, string>
}

export interface ContainerListData {
  items: Container[]
  total: number
}

export interface ContainerDetail extends Container {
  env?: string[]
  volumes?: any[]
  network_settings?: any
  host_config?: any
}

export interface ContainerLogData {
  content: string
}

export interface Image {
  id: string
  /** 仓库标签，如 ["nginx:latest"]；悬空镜像为空数组 */
  tags: string[]
  size: number
  /** 创建时间（Unix 秒） */
  created: number
  labels?: Record<string, string>
}

/* ==================== 容器管理 ==================== */

export function getContainers(all = false) {
  return request<ContainerListData>({
    url: '/containers',
    method: 'get',
    params: { all }
  })
}

export function getContainer(id: string) {
  return request<ContainerDetail>({
    url: `/containers/${id}`,
    method: 'get'
  })
}

export function startContainer(id: string) {
  return request({
    url: `/containers/${id}/start`,
    method: 'post'
  })
}

export function stopContainer(id: string) {
  return request({
    url: `/containers/${id}/stop`,
    method: 'post'
  })
}

export function restartContainer(id: string) {
  return request({
    url: `/containers/${id}/restart`,
    method: 'post'
  })
}

export function removeContainer(id: string, force = false) {
  return request({
    url: `/containers/${id}`,
    method: 'delete',
    params: { force }
  })
}

export function getContainerLogs(id: string, tail = 200, follow = false) {
  return request<ContainerLogData>({
    url: `/containers/${id}/logs`,
    method: 'get',
    params: { tail, follow }
  })
}

/* ==================== 镜像管理 ==================== */

export function getImages() {
  return request<Image[]>({
    url: '/images',
    method: 'get'
  })
}

export function removeImage(id: string, force = false) {
  return request({
    url: `/images/${id}`,
    method: 'delete',
    params: { force }
  })
}

/* ==================== 镜像拉取 ==================== */

/** 仓库选项（含直连、加速源、私有仓库） */
export interface RegistryOption {
  /** 选中值：'' 直连 / '@private' 私有仓库 / 其余为加速源地址 */
  id: string
  name: string
}

export interface RegistryListData {
  registries: RegistryOption[]
  has_private: boolean
  private_addr: string
  private_user: string
}

/** 拉取进度帧 */
export interface PullProgress {
  layer_id?: string
  status?: string
  current?: number
  total?: number
  percent?: number
  /** 出错帧 */
  error?: string
  /** 完成帧 */
  image_ref?: string
  message?: string
}

/** 获取可用的仓库列表 */
export function getRegistries() {
  return request<RegistryListData>({
    url: '/images/registries',
    method: 'get'
  })
}

/**
 * 拉取镜像（SSE 流式进度）。
 *
 * 为什么不用 axios：响应是永不结束的流式帧，axios 会挂起到超时。
 * 为什么不用 EventSource：它不支持 Authorization 头，带不了登录 token。
 * 与容器日志跟随同一套实现（utils/sse.ts）。
 */
export function pullImage(
  data: { image: string; mirror?: string; username?: string; password?: string },
  onProgress: (p: PullProgress) => void,
  onError?: (e: Error) => void
): () => void {
  const base = `${(window as any).__PANEL_BASE__ || ''}/api/v1`
  const token = useAuthStore().token
  const controller = new AbortController()

  ;(async () => {
    try {
      const resp = await fetch(`${base}/images/pull`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'text/event-stream',
          ...(token ? { Authorization: `Bearer ${token}` } : {})
        },
        body: JSON.stringify(data),
        signal: controller.signal
      })
      if (!resp.ok || !resp.body) throw new Error(`拉取请求失败: HTTP ${resp.status}`)

      const reader = resp.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        let idx: number
        while ((idx = buffer.indexOf('\n')) >= 0) {
          const line = buffer.slice(0, idx).replace(/\r$/, '')
          buffer = buffer.slice(idx + 1)
          if (!line.startsWith('data:')) continue
          const payload = line.slice(5).replace(/^ /, '')
          if (!payload) continue
          try {
            onProgress(JSON.parse(payload) as PullProgress)
          } catch {
            // 非 JSON 帧（如心跳），忽略
          }
        }
      }
    } catch (e: any) {
      if (e?.name === 'AbortError') return
      onError?.(e instanceof Error ? e : new Error(String(e)))
    }
  })()

  return () => controller.abort()
}

export interface ContainerStatsData {
  cpu_percent: number
  memory_usage: string
  memory_limit: string
  memory_percent: number
}

// 获取容器实时资源统计
export function getContainerStats(id: string) {
  return request<ContainerStatsData>({
    url: `/containers/${id}/stats`,
    method: 'get'
  })
}
