import { request } from './request'

/* ==================== 类型定义 ==================== */

export type RuleType = 'port' | 'ip'
export type RuleDirection = 'inbound' | 'outbound'
export type RuleProtocol = 'tcp' | 'udp' | 'icmp' | 'all'
export type RuleAction = 'accept' | 'drop'

export interface FirewallRule {
  id: number
  type: RuleType
  direction: RuleDirection
  protocol: RuleProtocol
  port?: number
  ip_address?: string
  action: RuleAction
  status: string
  remark?: string
  created_at: string
}

export interface FirewallRuleListData {
  items: FirewallRule[]
  total: number
}

export interface FirewallRuleParams {
  page?: number
  pageSize?: number
  type?: RuleType
}

export interface AddFirewallRuleData {
  type: RuleType
  direction: RuleDirection
  protocol: RuleProtocol
  port?: number
  ip_address?: string
  action: RuleAction
  remark?: string
}

export interface SecuritySuggestion {
  id: number
  port: number
  service: string
  level: 'high' | 'medium' | 'low'
  description: string
  suggestion: string
}

/* ==================== 规则管理 ==================== */

export function getFirewallRules(page = 1, pageSize = 20, type?: RuleType) {
  const params: FirewallRuleParams = { page, pageSize }
  if (type) params.type = type
  return request<FirewallRuleListData>({
    url: '/firewall/rules',
    method: 'get',
    params
  })
}

export function addFirewallRule(data: AddFirewallRuleData) {
  return request<FirewallRule>({
    url: '/firewall/rules',
    method: 'post',
    data
  })
}

export function deleteFirewallRule(id: number | string) {
  return request({
    url: `/firewall/rules/${id}`,
    method: 'delete'
  })
}

export function toggleFirewallRule(id: number | string) {
  return request({
    url: `/firewall/rules/${id}/toggle`,
    method: 'post'
  })
}

/* ==================== 安全建议 ==================== */

export function getSecuritySuggestions() {
  return request<SecuritySuggestion[]>({
    url: '/firewall/suggestions',
    method: 'get'
  })
}
