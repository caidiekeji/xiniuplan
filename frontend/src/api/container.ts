import { request } from './request'

/* ==================== 类型定义 ==================== */

export type ContainerStatus = 'running' | 'stopped' | 'exited' | 'created' | 'paused' | 'restarting'

export interface PortMapping {
  host_ip?: string
  host_port: number
  container_port: number
  protocol: 'tcp' | 'udp'
}

export interface Container {
  id: string
  name: string
  image: string
  image_id?: string
  status: ContainerStatus
  state?: string
  ports: PortMapping[]
  created_at: string
  started_at?: string
  command?: string
  labels?: Record<string, string>
  mounts?: any[]
  networks?: Record<string, any>
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
  repo_tags: string[]
  size: number
  created_at: string
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
