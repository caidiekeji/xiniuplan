<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch, h } from 'vue'
import { NGrid, NGi, NTabs, NTabPane, NButton, NSelect, NProgress, NTag, NIcon, useMessage } from 'naive-ui'
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
import { useThemeTokens, hexToRgba } from '@/composables/useThemeTokens'

const message = useMessage()
const { tokens } = useThemeTokens()

/* ==================== 顶部概览 ==================== */

const overviewLoading = ref(false)
const overview = ref<OverviewData | null>(null)

const cpuPercent = computed(() => (overview.value?.cpu.usage ?? 0).toFixed(1))
const memPercent = computed(() => (overview.value?.memory.usage ?? 0).toFixed(1))
const diskPercent = computed(() => (overview.value?.disk.usage ?? 0).toFixed(1))

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
    } else if (!overview.value) {
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
  } catch {
    if (!overview.value) {
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
  } finally {
    if (!silent) overviewLoading.value = false
  }
}

/* ==================== 趋势图 ==================== */

const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null
const metricRange = ref('24h')
const rangeOptions = [
  { label: '1小时', value: '1h' },
  { label: '6小时', value: '6h' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' }
]

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
}

function generateMockPoints(count: number) {
  const points = []
  const now = Date.now()
  const interval = count <= 24 ? 3600000 : count <= 168 ? 3600000 : 86400000
  for (let i = count - 1; i >= 0; i--) {
    points.push({
      timestamp: now - i * interval,
      cpu: Math.round(15 + Math.random() * 50),
      memory: Math.round(40 + Math.random() * 30)
    })
  }
  return points
}

function updateChart(data?: { timestamp: number; cpu: number; memory: number }[]) {
  if (!chartInstance) return
  const points = data || generateMockPoints(24)
  const times = points.map((p) => {
    const d = new Date(p.timestamp)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  })
  const cpuData = points.map((p) => p.cpu)
  const memData = points.map((p) => p.memory)

  const t = tokens.value
  const textColor = t.sub || '#9aa5b8'
  const lineColor = t.line || '#2a3348'
  const cpuColor = t.acc || '#1ba784'
  const memColor = t.acc2 || '#2dd4a8'

  chartInstance.setOption({
    grid: { top: 40, right: 20, bottom: 30, left: 50 },
    tooltip: {
      trigger: 'axis',
      backgroundColor: t.surface || '#1a2030',
      borderColor: t.line || '#2a3348',
      textStyle: { color: t.ink || '#e8ecf2', fontSize: 12 },
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
    legend: { data: ['CPU', '内存'], top: 8, right: 20, textStyle: { color: textColor, fontSize: 12 }, itemWidth: 12, itemHeight: 8 },
    xAxis: { type: 'category', data: times, axisLine: { lineStyle: { color: lineColor } }, axisLabel: { color: textColor, fontSize: 11 }, axisTick: { show: false } },
    yAxis: { type: 'value', max: 100, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: textColor, fontSize: 11, formatter: '{value}%' }, splitLine: { lineStyle: { color: lineColor, type: 'dashed' } } },
    series: [
      {
        name: 'CPU', type: 'line', data: cpuData, smooth: true, symbol: 'none', lineStyle: { width: 2, color: cpuColor },
        areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: hexToRgba(cpuColor, 0.25) }, { offset: 1, color: hexToRgba(cpuColor, 0.02) }]) }
      },
      {
        name: '内存', type: 'line', data: memData, smooth: true, symbol: 'none', lineStyle: { width: 2, color: memColor },
        areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: hexToRgba(memColor, 0.25) }, { offset: 1, color: hexToRgba(memColor, 0.02) }]) }
      }
    ]
  })
}

