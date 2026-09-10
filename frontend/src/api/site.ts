import { request } from './request'

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

export function getSites(params: SiteListParams) {
  return request<SiteListData>({
    url: '/sites',
    method: 'get',
    params
  })
}

export function createSite(data: CreateSiteData) {
  return request<Site>({
    url: '/sites',
    method: 'post',
    data
  })
}

export function getSite(id: number) {
  return request<Site>({
    url: `/sites/${id}`,
    method: 'get'
  })
}

export function updateSite(id: number, data: UpdateSiteData) {
  return request<Site>({
    url: `/sites/${id}`,
    method: 'put',
    data
  })
}

export function deleteSite(id: number, deleteFiles = false) {
  return request({
    url: `/sites/${id}`,
    method: 'delete',
    params: { deleteFiles }
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
