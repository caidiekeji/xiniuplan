<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch, h } from 'vue'
import { NGrid, NGi, NTabs, NTabPane, NButton, NSelect, NProgress, NEmpty } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
// echarts 按需引入：只注册实际用到的图表和组件，减小包体积
import * as echarts from 'echarts/core'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
import { PageCard, DataTable, StatCard, StatusTag, ConfirmDialog } from '@/components'
import type { DataTableColumns } from 'naive-ui'
import { getOverview, getMetrics, normalizeOverview, type OverviewData } from '@/api/dashboard'
import {
  getProcesses,
  getDiskUsage,
  getNetworkStats,
  killProcess,
  type ProcessInfo,
  type DiskMount,
  type NetworkStats
} from '@/api/monitor'
import { useThemeTokens, hexToRgba, fsPx } from '@/composables/useThemeTokens'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()
const { tokens } = useThemeTokens()

/* ==================== 顶部概览 ==================== */

const overviewLoading = ref(false)
const overview = ref<OverviewData | null>(null)

// 接口失败时保持 null 显示真实空态（—），不编造假数据
const cpuPercent = computed(() => (overview.value?.cpu.usage != null ? overview.value.cpu.usage.toFixed(1) : '—'))
const memPercent = computed(() => (overview.value?.memory.usage != null ? overview.value.memory.usage.toFixed(1) : '—'))
const diskPercent = computed(() => (overview.value?.disk.usage != null ? overview.value.disk.usage.toFixed(1) : '—'))

function formatSpeed(kbps: number | undefined | null): string {
  if (kbps == null || isNaN(kbps)) return '0 KB/s'
  if (kbps >= 1024) return (kbps / 1024).toFixed(1) + ' MB/s'
  return kbps.toFixed(0) + ' KB/s'
}

async function loadOverview(silent = false) {
  if (!silent) overviewLoading.value = true
  try {
    const res = await getOverview()
    if (res.data) {
      overview.value = normalizeOverview(res.data)
    }
    // 失败或空响应时保持原值/空态，不用假数据填充
  } catch {
    /* 保持空态 */
  } finally {
    if (!silent) overviewLoading.value = false
  }
}

/* ==================== 趋势图 ==================== */

const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null
const metricRange = ref('24h')
const rangeOptions = computed(() => [
  { label: t('monitor.range1h'), value: '1h' },
  { label: t('monitor.range6h'), value: '6h' },
  { label: t('monitor.range24h'), value: '24h' },
  { label: t('monitor.range7d'), value: '7d' }
])

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
}

// 图表文字摘要：读屏用户看不到 Canvas，用 aria-label 承载最新值（审计 D10①）
const chartAriaLabel = ref('')
function updateChartAria(points: { cpu: number; memory: number }[]) {
  if (points.length === 0) {
    chartAriaLabel.value = t('monitor.chartAriaEmpty')
    return
  }
  const last = points[points.length - 1]
  const rangeLabel = rangeOptions.value.find((o) => o.value === metricRange.value)?.label || metricRange.value
  chartAriaLabel.value = t('monitor.chartAria', {
    cpu: last.cpu.toFixed(1),
    mem: last.memory.toFixed(1),
    range: rangeLabel
  })
}

let lastPoints: { timestamp: number; cpu: number; memory: number }[] | null = null

// 各范围的聚合桶秒数（与后端 GetMetrics 保持一致），用于识别采集断档
const BUCKET_SEC: Record<string, number> = { '1h': 60, '6h': 300, '24h': 3600, '7d': 3600 }

