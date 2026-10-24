import { request } from './request'

/* ==================== 类型定义 ==================== */

export type DNSProviderType = 'aliyun' | 'tencent' | 'cloudflare'

export interface DNSAccount {
  id: number
  name: string
  provider_type: DNSProviderType
  enabled: boolean
  domains: string[]
  created_at: string
}

export interface DNSAccountPayload {
  name: string
  provider_type: DNSProviderType
  enabled: boolean
  access_key_id?: string
  access_key_secret?: string
  secret_id?: string
  secret_key?: string
  api_token?: string
}

export interface DNSRecord {
  id?: string
  type: string
  name: string
  value: string
  ttl: number
}

export interface DNSRecordPayload {
  type: string
  name: string
  value: string
  ttl: number
}

/* ==================== DNS 账号管理 ==================== */

export function getDNSAccounts() {
  return request<DNSAccount[]>({
    url: '/dns/accounts',
    method: 'get'
  })
}

/**
 * 凭据字段组装。
 *
 * 后端 dto.CreateDNSAccountRequest 要求 credentials 是一个必填的 map
 * （backend/internal/dto/dns.go），此前前端提交的是扁平字段
 * （access_key_id / secret_id / api_token …），必然 400
 * （2026-10-23 闭环审计）。
 */
function toDNSCredentials(data: Partial<DNSAccountPayload>): Record<string, string> {
  const creds: Record<string, string> = {}
  if (data.access_key_id) creds.access_key_id = data.access_key_id
  if (data.access_key_secret) creds.access_key_secret = data.access_key_secret
  if (data.secret_id) creds.secret_id = data.secret_id
  if (data.secret_key) creds.secret_key = data.secret_key
  if (data.api_token) creds.api_token = data.api_token
  return creds
}

export function createDNSAccount(data: DNSAccountPayload) {
  return request<DNSAccount>({
    url: '/dns/accounts',
    method: 'post',
    data: {
      name: data.name,
      provider_type: data.provider_type,
      credentials: toDNSCredentials(data)
    }
  })
}

export function updateDNSAccount(id: number, data: Partial<DNSAccountPayload>) {
  return request<DNSAccount>({
    url: `/dns/accounts/${id}`,
    method: 'put',
    data: {
      ...(data.name !== undefined ? { name: data.name } : {}),
      ...(data.enabled !== undefined ? { enabled: data.enabled } : {}),
      credentials: toDNSCredentials(data)
    }
  })
}

export function deleteDNSAccount(id: number) {
  return request({
    url: `/dns/accounts/${id}`,
    method: 'delete'
  })
}

/* ==================== 域名管理 ==================== */

export function getDNSDomains(id: number) {
  return request<string[]>({
    url: `/dns/accounts/${id}/domains`,
    method: 'get'
  })
}

/* ==================== DNS 记录管理 ==================== */

export function getDNSRecords(id: number, domain: string) {
  return request<DNSRecord[]>({
    url: `/dns/accounts/${id}/domains/${domain}/records`,
    method: 'get'
  })
}

export function addDNSRecord(id: number, domain: string, data: DNSRecordPayload) {
  return request<DNSRecord>({
    url: `/dns/accounts/${id}/domains/${domain}/records`,
    method: 'post',
    data
  })
}

export function deleteDNSRecord(id: number, domain: string, recordType: string, name: string) {
  return request({
    url: `/dns/accounts/${id}/domains/${domain}/records`,
    method: 'delete',
    params: { type: recordType, name }
  })
}
