import { h } from 'vue'
import { NButton, NIcon, NSpace, NTag } from 'naive-ui'
import { StatusTag } from '@/components'
import type { StatusType } from '@/components'
import type { AppCategory, AppStatus, AppSourceType } from '@/api/appstore'

/* ==================== h() 辅助函数 ==================== */

export function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

export function iconBtn(opts: {
  text: string
  icon: any
  onClick: () => void
  type?: string
  disabled?: boolean
  loading?: boolean
}) {
  return h(
    NButton,
    {
      size: 'tiny',
      quaternary: true,
      type: (opts.type || 'default') as any,
      disabled: opts.disabled,
      loading: opts.loading,
      onClick: opts.onClick
    },
    {
      icon: () => h(NIcon, null, { default: () => h(opts.icon) }),
      default: () => opts.text
    }
  )
}

export function btnGroup(children: any[]) {
  return h(NSpace, { size: 4 }, { default: () => children })
}

/* ==================== 工具函数 ==================== */

export function formatTime(s: string): string {
  if (!s) return '-'
  return s.replace('T', ' ').substring(0, 19)
}

/** 根据 app key 生成确定性颜色 */
const iconColors = [
  '#1BA784', '#2DD4A8', '#60A5FA', '#E85D4D', '#A78BFA',
  '#F472B6', '#34D399', '#4AC7A5', '#38BDF8', '#FB7185'
]

export function getIconColor(key: string): string {
  let hash = 0
  for (let i = 0; i < key.length; i++) {
    hash = key.charCodeAt(i) + ((hash << 5) - hash)
  }
  return iconColors[Math.abs(hash) % iconColors.length]
}

export function categoryLabel(cat: AppCategory): string {
  const map: Record<AppCategory, string> = {
    website: '建站',
    devtools: '开发工具',
    database: '数据库',
    middleware: '中间件',
    monitor: '监控',
    office: '办公',
    storage: '存储',
    network: '网络'
  }
  return map[cat] || cat
}

export function categoryTagType(cat: AppCategory): string {
  const map: Record<AppCategory, string> = {
    website: 'info',
    devtools: 'success',
    database: 'warning',
    middleware: 'default',
    monitor: 'error',
    office: 'info',
    storage: 'success',
    network: 'default'
  }
  return map[cat] || 'default'
}

export function mapAppStatus(status: AppStatus): StatusType {
  switch (status) {
    case 'installed':
      return 'running'
    case 'installing':
      return 'pending'
    case 'failed':
      return 'error'
    default:
      return 'stopped'
  }
}

export function appStatusText(status: AppStatus): string {
  const map: Record<AppStatus, string> = {
    installed: '已安装',
    installing: '安装中',
    failed: '安装失败',
    not_installed: '未安装'
  }
  return map[status] || status
}

/* ==================== 源相关工具函数 ==================== */

export function sourceTypeLabel(type: AppSourceType): string {
  const map: Record<AppSourceType, string> = {
    official: '官方',
    community: '社区',
    private: '私有'
  }
  return map[type] || type
}

export function sourceTypeTagType(type: AppSourceType): string {
  const map: Record<AppSourceType, string> = {
    official: 'warning',
    community: 'success',
    private: 'default'
  }
  return map[type] || 'default'
}

export function appSourceTag(app: { source_id?: string }, sourceMap: Record<string, { name: string; type: AppSourceType }>) {
  if (!app.source_id) return null
  const src = sourceMap[app.source_id]
  if (!src) return null
  return h(
    NTag,
    {
      size: 'tiny',
      bordered: false,
      type: sourceTypeTagType(src.type) as any,
      class: 'app-card__source'
    },
    { default: () => src.name }
  )
}
