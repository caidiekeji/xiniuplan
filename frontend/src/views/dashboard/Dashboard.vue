<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NGrid, NGi, NSpace, NButton } from 'naive-ui'
// echarts 按需引入：只注册实际用到的图表和组件，减小包体积
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
import { StatCard, PageCard, StatusTag } from '@/components'
import { getOverview, getMetrics, getAuditLogs, normalizeOverview, type OverviewData, type AuditLogItem } from '@/api/dashboard'
import { useThemeTokens, fsPx } from '@/composables/useThemeTokens'

const { t, locale } = useI18n()
const { tokens } = useThemeTokens()

/** hex 颜色转 rgba，用于 ECharts 渐变 */
function hexToRgba(hex: string, alpha: number): string {
  const h = hex.replace('#', '')
  const r = parseInt(h.substring(0, 2), 16)
  const g = parseInt(h.substring(2, 4), 16)
  const b = parseInt(h.substring(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

const loading = ref(true)
const overview = ref<OverviewData | null>(null)
const auditLogs = ref<AuditLogItem[]>([])
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null

// 图表文字摘要：读屏用户看不到 Canvas，用 aria-label 承载最新值（审计 D10①）
const chartAriaLabel = ref('')
function updateChartAria(points: { cpu: number; memory: number }[]) {
  if (points.length === 0) {
    chartAriaLabel.value = t('dashboard.chartAriaEmpty')
    return
  }
  const last = points[points.length - 1]
  const rangeLabel = rangeOptions.value.find((o) => o.value === metricRange.value)?.label || metricRange.value
  chartAriaLabel.value = t('dashboard.chartAria', {
    cpu: last.cpu.toFixed(1),
    mem: last.memory.toFixed(1),
    range: rangeLabel
  })
}

const metricRange = ref('24h')
const rangeOptions = computed(() => [
  { label: t('dashboard.range1h'), value: '1h' },
  { label: t('dashboard.range6h'), value: '6h' },
  { label: t('dashboard.range24h'), value: '24h' },
  { label: t('dashboard.range7d'), value: '7d' }
])

const cpuPercent = computed(() => (overview.value?.cpu.usage ?? 0).toFixed(1))
const memPercent = computed(() => (overview.value?.memory.usage ?? 0).toFixed(1))
const diskPercent = computed(() => (overview.value?.disk.usage ?? 0).toFixed(1))


function formatSpeed(kbps: number | undefined | null): string {
  if (kbps == null || isNaN(kbps)) return '0 KB/s'
  if (kbps >= 1024) return (kbps / 1024).toFixed(1) + ' MB/s'
  return kbps.toFixed(0) + ' KB/s'
}

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
}

let lastPoints: { timestamp: number; cpu: number; memory: number }[] | null = null

// 各范围的聚合桶秒数（与后端 GetMetrics 保持一致），用于识别采集断档
const BUCKET_SEC: Record<string, number> = { '1h': 60, '6h': 300, '24h': 3600, '7d': 3600 }

function updateChart(data?: { timestamp: number; cpu: number; memory: number }[]) {
  if (!chartInstance) return
  if (data) lastPoints = data

  // 空态与有数据走同一个完整 option 结构（series 始终带 type/lineStyle），
  // 避免 merge setOption 时 series 结构差异导致 ECharts 报 coordinateSystem 错误，
  // 也避免空态 title 在数据回来后残留
  const points = data || lastPoints || []
  updateChartAria(points)

  // 后端 timestamp 为 Unix 秒；时间轴按真实时间间隔排布，
  // 采集断档（面板停机/重启）处用 null 断开，避免 category 轴把空洞压缩成相邻点造成趋势失真
  const fmtTime = (ms: number) => {
    const d = new Date(ms)
    const mmdd = `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const hhmm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    if (metricRange.value === '7d' || metricRange.value === '24h') return `${mmdd} ${hhmm}`
    return hhmm
  }
  // 相邻点间隔超过 2 个聚合桶即视为断档，插入 null 让折线断开
  const gapLimit = (BUCKET_SEC[metricRange.value] || 3600) * 2 * 1000
  const cpuData: [number, number | null][] = []
  const memData: [number, number | null][] = []
  points.forEach((p, i) => {
    const ms = p.timestamp * 1000
    if (i > 0 && ms - points[i - 1].timestamp * 1000 > gapLimit) {
      cpuData.push([ms, null])
      memData.push([ms, null])
    }
    cpuData.push([ms, p.cpu])
    memData.push([ms, p.memory])
  })

  const tk = tokens.value
  const textColor = tk.sub || '#9aa5b8'
  const lineColor = tk.line || '#2a3348'
  // CPU 用主撞色青绿，内存用信息蓝 —— 两线必须强区分，同色系双线在趋势图上不可读
  const cpuColor = tk.acc || '#1ba784'
  const memColor = tk.info || '#60a5fa'
  const faintColor = tk.faint || '#8b95a8'
  // 图表字号统一走 token（审计 D01：12/13px 散点 → --fs-caption 一档）
  const fsCaption = fsPx('--fs-caption', 12)

  const cpuLabel = t('dashboard.cpu')
  const memLabel = t('dashboard.memory')
  const isEmpty = points.length === 0

  chartInstance.setOption({
    title: isEmpty
      ? {
          text: t('common.noData'),
          left: 'center',
          top: 'middle',
          textStyle: { color: faintColor, fontSize: fsCaption, fontWeight: 400 }
        }
      : { show: false },
    grid: { top: 40, right: 20, bottom: 30, left: 50 },
    tooltip: {
      trigger: 'axis',
      backgroundColor: tk.surface || '#1a2030',
      borderColor: tk.line || '#2a3348',
      textStyle: { color: tk.ink || '#e8ecf2', fontSize: fsCaption },
      formatter: (params: any) => {
        let html = `<div style="font-weight:600;margin-bottom:var(--space-1)">${fmtTime(params[0].value[0])}</div>`
        params.forEach((p: any) => {
          const v = p.value[1]
          html += `<div style="display:flex;align-items:center;gap:var(--space-2)">
            <span style="width:8px;height:8px;border-radius:50%;background:${p.color}"></span>
            ${p.seriesName}: <span style="font-family:monospace;font-weight:600">${v == null ? '—' : v.toFixed(1) + '%'}</span>
          </div>`
        })
        return html
      }
    },
    legend: {
      data: [cpuLabel, memLabel],
      top: 8,
      right: 20,
      textStyle: { color: textColor, fontSize: fsCaption },
      itemWidth: 12,
      itemHeight: 8
    },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { color: lineColor } },
      axisLabel: { color: textColor, fontSize: fsCaption, formatter: (val: number) => fmtTime(val) },
      axisTick: { show: false },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: textColor, fontSize: fsCaption, formatter: '{value}%' },
      splitLine: { lineStyle: { color: lineColor, type: 'dashed' } }
    },
    series: [
      {
        name: cpuLabel,
        type: 'line',
        data: cpuData,
        connectNulls: false,
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 2, color: cpuColor },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: hexToRgba(cpuColor, 0.25) },
            { offset: 1, color: hexToRgba(cpuColor, 0.02) }
          ])
        }
      },
      {
        name: memLabel,
        type: 'line',
        data: memData,
        connectNulls: false,
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 2, color: memColor },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: hexToRgba(memColor, 0.25) },
            { offset: 1, color: hexToRgba(memColor, 0.02) }
          ])
        }
      }
    ]
  })
}

async function loadData() {
  loading.value = true
  try {
    const [ovRes, metricsRes, logsRes] = await Promise.all([
      getOverview().catch(() => null),
      getMetrics(metricRange.value).catch(() => null),
      getAuditLogs(1, 10).catch(() => null)
    ])

    if (ovRes?.data) {
      overview.value = normalizeOverview(ovRes.data)
    } else {
      overview.value = null
    }

    if (metricsRes?.data?.points && metricsRes.data.points.length > 0) {
      updateChart(metricsRes.data.points)
    } else {
      updateChart()
    }

    if (logsRes?.data?.items) {
      auditLogs.value = logsRes.data.items
    } else {
      auditLogs.value = []
    }
  } finally {
    loading.value = false
  }
}

function handleRangeChange(range: string) {
  metricRange.value = range
  getMetrics(range)
    .then((res) => {
      if (res.data?.points && res.data.points.length > 0) updateChart(res.data.points)
      else updateChart()
    })
    .catch(() => updateChart())
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  return `${days}${t('dashboard.days')} ${hours}${t('dashboard.hours')}`
}

function handleResize() {
  chartInstance?.resize()
}

// 容器尺寸监听：侧栏折叠/展开只改容器宽度不触发 window resize，
// 用 ResizeObserver 保证图表始终跟随容器尺寸
let chartResizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  initChart()
  await loadData()
  window.addEventListener('resize', handleResize)
  if (chartRef.value && 'ResizeObserver' in window) {
    chartResizeObserver = new ResizeObserver(() => handleResize())
    chartResizeObserver.observe(chartRef.value)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartResizeObserver?.disconnect()
  chartResizeObserver = null
  chartInstance?.dispose()
  chartInstance = null
})

// 主题切换时重新读取 token 并刷新图表
watch(
  () => tokens.value.acc,
  () => {
    nextTick(() => {
      if (chartInstance) {
        updateChart()
      }
    })
  }
)

// 语言切换时刷新图表图例
watch(locale, () => {
  nextTick(() => {
    if (chartInstance) {
      updateChart()
    }
  })
})
</script>

<template>
  <div class="dashboard">
    <!-- 系统信息条 -->
    <div class="dashboard__sysbar" v-if="overview">
      <div class="sysbar-item">
        <span class="sysbar-label">{{ t('dashboard.hostname') }}</span>
        <span class="sysbar-value mono">{{ overview.hostname }}</span>
      </div>
      <div class="sysbar-divider" />
      <div class="sysbar-item">
        <span class="sysbar-label">{{ t('dashboard.system') }}</span>
        <span class="sysbar-value">{{ overview.os }}</span>
      </div>
      <div class="sysbar-divider" />
      <div class="sysbar-item">
        <span class="sysbar-label">{{ t('dashboard.runtime') }}</span>
        <span class="sysbar-value mono">{{ formatUptime(overview.uptime) }}</span>
      </div>
      <div class="sysbar-divider" />
      <div class="sysbar-item">
        <span class="sysbar-label">{{ t('dashboard.panel') }}</span>
        <span class="sysbar-value mono">{{ overview.panelVersion }}</span>
      </div>
      <div class="sysbar-right">
        <span class="led led--ok" />
        <span>{{ t('dashboard.allServicesNormal') }}</span>
      </div>
    </div>

    <!-- StatCard 四连排 -->
    <NGrid cols="1 s:2 l:4" :x-gap="16" :y-gap="16" responsive="screen" class="dashboard__stats">
      <NGi>
        <StatCard
          :title="t('dashboard.cpuUsage')"
          :value="cpuPercent"
          unit="%"
          :subtitle="`${overview?.cpu.cores || 4} ${t('dashboard.cores')}`"
          :loading="loading"
          accent="acc"
        />
      </NGi>
      <NGi>
        <StatCard
          :title="t('dashboard.memoryUsage')"
          :value="memPercent"
          unit="%"
          :subtitle="`${overview ? overview.memory.usedGB.toFixed(1) : '—'} / ${overview ? overview.memory.totalGB.toFixed(0) : 8} GB`"
          :loading="loading"
          accent="acc2"
        />
      </NGi>
      <NGi>
        <StatCard
          :title="t('dashboard.diskUsage')"
          :value="diskPercent"
          unit="%"
          :subtitle="`${overview ? overview.disk.usedGB.toFixed(1) : '—'} / ${overview ? overview.disk.totalGB.toFixed(1) : 40} GB`"
          :loading="loading"
          accent="warn"
        />
      </NGi>
      <NGi>
        <StatCard
          :title="t('dashboard.networkTraffic')"
          :value="overview ? formatSpeed(overview.network.downloadKbps) : '—'"
          :subtitle="t('dashboard.downloadUpload')"
          :loading="loading"
          accent="info"
        >
          <template #subtitle>
            <span class="mono" style="color: var(--acc2)">↓ {{ overview ? formatSpeed(overview.network.downloadKbps) : '—' }}</span>
            <span class="dashboard__net-sep">·</span>
            <span class="mono" style="color: var(--acc)">↑ {{ overview ? formatSpeed(overview.network.uploadKbps) : '—' }}</span>
          </template>
        </StatCard>
      </NGi>
    </NGrid>

    <!-- 图表 + 日志（最近操作放在资源趋势下面） -->
    <NGrid cols="1 l:3" :x-gap="16" :y-gap="16" responsive="screen" class="dashboard__main">
      <NGi :span="3" class="dashboard__chart-col">
        <PageCard :title="t('dashboard.resourceTrend')" :subtitle="t('dashboard.resourceTrendSubtitle')">
          <template #extra>
            <NSpace size="small" align="center">
              <NButton
                v-for="opt in rangeOptions"
                :key="opt.value"
                size="tiny"
                :type="metricRange === opt.value ? 'primary' : 'default'"
                @click="handleRangeChange(opt.value)"
              >
                {{ opt.label }}
              </NButton>
            </NSpace>
          </template>
          <div ref="chartRef" class="dashboard__chart" role="img" :aria-label="chartAriaLabel" />
        </PageCard>
      </NGi>
      <NGi :span="3" class="dashboard__audit-col">
        <PageCard :title="t('dashboard.recentOperations')" :subtitle="t('dashboard.recentOperationsSubtitle')">
          <div class="audit-list audit-list--horizontal">
            <div v-for="log in auditLogs" :key="log.id" class="audit-item">
              <div class="audit-item__top">
                <span class="audit-item__action">{{ log.action }}</span>
                <StatusTag :status="log.result === 'success' ? 'running' : 'error'" :text="log.result === 'success' ? t('common.success') : t('common.failed')" />
              </div>
              <div class="audit-item__bottom">
                <span class="audit-item__target">{{ log.target }}</span>
                <span class="audit-item__time mono">{{ log.timestamp }}</span>
              </div>
              <div class="audit-item__user">
                <span class="led" :class="log.user === 'system' ? 'led--offline' : 'led--ok'" />
                {{ log.user }}
              </div>
            </div>
            <div v-if="auditLogs.length === 0 && !loading" class="audit-empty">
              {{ t('dashboard.noOperationRecords') }}
            </div>
          </div>
        </PageCard>
      </NGi>
    </NGrid>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.dashboard__sysbar {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  font-size: var(--fs-caption);
}

.sysbar-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.sysbar-label {
  color: var(--text-quaternary);
}

.sysbar-value {
  color: var(--ink);
  font-weight: 500;
}

.sysbar-divider {
  width: var(--border-width);
  height: var(--space-4);
  background: var(--line);
}

.sysbar-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--sub);
}

.dashboard__stats {
  margin: 0;
}

.dashboard__main {
  margin: 0;
}

.dashboard__chart {
  height: var(--chart-height);
  width: 100%;
}

.audit-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

/* 独占一行时：宽屏横向排列，小屏纵向排列 */
.audit-list--horizontal {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--space-3);
}

@media (max-width: 768px) {
  .audit-list--horizontal {
    grid-template-columns: 1fr;
  }
}

.audit-item {
  padding: var(--space-3);
  background: var(--surface2);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  transition: border-color var(--t-fast) var(--ease);
}

.audit-item:hover {
  border-color: var(--line);
}

.audit-item__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-1);
}

.audit-item__action {
  font-size: var(--fs-body);
  font-weight: 500;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.audit-item__bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-1);
  gap: var(--space-2);
}

.audit-item__target {
  font-size: var(--fs-caption);
  color: var(--sub);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.audit-item__time {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.audit-item__user {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.audit-empty {
  text-align: center;
  padding: var(--space-6);
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}

.dashboard__net-sep {
  color: var(--text-quaternary);
  margin: 0 var(--space-1);
}

@media (max-width: 1024px) {
  .dashboard__stats {
    :deep(.n-grid-item) {
      flex-basis: 50% !important;
      max-width: 50% !important;
    }
  }
  .dashboard__main {
    :deep(.n-grid-item) {
      flex-basis: 100% !important;
      max-width: 100% !important;
    }
  }
}

@media (max-width: 640px) {
  .dashboard__stats {
    :deep(.n-grid-item) {
      flex-basis: 100% !important;
      max-width: 100% !important;
    }
  }
  .dashboard__sysbar {
    flex-wrap: wrap;
  }
  .sysbar-divider {
    display: none;
  }
}

/* 小屏优化：≤480px */
@media (max-width: 480px) {
  /* 系统信息条：2列网格 */
  .dashboard__sysbar {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
  }
  .sysbar-item {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-0);
  }
  .sysbar-label {
    font-size: 10px;
  }
  .sysbar-value {
    font-size: var(--fs-body);
  }
  .sysbar-right {
    grid-column: 1 / -1;
    justify-content: center;
    padding-top: var(--space-2);
    border-top: 1px solid var(--line);
  }

  /* StatCard 简化：隐藏趋势，减小高度 */
  .dashboard__stats :deep(.stat-card__bottom) {
    display: none;
  }
  .dashboard__stats :deep(.stat-card) {
    padding: var(--space-3) var(--space-4);
  }
  .dashboard__stats :deep(.stat-card__top) {
    margin-bottom: var(--space-2);
  }
  .dashboard__stats :deep(.stat-card__value-wrap) {
    min-height: 24px;
  }

  /* 图表减小高度 */
  .dashboard__main :deep(.echarts) {
    height: 220px !important;
  }

  /* 图表头部按钮缩小 */
  .page-card__header-right .n-button {
    font-size: 11px;
    padding: 0 var(--space-2);
  }

  /* 审计日志：简化 */
  .dashboard__main :deep(.n-data-table-th:nth-child(3)),
  .dashboard__main :deep(.n-data-table-td:nth-child(3)) {
    display: none;
  }

  /* 最近操作：重新设计，防止文字溢出 */
  .audit-list {
    gap: var(--space-2);
  }
  .audit-item {
    padding: var(--space-2) var(--space-3);
  }
  .audit-item__top {
    gap: var(--space-2);
  }
  .audit-item__action {
    font-size: var(--fs-caption);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
  }
  .audit-item__bottom {
    gap: var(--space-2);
  }
  .audit-item__target {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
    font-size: 11px;
  }
  .audit-item__time {
    flex-shrink: 0;
    font-size: 11px;
  }
  .audit-item__user {
    display: none;
  }
}
</style>
