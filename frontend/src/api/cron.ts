import { request } from './request'

/* ==================== 类型定义 ==================== */

export type CronTaskType = 'shell' | 'backup' | 'url' | 'app_restart'
export type CronTaskStatus = 'enabled' | 'disabled'

export interface CronTask {
  id: number
  name: string
  type: CronTaskType
  schedule: string
  command?: string
  target?: string
  status: CronTaskStatus
  lastRunAt?: string
  nextRunAt?: string
  nextRunTime?: string
  remark?: string
  createdAt: string
}

export interface CronTaskListData {
  items: CronTask[]
  total: number
}

export interface CreateCronTaskData {
  name: string
  type: CronTaskType
  schedule: string
  command?: string
  target?: string
  remark?: string
}

export type UpdateCronTaskData = Partial<CreateCronTaskData>

export interface CronLog {
  id: number
  startTime: string
  endTime?: string
  durationMs?: number
  result: 'success' | 'failed' | 'running'
  output?: string
}

export interface CronLogListData {
  items: CronLog[]
  total: number
}

/* ==================== API 方法 ==================== */

export function getCronTasks(page: number, pageSize: number) {
  return request<CronTaskListData>({
    url: '/cron/tasks',
    method: 'get',
    params: { page, pageSize }
  })
}

export function createCronTask(data: CreateCronTaskData) {
  return request<CronTask>({
    url: '/cron/tasks',
    method: 'post',
    data
  })
}

export function updateCronTask(id: number | string, data: UpdateCronTaskData) {
  return request<CronTask>({
    url: `/cron/tasks/${id}`,
    method: 'put',
    data
  })
}

export function deleteCronTask(id: number | string) {
  return request({
    url: `/cron/tasks/${id}`,
    method: 'delete'
  })
}

export function enableCronTask(id: number | string) {
  return request({
    url: `/cron/tasks/${id}/enable`,
    method: 'post'
  })
}

export function disableCronTask(id: number | string) {
  return request({
    url: `/cron/tasks/${id}/disable`,
    method: 'post'
  })
}

export function triggerCronTask(id: number | string) {
  return request({
    url: `/cron/tasks/${id}/trigger`,
    method: 'post'
  })
}

export function getCronLogs(id: number | string, page: number, pageSize: number) {
  return request<CronLogListData>({
    url: `/cron/tasks/${id}/logs`,
    method: 'get',
    params: { page, pageSize }
  })
}