function updateChart(data?: { timestamp: number; cpu: number; memory: number }[]) {
  if (!chartInstance) return
  if (data) lastPoints = data
  // 空数据画空轴 + "暂无数据"；主题/语言切换时沿用上次数据重绘
  const points = data || lastPoints || []
  updateChartAria(points)
  const fmtTime = (ms: number) => {
    const d = new Date(ms) // 后端为 Unix 秒，调用处已转毫秒
    const mmdd = `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const hhmm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    return metricRange.value === '7d' || metricRange.value === '24h' ? `${mmdd} ${hhmm}` : hhmm
  }
  // 相邻点间隔超过 2 个聚合桶视为断档，插入 null 让折线断开，避免空洞被压缩成假趋势
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
  // CPU 主撞色青绿，内存信息蓝 —— 双线必须强区分
  const cpuColor = tk.acc || '#1ba784'
  const memColor = tk.info || '#60a5fa'
  // 图表字号统一走 token（审计 D01：12/13px 散点 → --fs-caption 一档）
  const fsCaption = fsPx('--fs-caption', 12)

  chartInstance.setOption({
    title: points.length === 0
      ? {
          text: t('monitor.noData'),
          left: 'center',
          top: 'middle',
          textStyle: { color: tk.faint || '#8b95a8', fontSize: fsCaption, fontWeight: 400 }
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
            <span style="width: var(--led-size); height: var(--led-size); border-radius: 50%; background: ${p.color}"></span>
            ${p.seriesName}: <span style="font-family:monospace;font-weight:600">${v == null ? '—' : v.toFixed(1) + '%'}</span>
          </div>`
        })
        return html
      }
    },
    legend: { data: ['CPU', t('monitor.memory')], top: 8, right: 20, textStyle: { color: textColor, fontSize: fsCaption }, itemWidth: 12, itemHeight: 8 },
    xAxis: { type: 'time', axisLine: { lineStyle: { color: lineColor } }, axisLabel: { color: textColor, fontSize: fsCaption, formatter: (val: number) => fmtTime(val) }, axisTick: { show: false }, splitLine: { show: false } },
    yAxis: { type: 'value', max: 100, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: textColor, fontSize: fsCaption, formatter: '{value}%' }, splitLine: { lineStyle: { color: lineColor, type: 'dashed' } } },
    series: [
      {
        name: 'CPU', type: 'line', data: cpuData, connectNulls: false, smooth: true, symbol: 'none', lineStyle: { width: 2, color: cpuColor },
        areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: hexToRgba(cpuColor, 0.25) }, { offset: 1, color: hexToRgba(cpuColor, 0.02) }]) }
      },
      {
        name: t('monitor.memory'), type: 'line', data: memData, connectNulls: false, smooth: true, symbol: 'none', lineStyle: { width: 2, color: memColor },
        areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: hexToRgba(memColor, 0.25) }, { offset: 1, color: hexToRgba(memColor, 0.02) }]) }
      }
    ]
  })
}

async function handleRangeChange(range: string) {
  metricRange.value = range
  await loadTrend()
}

/** 首屏/自动刷新的趋势数据加载（此前仅点范围按钮才拉数，首屏一直是空轴） */
async function loadTrend() {
  try {
    const res = await getMetrics(metricRange.value)
    if (res.data?.points && res.data.points.length > 0) updateChart(res.data.points)
    else updateChart()
  } catch {
    updateChart()
  }
}

/* ==================== Tab 1: 进程管理 ==================== */

const processes = ref<ProcessInfo[]>([])
const processesLoading = ref(false)
const procSort = ref<'cpu' | 'memory'>('cpu')
const procTop = ref(20)
const topOptions = [
  { label: 'Top 10', value: 10 },
  { label: 'Top 20', value: 20 },
  { label: 'Top 50', value: 50 }
]
const sortOptions = computed(() => [
  { label: t('monitor.sortByCpu'), value: 'cpu' },
  { label: t('monitor.sortByMemory'), value: 'memory' }
])

const killDialog = ref({ show: false, pid: 0, name: '' })
const killLoading = ref(false)

