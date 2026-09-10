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

export function createDNSAccount(data: DNSAccountPayload) {
  return request<DNSAccount>({
    url: '/dns/accounts',
    method: 'post',
    data
  })
}

export function updateDNSAccount(id: number, data: Partial<DNSAccountPayload>) {
  return request<DNSAccount>({
    url: `/dns/accounts/${id}`,
    method: 'put',
    data
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