async function handleRangeChange(range: string) {
  metricRange.value = range
  try {
    const res = await getMetrics(range)
    if (res.data?.points) updateChart(res.data.points)
    else updateChart(generateMockPoints(range === '1h' ? 12 : range === '6h' ? 24 : range === '24h' ? 24 : 168))
  } catch {
    updateChart(generateMockPoints(range === '1h' ? 12 : range === '6h' ? 24 : range === '24h' ? 24 : 168))
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
const sortOptions = [
  { label: '按 CPU', value: 'cpu' },
  { label: '按内存', value: 'memory' }
]

const killDialog = ref({ show: false, pid: 0, name: '' })
const killLoading = ref(false)

const processColumns = computed(() => [
  { title: 'PID', key: 'pid', width: 80, render: (row: any) => h('span', { class: 'mono' }, row.pid ?? '—') },
  { title: '进程名', key: 'name', width: 160, ellipsis: { tooltip: true } },
  { title: 'CPU %', key: 'cpu', width: 100, render: (row: any) => h('span', { class: 'mono' }, row.cpu != null ? row.cpu.toFixed(1) : '—') },
  { title: '内存 %', key: 'memory', width: 100, render: (row: any) => h('span', { class: 'mono' }, row.memory != null ? row.memory.toFixed(1) : '—') },
  { title: '内存 RSS', key: 'memory_rss', width: 110, render: (row: any) => h('span', { class: 'mono' }, formatBytes(row.memory_rss)) },
  {
    title: '状态', key: 'status', width: 100,
    render: (row: any) => {
      const map: Record<string, { status: any; text: string }> = {
        running: { status: 'running', text: '运行中' },
        sleeping: { status: 'pending', text: '休眠' },
        stopped: { status: 'stopped', text: '已停止' },
        zombie: { status: 'error', text: '僵尸' }
      }
      const cfg = map[row.status] || map.sleeping
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '操作', key: 'actions', width: 90,
    render: (row: any) =>
      h(NButton, { size: 'tiny', type: 'error', ghost: true, onClick: () => { killDialog.value = { show: true, pid: row.pid, name: row.name } } }, () => '终止')
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
    } else if (processes.value.length === 0) {
      generateMockProcesses()
    }
  } catch {
    if (processes.value.length === 0) generateMockProcesses()
  } finally {
    if (!silent) processesLoading.value = false
  }
}

function generateMockProcesses() {
  const names = ['nginx', 'mysql', 'redis-server', 'node', 'python3', 'docker', 'containerd', 'systemd', 'sshd', 'postgres', 'java', 'php-fpm', 'mongod', 'elasticsearch', 'kubelet']
  const statuses: ProcessInfo['status'][] = ['running', 'sleeping', 'running', 'running', 'sleeping']
  processes.value = names.slice(0, procTop.value > names.length ? names.length : procTop.value).map((name, i) => ({
    pid: 1000 + i * 137,
    name,
    cpu: +(Math.random() * 30).toFixed(1),
    memory: +(Math.random() * 15).toFixed(1),
    memory_rss: Math.round(Math.random() * 500000),
    status: statuses[i % statuses.length]
  }))
  if (procSort.value === 'cpu') processes.value.sort((a, b) => b.cpu - a.cpu)
  else processes.value.sort((a, b) => b.memory - a.memory)
}

async function handleKillProcess() {
  killLoading.value = true
  try {
    await killProcess(killDialog.value.pid)
    message.success(`进程 ${killDialog.value.name} 已终止`)
    killDialog.value.show = false
    loadProcesses()
  } catch (e: any) {
    message.error(e.message || '终止失败')
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
      generateMockDisks()
    }
  } catch {
    generateMockDisks()
  } finally {
    disksLoading.value = false
  }
}

function generateMockDisks() {
  disks.value = [
    { mount_point: '/', filesystem: '/dev/sda1', total: 40, used: 27.2, available: 12.8, usage_percent: 68 },
    { mount_point: '/home', filesystem: '/dev/sdb1', total: 100, used: 45.3, available: 54.7, usage_percent: 45 },
    { mount_point: '/data', filesystem: '/dev/sdc1', total: 500, used: 320.5, available: 179.5, usage_percent: 64 },
    { mount_point: '/var/lib/docker', filesystem: 'overlay', total: 40, used: 15.8, available: 24.2, usage_percent: 39 }
  ]
}

function diskColor(percent: number): string {
  if (percent >= 85) return 'error'
  if (percent >= 70) return 'warning'
  return 'success'
}

function diskStatus(percent: number) {
  if (percent >= 85) return { status: 'error' as const, text: '危险' }
  if (percent >= 70) return { status: 'warning' as const, text: '偏高' }
  return { status: 'running' as const, text: '正常' }
}

/* ==================== Tab 3: 网络连接 ==================== */

const netStats = ref<NetworkStats | null>(null)
const netLoading = ref(false)
const pieChartRef = ref<HTMLDivElement | null>(null)
let pieChartInstance: echarts.ECharts | null = null

async function loadNetwork() {
  netLoading.value = true
  try {
    const res = await getNetworkStats()
    if (res.data) {
      netStats.value = res.data
    } else {
      generateMockNetwork()
    }
  } catch {
    generateMockNetwork()
  } finally {
    netLoading.value = false
  }
  nextTick(() => updatePieChart())
}

function generateMockNetwork() {
  netStats.value = {
    total_connections: 247,
    established: 156,
    listening: 23,
    time_wait: 45,
    close_wait: 12,
    other: 11,
    upload_speed: 2150,
    download_speed: 5430
  }
}

function initPieChart() {
  if (!pieChartRef.value) return
  pieChartInstance = echarts.init(pieChartRef.value)
  updatePieChart()
}

function updatePieChart() {
  if (!pieChartInstance || !netStats.value) return
  const t = tokens.value
  const textColor = t.sub || '#9aa5b8'
  const data = [
    { value: netStats.value.established, name: 'ESTABLISHED', itemStyle: { color: t.acc2 || '#2dd4a8' } },
    { value: netStats.value.listening, name: 'LISTENING', itemStyle: { color: t.acc || '#1ba784' } },
    { value: netStats.value.time_wait, name: 'TIME_WAIT', itemStyle: { color: t.info || '#60a5fa' } },
    { value: netStats.value.close_wait, name: 'CLOSE_WAIT', itemStyle: { color: t.warn || '#1ba784' } },
    { value: netStats.value.other, name: 'OTHER', itemStyle: { color: t.faint || '#5c6678' } }
  ]

  pieChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      backgroundColor: t.surface || '#1a2030',
      borderColor: t.line || '#2a3348',
      textStyle: { color: t.ink || '#e8ecf2', fontSize: 12 },
      formatter: '{b}: {c} ({d}%)'
    },
    legend: { orient: 'vertical', right: 10, top: 'center', textStyle: { color: textColor, fontSize: 11 }, itemWidth: 10, itemHeight: 10 },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['38%', '50%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: t.surface || '#1a2030', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 13, fontWeight: 'bold', color: textColor } },
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
        <StatCard title="CPU 使用率" :value="cpuPercent" unit="%" :subtitle="`${overview?.cpu.cores || 4} 核`" :loading="overviewLoading" accent="acc" />
      </NGi>
      <NGi>
        <StatCard
          title="内存使用"
          :value="memPercent"
          unit="%"
          :subtitle="`${overview ? overview.memory.usedGB.toFixed(1) : '—'} / ${overview ? overview.memory.totalGB.toFixed(0) : 8} GB`"
          :loading="overviewLoading"
          accent="acc2"
        />
      </NGi>
      <NGi>
        <StatCard
          title="磁盘使用"
          :value="diskPercent"
          unit="%"
          :subtitle="`${overview ? overview.disk.usedGB.toFixed(1) : '—'} / ${overview ? overview.disk.totalGB.toFixed(1) : 40} GB`"
          :loading="overviewLoading"
          accent="warn"
        />
      </NGi>
      <NGi>
        <StatCard title="网络流量" :value="overview ? formatSpeed(overview.network.downloadKbps) : '—'" subtitle="下载 / 上传" :loading="overviewLoading" accent="info">
          <template #subtitle>
            <span class="mono" style="color: var(--acc2)">↓ {{ overview ? formatSpeed(overview.network.downloadKbps) : '—' }}</span>
            <span style="color: var(--faint); margin: 0 4px">·</span>
            <span class="mono" style="color: var(--acc)">↑ {{ overview ? formatSpeed(overview.network.uploadKbps) : '—' }}</span>
          </template>
        </StatCard>
      </NGi>
    </NGrid>

    <!-- 趋势图 -->
    <PageCard title="资源趋势" subtitle="CPU 与内存使用率实时监控">
      <template #subtitle>
        <div class="chart-header">
          <span>CPU 与内存使用率</span>
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
        </div>
      </template>
      <div ref="chartRef" class="monitor__chart" />
    </PageCard>

    <!-- 三标签页 -->
    <NTabs type="line" animated>
      <!-- Tab 1: 进程管理 -->
      <NTabPane name="processes" tab="进程管理">
        <PageCard title="进程列表" subtitle="实时监控系统进程资源占用">
          <template #extra>
            <div class="monitor__toolbar">
              <NSelect v-model:value="procSort" :options="sortOptions" size="small" style="width: 110px" @update:value="loadProcesses" />
              <NSelect v-model:value="procTop" :options="topOptions" size="small" style="width: 110px" @update:value="loadProcesses" />
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
      <NTabPane name="disk" tab="磁盘分析">
        <PageCard title="磁盘使用情况" subtitle="各挂载点容量与使用率">
          <div v-if="disksLoading" class="monitor__loading">加载中...</div>
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
                <span class="mono">总量 {{ d.total }} GB</span>
                <span class="mono">已用 {{ d.used }} GB</span>
                <span class="mono">可用 {{ d.available }} GB</span>
                <span class="mono disk-item__percent">{{ d.usage_percent }}%</span>
              </div>
            </div>
          </div>
        </PageCard>
      </NTabPane>

      <!-- Tab 3: 网络连接 -->
      <NTabPane name="network" tab="网络连接">
        <PageCard title="网络连接统计" subtitle="TCP 连接状态分布">
          <NGrid cols="1 s:3" :x-gap="16" :y-gap="16" responsive="screen" class="monitor__net-stats">
            <NGi>
              <StatCard title="总连接数" :value="netStats?.total_connections ?? '—'" :loading="netLoading" accent="info" />
            </NGi>
            <NGi>
              <StatCard title="ESTABLISHED" :value="netStats?.established ?? '—'" :loading="netLoading" accent="acc2" />
            </NGi>
            <NGi>
              <StatCard title="LISTENING" :value="netStats?.listening ?? '—'" :loading="netLoading" accent="acc" />
            </NGi>
          </NGrid>
          <div class="monitor__pie-wrap">
            <div ref="pieChartRef" class="monitor__pie" />
          </div>
        </PageCard>
      </NTabPane>
    </NTabs>

    <!-- 终止进程确认 -->
    <ConfirmDialog
      :show="killDialog.show"
      title="终止进程"
      :content="`确定要终止进程 ${killDialog.name} (PID: ${killDialog.pid}) 吗？可能导致相关服务中断。`"
      confirm-text="终止"
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
  gap: var(--space-4);
}

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
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

.monitor__loading {
  padding: var(--space-8) 0;
  text-align: center;
  color: var(--faint);
}

/* 磁盘列表 */
.disk-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
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
  color: var(--faint);
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
