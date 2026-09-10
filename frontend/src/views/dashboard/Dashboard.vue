<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NGrid, NGi, NCard, NSpace, NTag, NButton, NIcon } from 'naive-ui'
// echarts 按需引入：只注册实际用到的图表和组件，减小包体积
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
import { StatCard, PageCard, StatusTag } from '@/components'
import { getOverview, getMetrics, getAuditLogs, normalizeOverview, type OverviewData, type AuditLogItem } from '@/api/dashboard'
import { useThemeTokens } from '@/composables/useThemeTokens'

const { t, locale } = useI18n()
const { tokens, refresh: refreshTokens } = useThemeTokens()

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

function formatBytes(mb: number | undefined | null): string {
  if (mb == null || isNaN(mb)) return '0 MB'
  if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
  return mb.toFixed(0) + ' MB'
}

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

function updateChart(data?: { timestamp: number; cpu: number; memory: number }[]) {
  if (!chartInstance) return

  const points = data || generateMockData()
  const times = points.map((p) => {
    const d = new Date(p.timestamp)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  })
  const cpuData = points.map((p) => p.cpu)
  const memData = points.map((p) => p.memory)

  const tk = tokens.value
  const textColor = tk.sub || '#9aa5b8'
  const lineColor = tk.line || '#2a3348'
  const cpuColor = tk.acc || '#1ba784'
  const memColor = tk.acc2 || '#2dd4a8'

  const cpuLabel = t('dashboard.cpu')
  const memLabel = t('dashboard.memory')

  chartInstance.setOption({
    grid: { top: 40, right: 20, bottom: 30, left: 50 },
    tooltip: {
      trigger: 'axis',
      backgroundColor: tk.surface || '#1a2030',
      borderColor: tk.line || '#2a3348',
      textStyle: { color: tk.ink || '#e8ecf2', fontSize: 12 },
      formatter: (params: any) => {
        let html = `<div style="font-weight:600;margin-bottom:4px">${params[0].axisValue}</div>`
        params.forEach((p: any) => {
          html += `<div style="display:flex;align-items:center;gap:6px">
            <span style="width:8px;height:8px;border-radius:50%;background:${p.color}"></span>
            ${p.seriesName}: <span style="font-family:monospace;font-weight:600">${p.value}%</span>
          </div>`
        })
        return html
      }
    },
    legend: {
      data: [cpuLabel, memLabel],
      top: 8,
      right: 20,
      textStyle: { color: textColor, fontSize: 12 },
      itemWidth: 12,
      itemHeight: 8
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLine: { lineStyle: { color: lineColor } },
      axisLabel: { color: textColor, fontSize: 11 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: textColor, fontSize: 11, formatter: '{value}%' },
      splitLine: { lineStyle: { color: lineColor, type: 'dashed' } }
    },
    series: [
      {
        name: cpuLabel,
        type: 'line',
        data: cpuData,
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

function generateMockData() {
  const points = []
  const now = Date.now()
  for (let i = 23; i >= 0; i--) {
    points.push({
      timestamp: now - i * 3600000,
      cpu: Math.round(20 + Math.random() * 40),
      memory: Math.round(40 + Math.random() * 30)
    })
  }
  return points
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
      // Mock data for development
      overview.value = {
        cpu: { usage: 23, cores: 4 },
        memory: { usedGB: 4.2, totalGB: 8, usage: 52 },
        disk: { usedGB: 27.2, totalGB: 40, usage: 68 },
        network: { uploadKbps: 2.1 * 1024, downloadKbps: 5.3 * 1024 },
        uptime: 86400 * 15,
        hostname: 'cloudpivot-node-01',
        os: 'Ubuntu 22.04 LTS',
        panelVersion: 'v1.0.0'
      }
    }

    if (metricsRes?.data?.points) {
      updateChart(metricsRes.data.points)
    } else {
      updateChart()
    }

    if (logsRes?.data?.items) {
      auditLogs.value = logsRes.data.items
    } else {
      auditLogs.value = [
        { id: 1, user: 'admin', action: '创建站点', target: 'example.com', result: 'success', timestamp: '2026-09-08 14:32:15' },
        { id: 2, user: 'admin', action: '重启服务', target: 'Nginx', result: 'success', timestamp: '2026-09-08 14:10:03' },
        { id: 3, user: 'system', action: '备份完成', target: 'daily-backup', result: 'success', timestamp: '2026-09-08 13:55:00' },
        { id: 4, user: 'admin', action: '安装应用', target: 'MySQL 8.0', result: 'success', timestamp: '2026-09-08 12:30:22' },
        { id: 5, user: 'admin', action: '登录失败', target: '192.168.1.100', result: 'failed', timestamp: '2026-09-08 11:45:08' }
      ]
    }
  } finally {
    loading.value = false
  }
}

function handleRangeChange(range: string) {
  metricRange.value = range
  getMetrics(range)
    .then((res) => {
      if (res.data?.points) updateChart(res.data.points)
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
          :trend="2.3"
          :trend-label="t('dashboard.comparedYesterday')"
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
          :trend="-1.2"
          :trend-label="t('dashboard.comparedYesterday')"
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
          :trend="0.5"
          :trend-label="t('dashboard.comparedYesterday')"
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
          <template #subtitle>
            <div class="chart-header">
              <span>{{ t('dashboard.cpuMemoryUsage') }}</span>
              <NSpace size="small">
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
            </div>
          </template>
          <div ref="chartRef" class="dashboard__chart" />
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
  gap: var(--space-4);
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
  color: var(--faint);
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

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
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
  color: var(--faint);
}

.audit-item__user {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--fs-caption);
  color: var(--faint);
}

.audit-empty {
  text-align: center;
  padding: var(--space-6);
  color: var(--faint);
  font-size: var(--fs-caption);
}

.dashboard__net-sep {
  color: var(--faint);
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
    gap: 2px;
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
  .chart-header {
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .chart-header .n-button {
    font-size: 11px;
    padding: 0 8px;
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
