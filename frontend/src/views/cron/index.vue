<script setup lang="ts">
import { ref, reactive, computed, onMounted, h, watch } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup,
  NRadio, NSpace, NTag, NDrawer, NDrawerContent, NThing, useMessage
} from 'naive-ui'
import {
  AddOutline, PlayOutline, StopOutline, RefreshOutline,
  DocumentTextOutline, TrashOutline, CreateOutline, TimerOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import {
  getCronTasks, createCronTask, updateCronTask, deleteCronTask,
  enableCronTask, disableCronTask, triggerCronTask, getCronLogs,
  type CronTask, type CronTaskType, type CronTaskStatus,
  type CreateCronTaskData, type CronLog
} from '@/api/cron'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()

/* ========== 常量映射 ========== */
const typeLabelMap: Record<CronTaskType, string> = {
  shell: 'Shell',
  backup: '备份',
  url: 'URL',
  app_restart: '应用重启'
}

const typeTagTypeMap: Record<CronTaskType, 'info' | 'success' | 'warning' | 'default'> = {
  shell: 'info',
  backup: 'success',
  url: 'warning',
  app_restart: 'default'
}

const statusMap: Record<CronTaskStatus, { status: StatusType; text: string }> = {
  enabled: { status: 'running', text: '已启用' },
  disabled: { status: 'stopped', text: '已禁用' }
}

/* ========== Cron 预设 ========== */
const cronPresets = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天', value: '0 3 * * *' },
  { label: '每周', value: '0 3 * * 0' },
  { label: '每月', value: '0 3 1 * *' }
]

