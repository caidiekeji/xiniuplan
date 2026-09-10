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
 * 从 CSS 变量读取设计 token，供 JS 端（Naive UI themeOverrides、ECharts、xterm 等）使用。
 * 所有颜色/尺寸的唯一来源是 tokens.css，此处只做读取，不硬编码。
 */
export function useThemeTokens() {
  const tokens = ref<Record<string, string>>({})

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
      termMagenta: readToken('--term-magenta'),
      termCyan: readToken('--term-cyan'),
      termFontSize: readToken('--term-font-size')
    }
  }

  onMounted(() => {
    refresh()
  })

  return { tokens, refresh, readToken }
}
