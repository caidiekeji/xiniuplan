import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface PluginInfo {
  name: string
  version: string
  description: string
  author: string
  icon: string
}

export interface PluginItem extends PluginInfo {
  enabled: boolean
  running: boolean
}

export interface PluginListData {
  items: PluginItem[]
  total: number
}

export interface PluginConfigData {
  name: string
  enabled: boolean
  config: Record<string, any>
  raw: string
}

/* ==================== 插件管理 API ==================== */

/** 获取已安装插件列表（含状态） */
export function getPlugins() {
  return request<PluginListData>({
    url: '/plugins',
    method: 'get'
  })
}

/** 启用插件 */
export function enablePlugin(name: string) {
  return request<{ name: string; enabled: boolean; running: boolean }>({
    url: `/plugins/${name}/enable`,
    method: 'post'
  })
}

/** 禁用插件 */
export function disablePlugin(name: string) {
  return request<{ name: string; enabled: boolean; running: boolean }>({
    url: `/plugins/${name}/disable`,
    method: 'post'
  })
}

/** 获取插件配置 */
export function getPluginConfig(name: string) {
  return request<PluginConfigData>({
    url: `/plugins/${name}/config`,
    method: 'get'
  })
}

/** 更新插件配置 */
export function updatePluginConfig(name: string, config: Record<string, any>) {
  return request<{ name: string; config: Record<string, any> }>({
    url: `/plugins/${name}/config`,
    method: 'put',
    data: { config }
  })
}