const processColumns = computed(() => [
  { title: 'PID', key: 'pid', width: 80, render: (row: any) => h('span', { class: 'mono' }, row.pid ?? '—') },
  { title: t('monitor.colProcess'), key: 'name', width: 160, ellipsis: { tooltip: true } },
  { title: 'CPU %', key: 'cpu', width: 100, render: (row: any) => h('span', { class: 'mono' }, row.cpu != null ? row.cpu.toFixed(1) : '—') },
  { title: t('monitor.colMemPercent'), key: 'memory', width: 100, render: (row: any) => h('span', { class: 'mono' }, row.memory != null ? row.memory.toFixed(1) : '—') },
  { title: t('monitor.colMemRss'), key: 'memory_rss', width: 110, render: (row: any) => h('span', { class: 'mono' }, formatBytes(row.memory_rss)) },
  {
    title: t('monitor.colStatus'), key: 'status', width: 100,
    render: (row: any) => {
      const map: Record<string, { status: any; text: string }> = {
        running: { status: 'running', text: t('monitor.statusRunning') },
        sleeping: { status: 'pending', text: t('monitor.statusSleeping') },
        stopped: { status: 'stopped', text: t('monitor.statusStopped') },
        zombie: { status: 'error', text: t('monitor.statusZombie') }
      }
      const cfg = map[row.status] || map.sleeping
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('monitor.colActions'), key: 'actions', width: 90,
    render: (row: any) =>
      h(NButton, { size: 'tiny', type: 'error', ghost: true, onClick: () => { killDialog.value = { show: true, pid: row.pid, name: row.name } } }, () => t('monitor.terminate'))
  }
] as DataTableColumns)

function formatBytes(kb: number | undefined | null): string {
  if (kb == null || isNaN(kb)) return '—'
  if (kb >= 1048576) return (kb / 1048576).toFixed(1) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(1) + ' MB'
  return kb.toFixed(0) + ' KB'
}

async function loadProcesses(silent = false) {
  if (!silent) processesLoading.value = true
  try {
    const res = await getProcesses(procTop.value, procSort.value)
    if (res.data && Array.isArray(res.data)) {
      processes.value = res.data
    } else if (!silent) {
      message.error(t('monitor.loadProcessesFailed'))
    }
  } catch (e: any) {
    processes.value = []
    if (!silent) message.error(e.message || t('monitor.loadProcessesFailed'))
  } finally {
    if (!silent) processesLoading.value = false
  }
}

async function handleKillProcess() {
  killLoading.value = true
  try {
    await killProcess(killDialog.value.pid)
    message.success(t('monitor.processKilled', { name: killDialog.value.name }))
    killDialog.value.show = false
    loadProcesses()
  } catch (e: any) {
    message.error(e.message || t('monitor.killFailed'))
  } finally {
    killLoading.value = false
  }
}

/* ==================== Tab 2: 磁盘分析 ==================== */

const disks = ref<DiskMount[]>([])
const disksLoading = ref(false)

async function loadDisks() {
  disksLoading.value = true
  try {
    const res = await getDiskUsage()
    if (res.data && Array.isArray(res.data)) {
      disks.value = res.data
    } else {
      disks.value = []
    }
  } catch {
    disks.value = []
    message.error(t('monitor.loadDiskFailed'))
  } finally {
    disksLoading.value = false
  }
}


function diskColor(percent: number): string {
  if (percent >= 85) return 'error'
  if (percent >= 70) return 'warning'
  return 'success'
}

function diskStatus(percent: number) {
  if (percent >= 85) return { status: 'error' as const, text: t('monitor.diskDanger') }
  if (percent >= 70) return { status: 'warning' as const, text: t('monitor.diskHigh') }
  return { status: 'running' as const, text: t('monitor.diskNormal') }
}

/* ==================== Tab 3: 网络连接 ==================== */

const netStats = ref<NetworkStats | null>(null)
const netLoading = ref(false)
const pieChartRef = ref<HTMLDivElement | null>(null)
let pieChartInstance: echarts.ECharts | null = null

// 连接分布饼图文字摘要（审计 D10①）
const pieAriaLabel = ref('')
function updatePieAria() {
  const n = netStats.value
  pieAriaLabel.value = !n
    ? t('monitor.pieAriaEmpty')
    : t('monitor.pieAria', {
        established: n.established,
        listening: n.listening,
        timeWait: n.time_wait,
        closeWait: n.close_wait,
        other: n.other
      })
}

async function loadNetwork() {
  netLoading.value = true
  try {
    const res = await getNetworkStats()
    if (res.data) {
      netStats.value = res.data
    } else {
      netStats.value = null
    }
  } catch {
    netStats.value = null
  } finally {
    netLoading.value = false
  }
  nextTick(() => updatePieChart())
}

function initPieChart() {
  if (!pieChartRef.value) return
  pieChartInstance = echarts.init(pieChartRef.value)
  updatePieChart()
}

function updatePieChart() {
  updatePieAria()
  if (!pieChartInstance || !netStats.value) return
  const tk = tokens.value
  const textColor = tk.sub || '#9aa5b8'
  const fsCaption = fsPx('--fs-caption', 12)
  const data = [
    { value: netStats.value.established, name: 'ESTABLISHED', itemStyle: { color: tk.acc || '#1ba784' } },
    { value: netStats.value.listening, name: 'LISTENING', itemStyle: { color: tk.info || '#60a5fa' } },
    { value: netStats.value.time_wait, name: 'TIME_WAIT', itemStyle: { color: tk.warn || '#f59e0b' } },
    { value: netStats.value.close_wait, name: 'CLOSE_WAIT', itemStyle: { color: tk.err || '#e85d4d' } },
    { value: netStats.value.other, name: 'OTHER', itemStyle: { color: tk.faint || '#5c6678' } }
  ]

  pieChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      backgroundColor: tk.surface || '#1a2030',
      borderColor: tk.line || '#2a3348',
      textStyle: { color: tk.ink || '#e8ecf2', fontSize: fsCaption },
      formatter: '{b}: {c} ({d}%)'
    },
    legend: { orient: 'vertical', right: 10, top: 'center', textStyle: { color: textColor, fontSize: fsCaption }, itemWidth: 10, itemHeight: 10 },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['38%', '50%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: tk.surface || '#1a2030', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: fsCaption, fontWeight: 'bold', color: textColor } },
      data
    }]
  })
}

