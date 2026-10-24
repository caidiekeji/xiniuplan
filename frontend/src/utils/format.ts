/**
 * 通用格式化工具函数
 */

/**
 * 将字节数格式化为人类可读的字符串（KB/MB/GB/TB）
 * @param bytes 字节数
 * @returns 格式化后的字符串，如 "1.5 MB"
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0 || bytes == null) return '0 B'
  if (bytes < 0) return `-${formatBytes(Math.abs(bytes))}`
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  const idx = Math.min(i, sizes.length - 1)
  const value = bytes / Math.pow(k, idx)
  return `${value.toFixed(value >= 100 || idx === 0 ? 0 : 1)} ${sizes[idx]}`
}

/**
 * 格式化日期为 YYYY-MM-DD HH:mm:ss
 * @param date Date 对象或可被 Date 构造的字符串/时间戳
 * @returns 格式化后的日期字符串
 */
export function formatDate(date: Date | string | number): string {
  const d = date instanceof Date ? date : new Date(date)
  if (isNaN(d.getTime())) return '-'
  const pad = (n: number) => String(n).padStart(2, '0')
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  )
}

/**
 * 截断字符串到指定长度，超出部分用省略号表示
 * @param str 原始字符串
 * @param maxLen 最大长度（包含省略号）
 * @returns 截断后的字符串
 */
export function truncate(str: string, maxLen: number): string {
  if (str == null) return ''
  if (maxLen <= 0) return ''
  if (str.length <= maxLen) return str
  if (maxLen === 1) return '…'
  return str.slice(0, maxLen - 1) + '…'
}

/**
 * 将 ISO 8601 时间字符串格式化为 "YYYY-MM-DD HH:mm:ss"（无效输入原样返回，空值返回 "-"）
 */
export function formatISOTime(s: string | null | undefined): string {
  if (!s) return '-'
  return s.replace('T', ' ').substring(0, 19)
}
