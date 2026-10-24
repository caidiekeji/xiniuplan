import { request, type ApiResponse } from './request'
import { getSites } from './site'

export type SSLProvider = 'letsencrypt' | 'manual'
export type SSLStatus = 'valid' | 'expiring' | 'expired' | 'pending'

export interface SSLInfo {
  id?: number
  siteId: number
  domain: string
  provider: SSLProvider
  email?: string
  issuer?: string
  expireAt?: string
  status: SSLStatus
  enabled?: boolean
}

export interface ApplySSLData {
  provider: SSLProvider
  email?: string
  cert?: string
  key?: string
  verification_method?: 'http' | 'dns'
  dns_account_id?: number
}

/** 获取所有站点的 SSL 信息（从站点列表提取已启用 SSL 的站点） */
export async function getSSLList(): Promise<ApiResponse<SSLInfo[]>> {
  try {
    const res = await getSites({ page: 1, pageSize: 100 })
    const sites = res.data?.items || []
    const list: SSLInfo[] = sites
      .filter((s: any) => s.ssl_enabled || s.ssl)
      .map((s: any): SSLInfo => ({
        // id 必须是真实 siteId —— 此前用 idx+1 伪 ID，导致申请/续期/删除
        // 打到别的站点上（2026-10-23 闭环审计）
        id: s.id,
        siteId: s.id,
        domain: s.primary_domain || s.domain,
        provider: (s.ssl_provider as SSLProvider) || 'manual',
        expireAt: s.ssl_expire_at,
        enabled: s.ssl_enabled ?? s.ssl ?? false,
        status: computeSSLStatus(s.ssl_expire_at, s.ssl_enabled ?? s.ssl)
      }))
    return { code: 200, message: 'success', data: list }
  } catch {
    return { code: 200, message: 'success', data: [] }
  }
}

function computeSSLStatus(expireAt?: string, enabled?: boolean): SSLStatus {
  if (!enabled) return 'pending'
  if (!expireAt) return 'pending'
  const expire = new Date(expireAt).getTime()
  const now = Date.now()
  const days = (expire - now) / (1000 * 60 * 60 * 24)
  if (days < 0) return 'expired'
  if (days < 30) return 'expiring'
  return 'valid'
}

/** 为指定站点申请/上传 SSL 证书 */
export function applySSL(siteId: number, data: ApplySSLData) {
  return request({
    url: `/sites/${siteId}/ssl/apply`,
    method: 'post',
    data
  })
}

/** 获取指定站点的 SSL 信息 */
export function getSSLInfo(siteId: number) {
  return request<SSLInfo>({
    url: `/sites/${siteId}/ssl`,
    method: 'get'
  })
}

/**
 * 续期 SSL（重新申请）。
 * email 是后端必填项——此前不传，续期按钮必然 400（2026-10-23 闭环审计）。
 */
export function renewSSL(siteId: number, email: string) {
  return request<{ site?: unknown; status?: string }>({
    url: `/sites/${siteId}/ssl/apply`,
    method: 'post',
    data: { provider: 'letsencrypt', email }
  })
}

/** 删除 SSL（禁用并清除证书）
 *  此前用「申请空证书」变相删除，后端要求 cert/key 非空必然 400；
 *  现由后端提供 DELETE /sites/:id/ssl 真正执行禁用与文件清理。
 */
export function deleteSSL(siteId: number) {
  return request({
    url: `/sites/${siteId}/ssl`,
    method: 'delete'
  })
}