/* ==================== 自动刷新 & 生命周期 ==================== */

let refreshTimer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
  refreshTimer = setInterval(() => {
    loadProcesses(true)
    loadOverview(true)
  }, 5000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

function handleResize() {
  chartInstance?.resize()
  pieChartInstance?.resize()
}

// 容器尺寸监听：侧栏折叠/展开不触发 window resize，用 ResizeObserver 兜底
let chartResizeObserver: ResizeObserver | null = null
let pieResizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  initChart()
  initPieChart()
  loadTrend()
  await Promise.all([loadOverview(), loadProcesses(), loadDisks(), loadNetwork()])
  startAutoRefresh()
  window.addEventListener('resize', handleResize)
  if (chartRef.value && 'ResizeObserver' in window) {
    chartResizeObserver = new ResizeObserver(() => chartInstance?.resize())
    chartResizeObserver.observe(chartRef.value)
  }
  if (pieChartRef.value && 'ResizeObserver' in window) {
    pieResizeObserver = new ResizeObserver(() => pieChartInstance?.resize())
    pieResizeObserver.observe(pieChartRef.value)
  }
})

onBeforeUnmount(() => {
  stopAutoRefresh()
  window.removeEventListener('resize', handleResize)
  chartResizeObserver?.disconnect()
  chartResizeObserver = null
  pieResizeObserver?.disconnect()
  pieResizeObserver = null
  chartInstance?.dispose()
  chartInstance = null
  pieChartInstance?.dispose()
  pieChartInstance = null
})

watch(
  () => tokens.value.acc,
  () => {
    nextTick(() => {
      if (chartInstance) updateChart()
      if (pieChartInstance) updatePieChart()
    })
  }
)
</script>

