import { request } from './request'

/* ==================== 类型定义 ==================== */

export type RedisKeyType = 'string' | 'hash' | 'list' | 'set' | 'zset' | 'none'

export interface RedisScanParams {
  pattern?: string
  cursor?: number
  count?: number
}

export interface RedisKeyInfo {
  key: string
  type: RedisKeyType
  ttl: number
  size?: number
}

export interface RedisScanResult {
  keys: RedisKeyInfo[]
  cursor: number
  total?: number
}

export interface RedisKeyValue {
  key: string
  type: RedisKeyType
  ttl: number
  size?: number
  value: string | Record<string, string> | string[] | Array<{ member: string; score: number }>
}

export interface RedisSetKeyData {
  type: RedisKeyType
  value: any
  ttl?: number
}

export interface RedisCLIResult {
  result: string
  error?: string
}

export interface RedisMemoryInfo {
  used_memory: number
  used_memory_human: string
  maxmemory: number
  maxmemory_human: string
  used_memory_peak_human?: string
  mem_fragmentation_ratio?: string
  maxmemory_policy: string
  total_keys?: number
}

export interface RedisBigKey {
  key: string
  type: RedisKeyType
  size: number
  size_human: string
  ratio?: number
}

/* ==================== Key 管理 ==================== */

export function scanKeys(id: number | string, params?: RedisScanParams) {
  return request<RedisScanResult>({
    url: `/redis/${id}/keys`,
    method: 'get',
    params
  })
}

export function getKeyValue(id: number | string, key: string) {
  return request<RedisKeyValue>({
    url: `/redis/${id}/keys/${encodeURIComponent(key)}`,
    method: 'get'
  })
}

export function setKeyValue(id: number | string, key: string, data: RedisSetKeyData) {
  return request({
    url: `/redis/${id}/keys/${encodeURIComponent(key)}`,
    method: 'put',
    data
  })
}

export function deleteKey(id: number | string, key: string) {
  return request({
    url: `/redis/${id}/keys/${encodeURIComponent(key)}`,
    method: 'delete'
  })
}

/* ==================== CLI ==================== */

export function execCLI(id: number | string, command: string) {
  return request<RedisCLIResult>({
    url: `/redis/${id}/cli`,
    method: 'post',
    data: { command }
  })
}

/* ==================== 监控 ==================== */

export function getMemory(id: number | string) {
  return request<RedisMemoryInfo>({
    url: `/redis/${id}/memory`,
    method: 'get'
  })
}

export function getInfo(id: number | string) {
  return request<Record<string, any> | string>({
    url: `/redis/${id}/info`,
    method: 'get'
  })
}

// 设置 Redis Key 过期时间
export function expireKey(id: number | string, key: string, ttl: number) {
  return request({
    url: `/redis/${id}/keys/expire`,
    method: 'post',
    data: { key, ttl }
  })
}
