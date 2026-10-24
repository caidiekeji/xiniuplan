import { ref, onMounted } from 'vue'

/** hex 颜色转 rgba，用于 ECharts/xterm 等需要具体色值的 JS 配置 */
export function hexToRgba(hex: string, alpha: number): string {
  const h = hex.replace('#', '')
  const r = parseInt(h.substring(0, 2), 16)
  const g = parseInt(h.substring(2, 4), 16)
  const b = parseInt(h.substring(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

/**
 * 读取字号 token 并换算为 px 数字，供 ECharts 等只接受数字字号的 Canvas 渲染使用。
 * 字号唯一来源仍是 tokens.css 的 --fs-*，此处仅做单位换算，不产生新的字号档。
 * 例：fsPx('--fs-caption', 12) —— root 16px 时 0.75rem 换算为 12。
 * 注意：在组件挂载后（CSS 已加载）调用；模块顶层调用可能读到空值。
 */
export function fsPx(token: string, fallback: number): number {
  const raw = getComputedStyle(document.documentElement).getPropertyValue(token).trim()
  if (!raw) return fallback
  const rem = raw.match(/^([\d.]+)rem$/)
  if (rem) {
    const rootPx = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16
    return Math.round(parseFloat(rem[1]) * rootPx)
  }
  const px = raw.match(/^([\d.]+)px$/)
  if (px) return parseFloat(px[1])
  return fallback
}

/**
 * 从 CSS 变量读取设计 token，供 JS 端（Naive UI themeOverrides、ECharts、xterm 等）使用。
 * 所有颜色/尺寸的唯一来源是 tokens.css，此处只做读取，不硬编码。
 */
export function useThemeTokens() {
  const tokens = ref<Record<string, string>>({})
  /** 图表/头像 10 色板（--chart-palette-1..10），数组型 token 单独存放 */
  const chartPalette = ref<string[]>([])

  function readToken(name: string): string {
    const val = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
    return val
  }

  function refresh() {
    tokens.value = {
      acc: readToken('--acc'),
      accHover: readToken('--acc-hover'),
      accPressed: readToken('--acc') + 'dd',
      acc2: readToken('--acc2'),
      ok: readToken('--ok'),
      warn: readToken('--warn'),
      err: readToken('--err'),
      info: readToken('--info'),
      ink: readToken('--ink'),
      sub: readToken('--sub'),
      faint: readToken('--faint'),
      textQuaternary: readToken('--text-quaternary'),
      bg: readToken('--bg'),
      bg2: readToken('--bg2'),
      surface: readToken('--surface'),
      surface2: readToken('--surface2'),
      line: readToken('--line'),
      lineStrong: readToken('--line-strong'),
      radiusMd: readToken('--radius-md'),
      fontSans: readToken('--font-sans'),
      fontMono: readToken('--font-mono'),
      fsBody: readToken('--fs-body'),
      fsCaption: readToken('--fs-caption'),
      btnBg: readToken('--btn-bg'),
      btnBgHover: readToken('--btn-bg-hover'),
      btnFg: readToken('--btn-fg'),
      termMagenta: readToken('--term-magenta'),
      termCyan: readToken('--term-cyan'),
      termFontSize: readToken('--term-font-size')
    }
    chartPalette.value = Array.from({ length: 10 }, (_, i) => readToken(`--chart-palette-${i + 1}`))
  }

  onMounted(() => {
    refresh()
  })

  return { tokens, chartPalette, refresh, readToken }
}