<template>
  <div class="monitor-page">
    <!-- 顶部 4 StatCard -->
    <NGrid cols="1 s:2 l:4" :x-gap="16" :y-gap="16" responsive="screen">
      <NGi>
        <StatCard :title="t('monitor.cpuUsage')" :value="cpuPercent" unit="%" :subtitle="t('monitor.cores', { n: overview?.cpu.cores || 4 })" :loading="overviewLoading" accent="acc" />
      </NGi>
      <NGi>
        <StatCard
          :title="t('monitor.memoryUsage')"
          :value="memPercent"
          unit="%"
          :subtitle="t('monitor.usageOfTotal', { used: overview ? overview.memory.usedGB.toFixed(1) : '—', total: overview ? overview.memory.totalGB.toFixed(0) : 8 })"
          :loading="overviewLoading"
          accent="acc2"
        />
      </NGi>
      <NGi>
        <StatCard
          :title="t('monitor.diskUsage')"
          :value="diskPercent"
          unit="%"
          :subtitle="t('monitor.usageOfTotal', { used: overview ? overview.disk.usedGB.toFixed(1) : '—', total: overview ? overview.disk.totalGB.toFixed(1) : 40 })"
          :loading="overviewLoading"
          accent="warn"
        />
      </NGi>
      <NGi>
        <StatCard :title="t('monitor.networkTraffic')" :value="overview ? formatSpeed(overview.network.downloadKbps) : '—'" :subtitle="t('monitor.downloadUpload')" :loading="overviewLoading" accent="info">
          <template #subtitle>
            <span class="mono" style="color: var(--acc2)">↓ {{ overview ? formatSpeed(overview.network.downloadKbps) : '—' }}</span>
            <span style="color: var(--text-quaternary); margin: 0 var(--space-1)">·</span>
            <span class="mono" style="color: var(--acc)">↑ {{ overview ? formatSpeed(overview.network.uploadKbps) : '—' }}</span>
          </template>
        </StatCard>
      </NGi>
    </NGrid>

    <!-- 趋势图（PageCard 只有 extra slot，此前误写 subtitle 致范围按钮被静默丢弃） -->
    <PageCard :title="t('monitor.resourceTrend')" :subtitle="t('monitor.trendSubtitle')">
      <template #extra>
        <div class="chart-header__btns">
          <NButton
            v-for="opt in rangeOptions"
            :key="opt.value"
            size="tiny"
            :type="metricRange === opt.value ? 'primary' : 'default'"
            @click="handleRangeChange(opt.value)"
          >
            {{ opt.label }}
          </NButton>
        </div>
      </template>
      <div ref="chartRef" class="monitor__chart" role="img" :aria-label="chartAriaLabel" />
    </PageCard>

    <!-- 三标签页 -->
    <NTabs type="line" animated>
      <!-- Tab 1: 进程管理 -->
      <NTabPane name="processes" :tab="t('monitor.tabProcesses')">
        <PageCard :title="t('monitor.procList')" :subtitle="t('monitor.procListSubtitle')">
          <template #extra>
            <div class="monitor__toolbar">
              <NSelect v-model:value="procSort" :options="sortOptions" size="small" style="width: var(--input-sm)" @update:value="loadProcesses" />
              <NSelect v-model:value="procTop" :options="topOptions" size="small" style="width: var(--input-sm)" @update:value="loadProcesses" />
            </div>
          </template>
          <DataTable
            :columns="processColumns"
            :data="processes"
            :loading="processesLoading"
            :row-key="(row: any) => row.pid"
            :pagination="false"
            size="small"
          />
        </PageCard>
      </NTabPane>

      <!-- Tab 2: 磁盘分析 -->
      <NTabPane name="disk" :tab="t('monitor.tabDisk')">
        <PageCard :title="t('monitor.diskUsageTitle')" :subtitle="t('monitor.diskUsageSubtitle')" :loading="disksLoading">
          <div v-if="disks.length === 0 && !disksLoading" class="monitor__empty">
            <NEmpty :description="t('common.noData')" size="small" />
          </div>
          <div v-else class="disk-list">
            <div v-for="d in disks" :key="d.mount_point" class="disk-item">
              <div class="disk-item__header">
                <div class="disk-item__mount">
                  <span class="mono disk-item__path">{{ d.mount_point }}</span>
                  <span v-if="d.filesystem" class="disk-item__fs mono">{{ d.filesystem }}</span>
                </div>
                <StatusTag :status="diskStatus(d.usage_percent).status" :text="diskStatus(d.usage_percent).text" />
              </div>
              <NProgress
                :percentage="d.usage_percent"
                :type="diskColor(d.usage_percent) as any"
                :indicator-placement="'inside'"
                :height="18"
                :border-radius="4"
              />
              <div class="disk-item__meta">
                <span class="mono">{{ t('monitor.diskTotal', { n: d.total }) }}</span>
                <span class="mono">{{ t('monitor.diskUsed', { n: d.used }) }}</span>
                <span class="mono">{{ t('monitor.diskAvailable', { n: d.available }) }}</span>
                <span class="mono disk-item__percent">{{ d.usage_percent }}%</span>
              </div>
            </div>
          </div>
        </PageCard>
      </NTabPane>

      <!-- Tab 3: 网络连接 -->
      <NTabPane name="network" :tab="t('monitor.tabNetwork')">
        <PageCard :title="t('monitor.netStats')" :subtitle="t('monitor.netStatsSubtitle')">
          <NGrid cols="1 s:3" :x-gap="16" :y-gap="16" responsive="screen" class="monitor__net-stats">
            <NGi>
              <StatCard :title="t('monitor.totalConnections')" :value="netStats?.total_connections ?? '—'" :loading="netLoading" accent="info" />
            </NGi>
            <NGi>
              <StatCard title="ESTABLISHED" :value="netStats?.established ?? '—'" :loading="netLoading" accent="acc2" />
            </NGi>
            <NGi>
              <StatCard title="LISTENING" :value="netStats?.listening ?? '—'" :loading="netLoading" accent="acc" />
            </NGi>
          </NGrid>
          <div class="monitor__pie-wrap">
            <div ref="pieChartRef" class="monitor__pie" role="img" :aria-label="pieAriaLabel" />
          </div>
        </PageCard>
      </NTabPane>
    </NTabs>

    <!-- 终止进程确认 -->
    <ConfirmDialog
      :show="killDialog.show"
      :title="t('monitor.killTitle')"
      :content="t('monitor.killContent', { name: killDialog.name, pid: killDialog.pid })"
      :confirm-text="t('monitor.terminate')"
      type="error"
      :loading="killLoading"
      @update:show="(v: boolean) => (killDialog.show = v)"
      @confirm="handleKillProcess"
    />
  </div>
