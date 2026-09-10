import { request } from './request'

/* ==================== 类型定义 ==================== */

export type MetricType = 'cpu' | 'memory' | 'disk' | 'network'
export type AlertCondition = 'gt' | 'lt' | 'eq'
export type ChannelType = 'email' | 'dingtalk' | 'wechat' | 'feishu' | 'telegram' | 'webhook'
export type AlertStatus = 'triggered' | 'recovered'
export type NotificationStatus = 'pending' | 'sent' | 'failed'

export interface AlertRule {
  id: number
  name: string
  metric_type: MetricType
  condition: AlertCondition
  threshold: number
  duration: number
  enabled: boolean
  channels: string[]
  remark?: string
  created_at: string
}

export interface AlertRuleListData {
  items: AlertRule[]
  total: number
}

export interface CreateAlertRuleData {
  name: string
  metric_type: MetricType
  condition: AlertCondition
  threshold: number
  duration: number
  channels: string[]
  remark?: string
}

export type UpdateAlertRuleData = Partial<CreateAlertRuleData>

export interface AlertChannel {
  id: number
  name: string
  type: ChannelType
  config: Record<string, any>
  enabled: boolean
  created_at: string
}

export interface AlertChannelListData {
  items: AlertChannel[]
  total: number
}

export interface CreateAlertChannelData {
  name: string
  type: ChannelType
  config: Record<string, any>
  enabled?: boolean
}

export type UpdateAlertChannelData = Partial<CreateAlertChannelData>

export interface AlertHistory {
  id: number
  rule_id: number
  rule_name: string
  metric_type: MetricType
  trigger_value: number
  threshold: number
  status: AlertStatus
  triggered_at: string
  recovered_at?: string
  notification_status: NotificationStatus
  notification_error?: string
}

export interface AlertHistoryListData {
  items: AlertHistory[]
  total: number
}

export interface AlertHistoryParams {
  rule_id?: number
  status?: AlertStatus
  page: number
  pageSize: number
}

/* ==================== 告警规则 API ==================== */

export function getAlertRules(page: number, pageSize: number) {
  return request<AlertRuleListData>({
    url: '/alert/rules',
    method: 'get',
    params: { page, pageSize }
  })
}

export function createAlertRule(data: CreateAlertRuleData) {
  return request<AlertRule>({
    url: '/alert/rules',
    method: 'post',
    data
  })
}

export function getAlertRule(id: number | string) {
  return request<AlertRule>({
    url: `/alert/rules/${id}`,
    method: 'get'
  })
}

export function updateAlertRule(id: number | string, data: UpdateAlertRuleData) {
  return request<AlertRule>({
    url: `/alert/rules/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlertRule(id: number | string) {
  return request({
    url: `/alert/rules/${id}`,
    method: 'delete'
  })
}

export function enableAlertRule(id: number | string) {
  return request({
    url: `/alert/rules/${id}/enable`,
    method: 'post'
  })
}

export function disableAlertRule(id: number | string) {
  return request({
    url: `/alert/rules/${id}/disable`,
    method: 'post'
  })
}

export function testAlertRule(id: number | string) {
  return request({
    url: `/alert/rules/${id}/test`,
    method: 'post'
  })
}

/* ==================== 通知渠道 API ==================== */

export function getAlertChannels() {
  return request<AlertChannelListData>({
    url: '/alert/channels',
    method: 'get'
  })
}

export function createAlertChannel(data: CreateAlertChannelData) {
  return request<AlertChannel>({
    url: '/alert/channels',
    method: 'post',
    data
  })
}

export function updateAlertChannel(id: number | string, data: UpdateAlertChannelData) {
  return request<AlertChannel>({
    url: `/alert/channels/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlertChannel(id: number | string) {
  return request({
    url: `/alert/channels/${id}`,
    method: 'delete'
  })
}

export function testAlertChannel(id: number | string) {
  return request({
    url: `/alert/channels/${id}/test`,
    method: 'post'
  })
}

/* ==================== 告警历史 API ==================== */

export function getAlertHistory(params: AlertHistoryParams) {
  return request<AlertHistoryListData>({
    url: '/alert/history',
    method: 'get',
    params
  })
}
