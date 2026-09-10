import { h } from 'vue'
import { NButton, NIcon, NSpace } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { StatusTag } from '@/components'
import type { StatusType } from '@/components'
import type { DBType } from '@/api/database'
import { DB_TYPE_LABELS } from '@/api/database'

/* ==================== h() 辅助函数 ==================== */

export function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

export function iconBtn(opts: { text: string; icon: any; onClick: () => void; type?: string }) {
  return h(
    NButton,
    {
      size: 'tiny',
      quaternary: true,
      type: (opts.type || 'default') as any,
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

/* ==================== 通用工具 ==================== */

export function rowsToObjects(columns: string[], rows: any[][]): Record<string, any>[] {
  return rows.map((row) => {
    const obj: Record<string, any> = {}
    columns.forEach((col, i) => {
      obj[col] = row[i]
    })
    return obj
  })
}

export function buildDynamicColumns(columns: string[]): DataTableColumns {
  return columns.map((col) => ({ title: col, key: col }))
}

export function formatTime(s: string): string {
  if (!s) return '-'
  return s.replace('T', ' ').substring(0, 19)
}

/* ==================== 类型映射 ==================== */

export const dbTypeStatusMap: Record<DBType, StatusType> = {
  mysql: 'pending',
  postgres: 'running',
  redis: 'error'
}

export function typeTag(type: DBType) {
  return h(StatusTag, {
    status: dbTypeStatusMap[type] || 'stopped',
    text: DB_TYPE_LABELS[type] || type.toUpperCase()
  })
}