</template>

<style scoped>
.monitor-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.chart-header__btns {
  display: flex;
  gap: var(--space-1);
}

.monitor__chart {
  height: var(--chart-height);
  width: 100%;
}

.monitor__toolbar {
  display: flex;
  gap: var(--space-2);
}

/* 磁盘列表 */
.disk-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

/* 列表空态：与磁盘项保持同宽居中，避免空白塌陷 */
.monitor__empty {
  display: flex;
  justify-content: center;
  padding: var(--space-6) 0;
}

.disk-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.disk-item__mount {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
}

.disk-item__path {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
}

.disk-item__fs {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.disk-item__meta {
  display: flex;
  gap: var(--space-5);
  margin-top: var(--space-2);
  font-size: var(--fs-caption);
  color: var(--sub);
}

.disk-item__percent {
  margin-left: auto;
  font-weight: 600;
  color: var(--ink);
}

/* 网络饼图 */
.monitor__net-stats {
  margin-bottom: var(--space-4);
}

.monitor__pie-wrap {
  border-top: 1px solid var(--line);
  padding-top: var(--space-4);
}

.monitor__pie {
  height: 280px;
  width: 100%;
}

@media (max-width: 1024px) {
  .monitor__net-stats :deep(.n-grid-item) {
    flex-basis: 33.33% !important;
    max-width: 33.33% !important;
  }
}
</style>