/* ========== Cron 人类可读描述 ========== */
function describeCron(expr: string): string {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return expr
  const [minute, hour, day, month, week] = parts

  if (minute === '*' && hour === '*' && day === '*' && month === '*' && week === '*') return '每分钟'
  if (minute === '0' && hour === '*' && day === '*' && month === '*' && week === '*') return '每小时整点'
  if (hour === '*' && day === '*' && month === '*' && week === '*') return `每小时第 ${minute} 分`
  if (day === '*' && month === '*' && week === '*') {
    return `每天 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
  }
  if (month === '*' && week === '*' && day !== '*') {
    return `每月 ${day} 日 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
  }
  if (day === '*' && month === '*' && week !== '*') {
    const weekMap: Record<string, string> = { '0': '日', '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' }
    return `每周${weekMap[week] || week} ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
  }
  return expr
}

/* ========== 列表状态 ========== */
const loading = ref(false)
const tasks = ref<CronTask[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

/* ========== Mock 数据兜底 ========== */
function generateMockTasks(): { items: CronTask[]; total: number } {
  const mock: CronTask[] = [
    { id: 1, name: '每日网站备份', type: 'backup', schedule: '0 3 * * *', target: 'website', status: 'enabled', lastRunAt: '2026-09-08 03:00:00', nextRunAt: '2026-09-09 03:00:00', remark: '每天凌晨3点备份所有站点', createdAt: '2026-08-01 10:00:00' },
    { id: 2, name: '数据库清理', type: 'shell', schedule: '0 4 * * 0', command: 'mysql -e "DELETE FROM logs WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)"', status: 'enabled', lastRunAt: '2026-09-07 04:00:00', nextRunAt: '2026-09-14 04:00:00', remark: '每周日清理30天前日志', createdAt: '2026-08-05 14:00:00' },
    { id: 3, name: '健康检查', type: 'url', schedule: '*/5 * * * *', target: 'https://example.com/health', status: 'enabled', lastRunAt: '2026-09-08 14:30:00', nextRunAt: '2026-09-08 14:35:00', remark: '每5分钟检查站点健康', createdAt: '2026-08-10 09:00:00' },
    { id: 4, name: '应用重启', type: 'app_restart', schedule: '0 5 * * *', target: 'nginx', status: 'disabled', lastRunAt: '2026-09-01 05:00:00', nextRunAt: '—', remark: '每日凌晨5点重启Nginx', createdAt: '2026-08-15 11:00:00' },
    { id: 5, name: 'SSL证书续期', type: 'shell', schedule: '0 2 1 * *', command: 'certbot renew --quiet', status: 'enabled', lastRunAt: '2026-09-01 02:00:00', nextRunAt: '2026-10-01 02:00:00', remark: '每月1日续期证书', createdAt: '2026-07-20 16:00:00' }
  ]
  return { items: mock, total: mock.length }
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await getCronTasks(page.value, pageSize.value)
    if (res.data?.items) {
      tasks.value = res.data.items
      total.value = res.data.total
    } else {
      const mock = generateMockTasks()
      tasks.value = mock.items
      total.value = mock.total
    }
  } catch {
    const mock = generateMockTasks()
    tasks.value = mock.items
    total.value = mock.total
  } finally {
    loading.value = false
  }
}

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { page.value = p; loadTasks() },
  'onUpdate:page-size': (ps: number) => { pageSize.value = ps; page.value = 1; loadTasks() }
}))

/* ========== 表格列 ========== */
const columns = computed<DataTableColumns>(() => [
  {
    title: '任务名称',
    key: 'name',
    width: 180,
    render: (row: any) => h('div', { class: 'cron-name' }, [
      h('span', { class: 'cron-name__main' }, row.name),
      row.remark
        ? h('span', { class: 'cron-name__remark' }, row.remark)
        : null
    ])
  },
  {
    title: '类型',
    key: 'type',
    width: 100,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as CronTaskType],
      bordered: false,
      round: true
    }, { default: () => typeLabelMap[row.type as CronTaskType] })
  },
  {
    title: 'Cron 表达式',
    key: 'cron',
    width: 240,
    render: (row: any) => h('div', { class: 'cron-expr' }, [
      h('span', { class: 'cron-expr__code mono' }, row.schedule),
      h('span', { class: 'cron-expr__desc' }, describeCron(row.schedule))
    ])
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const cfg = statusMap[row.status as CronTaskStatus]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '上次执行',
    key: 'lastRunAt',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.lastRunAt || '—')
  },
  {
    title: '下次执行',
    key: 'nextRunAt',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.nextRunAt || '—')
  },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    render: (row: any) => {
      const isEnabled = row.status === 'enabled'
      const actionLoading = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: isEnabled ? 'warning' : 'success',
            loading: actionLoading,
            onClick: () => isEnabled ? handleDisable(row.id) : handleEnable(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(isEnabled ? StopOutline : PlayOutline) }),
            default: () => isEnabled ? '禁用' : '启用'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTrigger(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
            default: () => '触发'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openLogs(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }),
            default: () => '日志'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => '编辑'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => '删除'
          })
        ]
      })
    }
  }
])

/* ========== 操作 loading ========== */
const actionLoadingId = ref<number | null>(null)

async function handleEnable(id: number) {
  actionLoadingId.value = id
  try {
    await enableCronTask(id)
    message.success('任务已启用')
    loadTasks()
  } catch (err: any) {
    message.error(err.message || '启用失败')
  } finally {
    actionLoadingId.value = null
  }
}

async function handleDisable(id: number) {
  actionLoadingId.value = id
  try {
    await disableCronTask(id)
    message.success('任务已禁用')
    loadTasks()
  } catch (err: any) {
    message.error(err.message || '禁用失败')
  } finally {
    actionLoadingId.value = null
  }
}

async function handleTrigger(id: number) {
  actionLoadingId.value = id
  try {
    await triggerCronTask(id)
    message.success('任务已触发')
  } catch (err: any) {
    message.error(err.message || '触发失败')
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 创建/编辑任务对话框 ========== */
const showForm = ref(false)
const formLoading = ref(false)
const formRef = ref()
const isEdit = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  name: '',
  type: 'shell' as CronTaskType,
  cronMode: 'preset' as 'preset' | 'custom',
  cronPreset: '0 3 * * *',
  cronMinute: '0',
  cronHour: '3',
  cronDay: '*',
  cronMonth: '*',
  cronWeek: '*',
  command: '',
  target: '',
  remark: ''
})

const cronExpression = computed(() => {
  if (form.cronMode === 'preset') return form.cronPreset
  return `${form.cronMinute} ${form.cronHour} ${form.cronDay} ${form.cronMonth} ${form.cronWeek}`
})

const cronDescription = computed(() => describeCron(cronExpression.value))

const formRules = {
  name: {
    required: true,
    message: '请输入任务名称',
    trigger: 'blur'
  },
  command: {
    required: true,
    message: '请输入 Shell 命令',
    trigger: 'blur'
  },
  target: {
    required: true,
    message: '请输入目标',
    trigger: 'blur'
  }
}

const backupTypeOptions = [
  { label: '网站备份', value: 'website' },
  { label: '数据库备份', value: 'database' },
  { label: '系统备份', value: 'system' }
]

function resetForm() {
  form.name = ''
  form.type = 'shell'
  form.cronMode = 'preset'
  form.cronPreset = '0 3 * * *'
  form.cronMinute = '0'
  form.cronHour = '3'
  form.cronDay = '*'
  form.cronMonth = '*'
  form.cronWeek = '*'
  form.command = ''
  form.target = ''
  form.remark = ''
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(id: number) {
  const task = tasks.value.find((t) => t.id === id)
  if (!task) return
  isEdit.value = true
  editId.value = id
  resetForm()
  form.name = task.name
  form.type = task.type
  form.command = task.command || ''
  form.target = task.target || ''
  form.remark = task.remark || ''
  // 解析 cron
  const parts = task.schedule.trim().split(/\s+/)
  if (parts.length === 5) {
    const matched = cronPresets.find((p) => p.value === task.schedule)
    if (matched) {
      form.cronMode = 'preset'
      form.cronPreset = task.schedule
    } else {
      form.cronMode = 'custom'
      form.cronMinute = parts[0]
      form.cronHour = parts[1]
      form.cronDay = parts[2]
      form.cronMonth = parts[3]
      form.cronWeek = parts[4]
    }
  }
  showForm.value = true
}

async function handleFormSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  const data: CreateCronTaskData = {
    name: form.name,
    type: form.type,
    schedule: cronExpression.value,
    remark: form.remark
  }

  if (form.type === 'shell') {
    if (!form.command.trim()) {
      message.warning('请输入 Shell 命令')
      return
    }
    data.command = form.command
  } else if (form.type === 'backup') {
    if (!form.target) {
      message.warning('请选择备份类型')
      return
    }
    data.target = form.target
  } else if (form.type === 'url') {
    if (!form.target.trim()) {
      message.warning('请输入 URL')
      return
    }
    data.target = form.target
  } else if (form.type === 'app_restart') {
    if (!form.target.trim()) {
      message.warning('请输入容器名')
      return
    }
    data.target = form.target
  }

  formLoading.value = true
  try {
    if (isEdit.value && editId.value !== null) {
      await updateCronTask(editId.value, data)
      message.success('任务已更新')
    } else {
      await createCronTask(data)
      message.success('任务已创建')
    }
    showForm.value = false
    loadTasks()
  } catch (err: any) {
    message.error(err.message || '保存失败')
  } finally {
    formLoading.value = false
  }
}

/* ========== 执行日志抽屉 ========== */
const showLogs = ref(false)
const logsLoading = ref(false)
const logsTaskId = ref<number | null>(null)
const logsTaskName = ref('')
const logs = ref<CronLog[]>([])
const logsTotal = ref(0)
const logsPage = ref(1)
const logsPageSize = ref(10)
const expandedLogId = ref<number | null>(null)

const logResultMap: Record<string, { status: StatusType; text: string }> = {
  success: { status: 'running', text: '成功' },
  failed: { status: 'error', text: '失败' },
  running: { status: 'warning', text: '执行中' }
}

function formatDuration(ms?: number): string {
  if (!ms && ms !== 0) return '—'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

async function openLogs(id: number) {
  logsTaskId.value = id
  const task = tasks.value.find((t) => t.id === id)
  logsTaskName.value = task?.name || ''
  logsPage.value = 1
  expandedLogId.value = null
  showLogs.value = true
  await loadLogs()
}

async function loadLogs() {
  if (logsTaskId.value === null) return
  logsLoading.value = true
  try {
    const res = await getCronLogs(logsTaskId.value, logsPage.value, logsPageSize.value)
    if (res.data?.items) {
      logs.value = res.data.items
      logsTotal.value = res.data.total
    } else {
      const mock: CronLog[] = [
        { id: 1, startTime: '2026-09-08 03:00:00', endTime: '2026-09-08 03:02:15', durationMs: 135000, result: 'success', output: '[2026-09-08 03:00:00] Starting backup...\n[2026-09-08 03:00:01] Compressing files...\n[2026-09-08 03:02:10] Upload completed\n[2026-09-08 03:02:15] Backup finished successfully' },
        { id: 2, startTime: '2026-09-07 03:00:00', endTime: '2026-09-07 03:01:45', durationMs: 105000, result: 'success', output: '[2026-09-07 03:00:00] Starting backup...\n[2026-09-07 03:01:45] Backup finished successfully' },
        { id: 3, startTime: '2026-09-06 03:00:00', endTime: '2026-09-06 03:00:30', durationMs: 30000, result: 'failed', output: '[2026-09-06 03:00:00] Starting backup...\n[2026-09-06 03:00:30] ERROR: Connection refused to storage endpoint\n[2026-09-06 03:00:30] Backup failed' }
      ]
      logs.value = mock
      logsTotal.value = mock.length
    }
  } catch {
    logs.value = []
    logsTotal.value = 0
  } finally {
    logsLoading.value = false
  }
}

const logsPagination = computed(() => ({
  page: logsPage.value,
  pageSize: logsPageSize.value,
  itemCount: logsTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  'onUpdate:page': (p: number) => { logsPage.value = p; loadLogs() },
  'onUpdate:page-size': (ps: number) => { logsPageSize.value = ps; logsPage.value = 1; loadLogs() }
}))

const logColumns = computed<DataTableColumns>(() => [
  {
    title: '开始时间',
    key: 'startTime',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.startTime)
  },
  {
    title: '结束时间',
    key: 'endTime',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.endTime || '—')
  },
  {
    title: '耗时',
    key: 'durationMs',
    width: 100,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, formatDuration(row.durationMs))
  },
  {
    title: '结果',
    key: 'result',
    width: 90,
    render: (row: any) => {
      const cfg = logResultMap[row.result] || logResultMap.failed
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '输出',
    key: 'output',
    render: (row: any) => {
      const isExpanded = expandedLogId.value === row.id
      const preview = (row.output || '').split('\n')[0] || '—'
      return h('div', { class: 'log-output' }, [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => { expandedLogId.value = isExpanded ? null : row.id }
        }, {
          default: () => isExpanded ? '收起' : '展开'
        }),
        isExpanded
          ? h('div', { class: 'log-output__detail' }, [
              h(CodeEditor, {
                modelValue: row.output || '',
                language: 'log',
                readOnly: true,
                height: '180px',
                minimap: false
              })
            ])
          : h('span', { class: 'log-output__preview mono' }, preview.length > 60 ? preview.slice(0, 60) + '...' : preview)
      ])
    }
  }
])

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteId = ref<number | null>(null)
const deleteName = ref('')

function openDelete(id: number) {
  deleteId.value = id
  const task = tasks.value.find((t) => t.id === id)
  deleteName.value = task?.name || ''
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteCronTask(deleteId.value)
    message.success('任务已删除')
    showDelete.value = false
    loadTasks()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<template>
  <PageCard title="计划任务" subtitle="定时执行 Shell 脚本 / 备份 / URL 访问 / 应用重启">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        创建任务
      </NButton>
    </template>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="tasks"
      :loading="loading"
      :pagination="pagination"
    />
  </PageCard>

  <!-- 创建/编辑任务对话框 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? '编辑任务' : '创建任务'"
    :loading="formLoading"
    :width="640"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem label="任务名称" path="name">
        <NInput v-model:value="form.name" placeholder="例如：每日网站备份" clearable />
      </NFormItem>

      <NFormItem label="任务类型" path="type">
        <NRadioGroup v-model:value="form.type">
          <NSpace size="medium">
            <NRadio value="shell">Shell 脚本</NRadio>
            <NRadio value="backup">备份任务</NRadio>
            <NRadio value="url">URL 访问</NRadio>
            <NRadio value="app_restart">应用重启</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <!-- Cron 配置 -->
      <NFormItem label="执行周期">
        <div class="cron-config">
          <div class="cron-config__mode">
            <NRadioGroup v-model:value="form.cronMode" size="small">
              <NSpace size="small">
                <NRadio value="preset">预设</NRadio>
                <NRadio value="custom">自定义</NRadio>
              </NSpace>
            </NRadioGroup>
          </div>

          <div v-if="form.cronMode === 'preset'" class="cron-config__preset">
            <NSelect
              v-model:value="form.cronPreset"
              :options="cronPresets"
              placeholder="选择预设周期"
              style="width: var(--input-2xl)"
            />
          </div>

          <div v-else class="cron-config__custom">
            <NSpace size="small" align="center">
              <div class="cron-field">
                <span class="cron-field__label">分</span>
                <NInput v-model:value="form.cronMinute" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">时</span>
                <NInput v-model:value="form.cronHour" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">日</span>
                <NInput v-model:value="form.cronDay" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">月</span>
                <NInput v-model:value="form.cronMonth" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">周</span>
                <NInput v-model:value="form.cronWeek" placeholder="*" class="cron-field__input" />
              </div>
            </NSpace>
          </div>

          <div class="cron-config__preview">
            <NIcon><TimerOutline /></NIcon>
            <span class="cron-config__expr mono">{{ cronExpression }}</span>
            <span class="cron-config__desc">{{ cronDescription }}</span>
          </div>
        </div>
      </NFormItem>

      <!-- 命令/目标 -->
      <NFormItem v-if="form.type === 'shell'" label="Shell 命令" path="command">
        <NInput
          v-model:value="form.command"
          type="textarea"
          placeholder="输入要执行的 Shell 命令"
          :rows="4"
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'backup'" label="备份类型" path="target">
        <NSelect
          v-model:value="form.target"
          :options="backupTypeOptions"
          placeholder="选择备份类型"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'url'" label="URL 地址" path="target">
        <NInput
          v-model:value="form.target"
          placeholder="例如 https://example.com/cron"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'app_restart'" label="容器名" path="target">
        <NInput
          v-model:value="form.target"
          placeholder="例如 nginx / mysql"
          clearable
        />
      </NFormItem>

      <NFormItem label="备注" path="remark">
        <NInput v-model:value="form.remark" placeholder="可选备注信息" clearable />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 执行日志抽屉 -->
  <NDrawer
    :show="showLogs"
    :width="780"
    placement="right"
    @update:show="(v: boolean) => (showLogs = v)"
  >
    <NDrawerContent
      :title="`执行日志 — ${logsTaskName}`"
      closable
    >
      <div class="logs-drawer">
        <DataTable
          :columns="logColumns"
          :data="logs"
          :loading="logsLoading"
          :pagination="logsPagination"
          size="small"
        />
      </div>
    </NDrawerContent>
  </NDrawer>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    title="删除任务"
    :content="`确定要删除任务「${deleteName}」吗？此操作不可撤销。`"
    confirm-text="删除"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.cron-name {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.cron-name__main {
  font-weight: 500;
  color: var(--ink);
}

.cron-name__remark {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.cron-expr {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.cron-expr__code {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-mono);
}

.cron-expr__desc {
  font-size: var(--fs-caption);
  color: var(--faint);
}

/* Cron 配置区 */
.cron-config {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.cron-config__preset,
.cron-config__custom {
  width: 100%;
}

.cron-field {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
}

.cron-field__label {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.cron-field__input {
  width: 70px;
  text-align: center;
}

.cron-field__input :deep(.n-input__input-el) {
  text-align: center;
  font-family: var(--font-mono);
}

.cron-config__preview {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  border: 1px solid var(--line);
}

.cron-config__preview .n-icon {
  color: var(--acc);
  flex-shrink: 0;
}

.cron-config__expr {
  font-weight: 600;
  color: var(--ink);
  font-size: var(--fs-mono);
}

.cron-config__desc {
  font-size: var(--fs-caption);
  color: var(--sub);
  margin-left: var(--space-2);
  padding-left: var(--space-3);
  border-left: 1px solid var(--line);
}

/* 日志抽屉 */
.logs-drawer {
  height: 100%;
}

.log-output {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.log-output__preview {
  font-size: var(--fs-caption);
  color: var(--sub);
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-output__detail {
  width: 100%;
}
</style>
