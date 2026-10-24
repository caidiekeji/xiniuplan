import { defineComponent, h, type Component } from 'vue'

/**
 * 云枢面板专属图标集 — 「深夜数据中心」器物语言
 *
 * 与通用图标库的区别：每个图形都从机房器物提炼——机柜、磁带、仪表、集装箱——
 * 笔画统一 1.7 圆头，几何精密不手绘；颜色全部 currentColor，
 * 由 Naive 菜单/按钮的状态色驱动（选中即信号青绿）。
 *
 * 使用：renderIcon(RackIcon) 或 <NIcon><RackIcon /></NIcon>
 */

const base = {
  viewBox: '0 0 24 24',
  fill: 'none' as const,
  stroke: 'currentColor',
  'stroke-width': 1.7,
  'stroke-linecap': 'round' as const,
  'stroke-linejoin': 'round' as const,
  'aria-hidden': 'true',
}

function icon(name: string, render: () => ReturnType<typeof h>): Component {
  return defineComponent({
    name,
    setup() {
      return render
    },
  })
}

/** 机柜（仪表盘）：三层柜体 + 每层一枚指示灯 */
export const RackIcon = icon('RackIcon', () =>
  h('svg', base, [
    h('rect', { x: 4.5, y: 3.5, width: 15, height: 17, rx: 2 }),
    h('line', { x1: 4.5, y1: 9.2, x2: 19.5, y2: 9.2 }),
    h('line', { x1: 4.5, y1: 14.9, x2: 19.5, y2: 14.9 }),
    h('circle', { cx: 8, cy: 6.4, r: 0.9, fill: 'currentColor', stroke: 'none' }),
    h('circle', { cx: 8, cy: 12.05, r: 0.9, fill: 'currentColor', stroke: 'none' }),
    h('circle', { cx: 8, cy: 17.7, r: 0.9, fill: 'currentColor', stroke: 'none' }),
    h('line', { x1: 12.5, y1: 6.4, x2: 16.5, y2: 6.4 }),
    h('line', { x1: 12.5, y1: 12.05, x2: 16.5, y2: 12.05 }),
    h('line', { x1: 12.5, y1: 17.7, x2: 16.5, y2: 17.7 }),
  ])
)

/** 经纬球 + 轨道（网站）：域名解析的轨道感 */
export const GlobeIcon = icon('GlobeIcon', () =>
  h('svg', base, [
    h('circle', { cx: 12, cy: 12, r: 8.5 }),
    h('ellipse', { cx: 12, cy: 12, rx: 3.8, ry: 8.5 }),
    h('line', { x1: 3.5, y1: 12, x2: 20.5, y2: 12 }),
    // 轨道卫星点：网站可被"从任意方位访问"
    h('circle', { cx: 19.4, cy: 6.8, r: 1.1, fill: 'currentColor', stroke: 'none' }),
  ])
)

/** 数据柜（数据库）：立式圆柱带刻度线，机柜同款 LED */
export const DbIcon = icon('DbIcon', () =>
  h('svg', base, [
    h('ellipse', { cx: 12, cy: 5.5, rx: 7.5, ry: 2.6 }),
    h('path', { d: 'M4.5 5.5 V18.5 C4.5 19.9 7.9 21.1 12 21.1 C16.1 21.1 19.5 19.9 19.5 18.5 V5.5' }),
    h('path', { d: 'M4.5 12 C4.5 13.4 7.9 14.6 12 14.6 C16.1 14.6 19.5 13.4 19.5 12' }),
    h('circle', { cx: 16.2, cy: 9.4, r: 0.9, fill: 'currentColor', stroke: 'none' }),
  ])
)

/** 集装箱（容器）：立方线框 + 顶部吊点 */
export const CubeIcon = icon('CubeIcon', () =>
  h('svg', base, [
    h('path', { d: 'M12 3 L20 7.2 V16.8 L12 21 L4 16.8 V7.2 Z' }),
    h('path', { d: 'M4 7.2 L12 11.4 L20 7.2' }),
    h('line', { x1: 12, y1: 11.4, x2: 12, y2: 21 }),
    h('circle', { cx: 12, cy: 16.2, r: 0.9, fill: 'currentColor', stroke: 'none' }),
  ])
)

