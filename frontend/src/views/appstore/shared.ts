import { h } from 'vue'
import { NButton, NIcon, NSpace, NTag } from 'naive-ui'
 '@/components'
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

export { formatISOTime as formatTime } from '@/utils/format'

/** 根据 app key 生成确定性色板序号（1-10），色值统一由 --chart-palette-* token 提供 */
export function getIconColor(key: string): string {
  let hash = 0
  for (let i = 0; i < key.length; i++) {
    hash = key.charCodeAt(i) + ((hash << 5) - hash)
  }
  return `var(--chart-palette-${(Math.abs(hash) % 10) + 1})`
}

export function categoryLabel(t: (k: string, prms?: any) => string, cat: AppCategory): string {
  return t('appstore.categories.' + cat) || cat
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

export function appStatusText(t: (k: string, prms?: any) => string, status: AppStatus): string {
  return t('appstore.statuses.' + status) || status
}

/* ==================== 源相关工具函数 ==================== */

export function sourceTypeLabel(t: (k: string, prms?: any) => string, type: AppSourceType): string {
  return t('appstore.sourceTypes.' + type) || type
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
