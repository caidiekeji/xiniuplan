import { request, type ApiResponse } from './request'

export type SiteType = 'static' | 'runtime' | 'reverse_proxy'
export type SiteStatus = 'running' | 'stopped' | 'error'

export interface Site {
  id: number
  domain: string
  aliases: string[]
  type: SiteType
  runtime?: string
  proxyTarget?: string
  rootDir: string
  status: SiteStatus
  ssl: boolean
  remark?: string
  createdAt: string
}

export interface SiteListParams {
  page: number
  pageSize: number
  domain?: string
  type?: SiteType
  status?: SiteStatus
}

export interface SiteListData {
  items: Site[]
  total: number
}

export interface CreateSiteData {
  domain: string
  aliases: string[]
  type: SiteType
  runtime?: string
  proxyTarget?: string
  rootDir: string
  remark?: string
}

export type UpdateSiteData = Partial<CreateSiteData>

/* ============================================================
 * 前后端字段契约适配
 * 后端 DTO（backend/internal/dto/site.go）使用 snake_case，
 * 且 ssl 为对象；前端页面按 camelCase + ssl:boolean 使用。
 * 此处集中做双向映射，页面代码无需改动，也避免再次写错字段名。
 * ============================================================ */

/** 后端原始站点对象（snake_case） */
interface RawSite {
  id: number
  primary_domain?: string
  domain?: string
  domains?: string[]
  aliases?: string[]
  type: SiteType
  runtime?: string
  root_dir?: string
  rootDir?: string
  proxy_target?: string
  proxyTarget?: string
  nginx_conf?: string
  status: SiteStatus
  container_id?: string
  remark?: string
  ssl?: boolean | { enabled?: boolean; expire_days?: number; days_left?: number }
  waf_enabled?: boolean
  cc_rpm?: number
  ua_blacklist?: string
  created_at?: string
  createdAt?: string
  updated_at?: string
}

/** 后端原始站点 → 前端 Site */
function toSite(raw: RawSite): Site {
  const ssl = raw.ssl
  return {
    id: raw.id,
    domain: raw.primary_domain ?? raw.domain ?? '',
    aliases: raw.domains ?? raw.aliases ?? [],
    type: raw.type,
    runtime: raw.runtime,
    proxyTarget: raw.proxy_target ?? raw.proxyTarget,
    rootDir: raw.root_dir ?? raw.rootDir ?? '',
    status: raw.status,
    ssl: typeof ssl === 'object' ? !!ssl?.enabled : !!ssl,
    remark: raw.remark,
    createdAt: raw.created_at ?? raw.createdAt ?? ''
  }
}

/** 前端表单 → 后端请求体（snake_case） */
function toCreatePayload(data: CreateSiteData) {
  return {
    primary_domain: data.domain,
    domains: data.aliases ?? [],
    type: data.type,
    runtime: data.runtime,
    root_dir: data.rootDir,
    proxy_target: data.proxyTarget,
    remark: data.remark
  }
}

export async function getSites(params: SiteListParams): Promise<ApiResponse<SiteListData>> {
  const res = await request<{ items: RawSite[]; total: number }>({
    url: '/sites',
    method: 'get',
    params
  })
  const items = (res.data.items || []).map(toSite)
  return { ...res, data: { items, total: res.data.total } }
}

export async function createSite(data: CreateSiteData): Promise<ApiResponse<Site>> {
  const res = await request<RawSite>({
    url: '/sites',
    method: 'post',
    data: toCreatePayload(data)
  })
  return { ...res, data: toSite(res.data) }
}

export async function getSite(id: number): Promise<ApiResponse<Site>> {
  const res = await request<RawSite>({
    url: `/sites/${id}`,
    method: 'get'
  })
  return { ...res, data: toSite(res.data) }
}

export async function updateSite(id: number, data: UpdateSiteData): Promise<ApiResponse<Site>> {
  const res = await request<RawSite>({
    url: `/sites/${id}`,
    method: 'put',
    data: toCreatePayload({ domain: '', aliases: [], type: 'static', rootDir: '', ...data })
  })
  return { ...res, data: toSite(res.data) }
}

export function deleteSite(id: number, deleteFiles = false) {
  return request({
    url: `/sites/${id}`,
    method: 'delete',
    // 后端读的是 snake_case 的 delete_files（backend/internal/api/site.go）
    params: { delete_files: deleteFiles }
  })
}

export function startSite(id: number) {
  return request({
    url: `/sites/${id}/start`,
    method: 'post'
  })
}

export function stopSite(id: number) {
  return request({
    url: `/sites/${id}/stop`,
    method: 'post'
  })
}

export function restartSite(id: number) {
  return request({
    url: `/sites/${id}/restart`,
    method: 'post'
  })
}

export function getNginxConfig(id: number) {
  return request<{ content: string }>({
    url: `/sites/${id}/config`,
    method: 'get'
  })
}

export function updateNginxConfig(id: number, content: string) {
  return request({
    url: `/sites/${id}/config`,
    method: 'put',
    data: { content }
  })
}

export function getSiteLogs(id: number, type: 'access' | 'error' = 'access', lines = 200) {
  return request<{ content: string }>({
    url: `/sites/${id}/logs`,
    method: 'get',
    params: { type, lines }
  })
}