/** 终端：CRT 方框 + 提示符光标（扫描线点缀） */
export const TerminalIcon = icon('TerminalIcon', () =>
  h('svg', base, [
    h('rect', { x: 3.5, y: 4.5, width: 17, height: 15, rx: 2 }),
    h('path', { d: 'M7 9.5 L10.2 12 L7 14.5' }),
    h('line', { x1: 12.5, y1: 15, x2: 16.5, y2: 15 }),
    h('circle', { cx: 18.2, cy: 6.9, r: 0.8, fill: 'currentColor', stroke: 'none' }),
  ])
)

/** 仪表（监控）：表盘刻度 + 指针，精密仪器感 */
export const GaugeIcon = icon('GaugeIcon', () =>
  h('svg', base, [
    h('path', { d: 'M4.5 17.5 A 8.5 8.5 0 1 1 19.5 17.5' }),
    // 刻度
    h('line', { x1: 6.4, y1: 9.5, x2: 7.9, y2: 10.4 }),
    h('line', { x1: 12, y1: 6.2, x2: 12, y2: 8 }),
    h('line', { x1: 17.6, y1: 9.5, x2: 16.1, y2: 10.4 }),
    // 指针指向安全区
    h('path', { d: 'M12 14.8 L15.4 11.2' }),
    h('circle', { cx: 12, cy: 15.2, r: 1 }),
  ])
)

/** 磁带（备份）：卷盘 + 走带窗 */
export const TapeIcon = icon('TapeIcon', () =>
  h('svg', base, [
    h('rect', { x: 3.5, y: 6, width: 17, height: 12, rx: 2 }),
    h('circle', { cx: 8.4, cy: 12, r: 2.4 }),
    h('circle', { cx: 15.6, cy: 12, r: 2.4 }),
    h('line', { x1: 10.8, y1: 12, x2: 13.2, y2: 12 }),
    h('circle', { cx: 8.4, cy: 12, r: 0.7, fill: 'currentColor', stroke: 'none' }),
    h('circle', { cx: 15.6, cy: 12, r: 0.7, fill: 'currentColor', stroke: 'none' }),
  ])
)

/** 烽火墙（防火墙）：盾内网格砖墙 */
export const FirewallIcon = icon('FirewallIcon', () =>
  h('svg', base, [
    h('path', { d: 'M12 3 L19.5 5.8 V11.5 C19.5 16.4 16.4 19.7 12 21 C7.6 19.7 4.5 16.4 4.5 11.5 V5.8 Z' }),
    h('line', { x1: 4.6, y1: 9.4, x2: 19.4, y2: 9.4 }),
    h('line', { x1: 4.7, y1: 13.6, x2: 19.1, y2: 13.6 }),
    h('line', { x1: 12, y1: 5.6, x2: 12, y2: 9.4 }),
    h('line', { x1: 8.2, y1: 9.4, x2: 8.2, y2: 13.6 }),
    h('line', { x1: 15.8, y1: 9.4, x2: 15.8, y2: 13.6 }),
    h('line', { x1: 12, y1: 13.6, x2: 12, y2: 20.8 }),
  ])
)

/** 证书（SSL）：印章式圆环 + 锁舌 */
export const CertIcon = icon('CertIcon', () =>
  h('svg', base, [
    h('circle', { cx: 12, cy: 10, r: 6.5 }),
    h('path', { d: 'M9.2 19.2 L8.6 21.5 L12 20.2 L15.4 21.5 L14.8 19.2' }),
    // 锁体
    h('rect', { x: 9.6, y: 8, width: 4.8, height: 3.6, rx: 0.8 }),
    h('path', { d: 'M10.6 8 V7 A1.4 1.4 0 0 1 13.4 7 V8' }),
  ])
)

/** 计时器（计划任务）：工位翻页钟，铆钉刻度 */
export const CronIcon = icon('CronIcon', () =>
  h('svg', base, [
    h('circle', { cx: 12, cy: 13, r: 8 }),
    h('path', { d: 'M12 8.6 V13 L15 15' }),
    h('line', { x1: 9.5, y1: 2.8, x2: 14.5, y2: 2.8 }),
    h('circle', { cx: 12, cy: 13, r: 0.9, fill: 'currentColor', stroke: 'none' }),
  ])
)
