import { request } from './request'

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
