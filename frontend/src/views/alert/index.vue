<script setup lang="ts">
import { ref, reactive, computed, onMounted, h, watch } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NCheckboxGroup,
  NCheckbox, NSpace, NTag, NSwitch, NTabs, NTabPane, useMessage
} from 'naive-ui'
import {
  AddOutline, PlayOutline, StopOutline, CreateOutline,
  TrashOutline, FlashOutline, ChevronDownOutline, ChevronForwardOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getAlertRules, createAlertRule, updateAlertRule, deleteAlertRule,
  enableAlertRule, disableAlertRule, testAlertRule,
  getAlertChannels, createAlertChannel, updateAlertChannel, deleteAlertChannel,
  testAlertChannel, getAlertHistory,
  type AlertRule, type AlertChannel, type AlertHistory,
  type MetricType, type AlertCondition, type ChannelType,
  type AlertStatus, type NotificationStatus,
  type CreateAlertRuleData, type CreateAlertChannelData
} from '@/api/alert'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()

/* ==================== 常量映射 ==================== */

const metricLabelMap: Record<MetricType, string> = {
  cpu: 'CPU',
  memory: '内存',
  disk: '磁盘',
  network: '网络'
}

const metricTagTypeMap: Record<MetricType, 'info' | 'success' | 'warning' | 'error' | 'default'> = {
  cpu: 'info',
  memory: 'success',
  disk: 'warning',
  network: 'default'
}

const metricUnitMap: Record<MetricType, string> = {
  cpu: '%',
  memory: '%',
  disk: '%',
  network: 'MB/s'
}

const conditionLabelMap: Record<AlertCondition, string> = {
  gt: '大于',
  lt: '小于',
  eq: '等于'
}

const conditionSymbolMap: Record<AlertCondition, string> = {
  gt: '>',
  lt: '<',
  eq: '='
}

const channelLabelMap: Record<ChannelType, string> = {
  email: '邮件',
  dingtalk: '钉钉',
  wechat: '企业微信',
  feishu: '飞书',
  telegram: 'Telegram',
  webhook: 'Webhook'
}

const channelTagTypeMap: Record<ChannelType, 'info' | 'success' | 'warning' | 'error' | 'default'> = {
  email: 'info',
  dingtalk: 'success',
  wechat: 'warning',
  feishu: 'default',
  telegram: 'error',
  webhook: 'info'
}

const historyStatusMap: Record<AlertStatus, { status: StatusType; text: string }> = {
  triggered: { status: 'error', text: '触发中' },
  recovered: { status: 'running', text: '已恢复' }
}

const notificationStatusMap: Record<NotificationStatus, { status: StatusType; text: string }> = {
  pending: { status: 'warning', text: '发送中' },
  sent: { status: 'running', text: '已发送' },
  failed: { status: 'error', text: '失败' }
}

const metricOptions = Object.entries(metricLabelMap).map(([value, label]) => ({ label, value }))
const conditionOptions = Object.entries(conditionLabelMap).map(([value, label]) => ({ label, value }))
const channelTypeOptions = Object.entries(channelLabelMap).map(([value, label]) => ({ label, value }))
const historyStatusOptions: Array<{ label: string; value: string }> = [
  { label: '全部', value: '' },
  { label: '触发中', value: 'triggered' },
  { label: '已恢复', value: 'recovered' }
]

/* ==================== Tab 1: 告警规则 ==================== */

const rulesLoading = ref(false)
const rules = ref<AlertRule[]>([])
const rulesTotal = ref(0)
const rulesPage = ref(1)
const rulesPageSize = ref(20)
const ruleActionLoadingId = ref<number | null>(null)

// 渠道列表（供规则表单多选）
const channelList = ref<AlertChannel[]>([])

async function loadChannels() {
  try {
    const res = await getAlertChannels()
    if (res.data?.items) {
      channelList.value = res.data.items
    }
  } catch {
    channelList.value = []
  }
}

function getChannelName(id: string): string {
  const ch = channelList.value.find((c) => String(c.id) === String(id))
  return ch?.name || id
}

function describeRule(row: AlertRule): string {
  const metric = metricLabelMap[row.metric_type]
  const symbol = conditionSymbolMap[row.condition]
  const unit = metricUnitMap[row.metric_type]
  return `${metric} ${symbol} ${row.threshold}${unit}`
}

async function loadRules() {
  rulesLoading.value = true
  try {
    const res = await getAlertRules(rulesPage.value, rulesPageSize.value)
    if (res.data?.items) {
      rules.value = res.data.items
      rulesTotal.value = res.data.total
    } else {
      rules.value = []
      rulesTotal.value = 0
    }
  } catch {
    rules.value = []
    rulesTotal.value = 0
  } finally {
    rulesLoading.value = false
  }
}

const rulesPagination = computed(() => ({
  page: rulesPage.value,
  pageSize: rulesPageSize.value,
  itemCount: rulesTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { rulesPage.value = p; loadRules() },
  'onUpdate:page-size': (ps: number) => { rulesPageSize.value = ps; rulesPage.value = 1; loadRules() }
}))

const ruleColumns = computed<DataTableColumns>(() => [
  {
    title: '规则名称',
    key: 'name',
    width: 180,
    render: (row: any) => h('div', { class: 'rule-name' }, [
      h('span', { class: 'rule-name__main' }, row.name),
      row.remark
        ? h('span', { class: 'rule-name__remark' }, row.remark)
        : null
    ])
  },
  {
    title: '指标类型',
    key: 'metric_type',
    width: 100,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: metricTagTypeMap[row.metric_type as MetricType],
      bordered: false,
      round: true
    }, { default: () => metricLabelMap[row.metric_type as MetricType] })
  },
  {
    title: '条件',
    key: 'condition',
    width: 160,
    render: (row: any) => h('span', { class: 'mono rule-condition' }, describeRule(row as AlertRule))
  },
  {
    title: '持续时间',
    key: 'duration',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, `${row.duration} 分钟`)
  },
  {
    title: '通知渠道',
    key: 'channels',
    width: 200,
    render: (row: any) => {
      const ids: string[] = row.channels || []
      if (ids.length === 0) return h('span', { style: 'color:var(--faint);font-size:var(--fs-caption)' }, '—')
      return h(NSpace, { size: 'small', wrap: true }, {
        default: () => ids.map((id) => h(NTag, {
          size: 'small',
          type: 'default',
          bordered: false,
          round: true
        }, { default: () => getChannelName(id) }))
      })
    }
  },
  {
    title: '状态',
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.enabled ? 'running' : 'stopped',
      text: row.enabled ? '启用' : '禁用'
    })
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row: any) => {
      const isEnabled = row.enabled
      const actionLoading = ruleActionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: isEnabled ? 'warning' : 'success',
            loading: actionLoading,
            onClick: () => isEnabled ? handleDisableRule(row.id) : handleEnableRule(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(isEnabled ? StopOutline : PlayOutline) }),
            default: () => isEnabled ? '禁用' : '启用'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTestRule(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(FlashOutline) }),
            default: () => '测试'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openRuleEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => '编辑'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openRuleDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => '删除'
          })
        ]
      })
    }
  }
])

async function handleEnableRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await enableAlertRule(id)
    message.success('规则已启用')
    loadRules()
  } catch (err: any) {
    message.error(err.message || '启用失败')
  } finally {
    ruleActionLoadingId.value = null
  }
}

async function handleDisableRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await disableAlertRule(id)
    message.success('规则已禁用')
    loadRules()
  } catch (err: any) {
    message.error(err.message || '禁用失败')
  } finally {
    ruleActionLoadingId.value = null
  }
}

async function handleTestRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await testAlertRule(id)
    message.success('测试消息已发送')
  } catch (err: any) {
    message.error(err.message || '测试失败')
  } finally {
    ruleActionLoadingId.value = null
  }
}

/* --- 规则创建/编辑对话框 --- */

const showRuleForm = ref(false)
const ruleFormLoading = ref(false)
const ruleFormRef = ref()
const ruleIsEdit = ref(false)
const ruleEditId = ref<number | null>(null)

const ruleForm = reactive({
  name: '',
  metric_type: 'cpu' as MetricType,
  condition: 'gt' as AlertCondition,
  threshold: '80',
  duration: '5',
  channels: [] as string[],
  remark: ''
})

const ruleFormRules = {
  name: { required: true, message: '请输入规则名称', trigger: 'blur' },
  threshold: { required: true, message: '请输入阈值', trigger: 'blur' }
}

function resetRuleForm() {
  ruleForm.name = ''
  ruleForm.metric_type = 'cpu'
  ruleForm.condition = 'gt'
  ruleForm.threshold = '80'
  ruleForm.duration = '5'
  ruleForm.channels = []
  ruleForm.remark = ''
}

function openRuleCreate() {
  ruleIsEdit.value = false
  ruleEditId.value = null
  resetRuleForm()
  showRuleForm.value = true
}

function openRuleEdit(id: number) {
  const rule = rules.value.find((r) => r.id === id)
  if (!rule) return
  ruleIsEdit.value = true
  ruleEditId.value = id
  resetRuleForm()
  ruleForm.name = rule.name
  ruleForm.metric_type = rule.metric_type
  ruleForm.condition = rule.condition
  ruleForm.threshold = String(rule.threshold)
  ruleForm.duration = String(rule.duration)
  ruleForm.channels = [...rule.channels]
  ruleForm.remark = rule.remark || ''
  showRuleForm.value = true
}

async function handleRuleFormSubmit() {
  try {
    await ruleFormRef.value?.validate()
  } catch {
    return
  }

  const data: CreateAlertRuleData = {
    name: ruleForm.name,
    metric_type: ruleForm.metric_type,
    condition: ruleForm.condition,
    threshold: Number(ruleForm.threshold),
    duration: Number(ruleForm.duration),
    channels: ruleForm.channels,
    remark: ruleForm.remark
  }

  ruleFormLoading.value = true
  try {
    if (ruleIsEdit.value && ruleEditId.value !== null) {
      await updateAlertRule(ruleEditId.value, data)
      message.success('规则已更新')
    } else {
      await createAlertRule(data)
      message.success('规则已创建')
    }
    showRuleForm.value = false
    loadRules()
  } catch (err: any) {
    message.error(err.message || '保存失败')
  } finally {
    ruleFormLoading.value = false
  }
}

/* --- 规则删除确认 --- */

const showRuleDelete = ref(false)
const ruleDeleteLoading = ref(false)
const ruleDeleteId = ref<number | null>(null)
const ruleDeleteName = ref('')

function openRuleDelete(id: number) {
  ruleDeleteId.value = id
  const rule = rules.value.find((r) => r.id === id)
  ruleDeleteName.value = rule?.name || ''
  showRuleDelete.value = true
}

async function handleRuleDeleteConfirm() {
  if (ruleDeleteId.value === null) return
  ruleDeleteLoading.value = true
  try {
    await deleteAlertRule(ruleDeleteId.value)
    message.success('规则已删除')
    showRuleDelete.value = false
    loadRules()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    ruleDeleteLoading.value = false
  }
}

/* ==================== Tab 2: 通知渠道 ==================== */

const channelsLoading = ref(false)
const channels = ref<AlertChannel[]>([])
const channelsTotal = ref(0)
const channelsPage = ref(1)
const channelsPageSize = ref(20)
const channelActionLoadingId = ref<number | null>(null)

async function loadChannelsList() {
  channelsLoading.value = true
  try {
    const res = await getAlertChannels()
    if (res.data?.items) {
      channels.value = res.data.items
      channelsTotal.value = res.data.total
    } else {
      channels.value = []
      channelsTotal.value = 0
    }
  } catch {
    channels.value = []
    channelsTotal.value = 0
  } finally {
    channelsLoading.value = false
  }
}

const channelsPagination = computed(() => ({
  page: channelsPage.value,
  pageSize: channelsPageSize.value,
  itemCount: channelsTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { channelsPage.value = p; loadChannelsList() },
  'onUpdate:page-size': (ps: number) => { channelsPageSize.value = ps; channelsPage.value = 1; loadChannelsList() }
}))

const channelColumns = computed<DataTableColumns>(() => [
  {
    title: '渠道名称',
    key: 'name',
    width: 180,
    render: (row: any) => h('span', { class: 'channel-name' }, row.name)
  },
  {
    title: '类型',
    key: 'type',
    width: 120,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: channelTagTypeMap[row.type as ChannelType],
      bordered: false,
      round: true
    }, { default: () => channelLabelMap[row.type as ChannelType] })
  },
  {
    title: '状态',
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.enabled ? 'running' : 'stopped',
      text: row.enabled ? '启用' : '禁用'
    })
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.created_at || '—')
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row: any) => {
      const actionLoading = channelActionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTestChannel(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(FlashOutline) }),
            default: () => '测试'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openChannelEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => '编辑'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openChannelDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => '删除'
          })
        ]
      })
    }
  }
])

async function handleTestChannel(id: number) {
  channelActionLoadingId.value = id
  try {
    await testAlertChannel(id)
    message.success('测试消息已发送')
  } catch (err: any) {
    message.error(err.message || '测试失败')
  } finally {
    channelActionLoadingId.value = null
  }
}

/* --- 渠道创建/编辑对话框 --- */

const showChannelForm = ref(false)
const channelFormLoading = ref(false)
const channelFormRef = ref()
const channelIsEdit = ref(false)
const channelEditId = ref<number | null>(null)

const channelForm = reactive({
  name: '',
  type: 'email' as ChannelType,
  enabled: true,
  // 动态配置字段
  email_recipients: '',
  dingtalk_webhook: '',
  dingtalk_secret: '',
  wechat_webhook: '',
  feishu_webhook: '',
  telegram_bot_token: '',
  telegram_chat_id: '',
  webhook_url: '',
  webhook_headers: ''
})

const channelFormRules = {
  name: { required: true, message: '请输入渠道名称', trigger: 'blur' }
}

function resetChannelForm() {
  channelForm.name = ''
  channelForm.type = 'email'
  channelForm.enabled = true
  channelForm.email_recipients = ''
  channelForm.dingtalk_webhook = ''
  channelForm.dingtalk_secret = ''
  channelForm.wechat_webhook = ''
  channelForm.feishu_webhook = ''
  channelForm.telegram_bot_token = ''
  channelForm.telegram_chat_id = ''
  channelForm.webhook_url = ''
  channelForm.webhook_headers = ''
}

function buildChannelConfig(): Record<string, any> {
  switch (channelForm.type) {
    case 'email':
      return { recipients: channelForm.email_recipients }
    case 'dingtalk':
      return {
        webhook: channelForm.dingtalk_webhook,
        secret: channelForm.dingtalk_secret || undefined
      }
    case 'wechat':
      return { webhook: channelForm.wechat_webhook }
    case 'feishu':
      return { webhook: channelForm.feishu_webhook }
    case 'telegram':
      return {
        bot_token: channelForm.telegram_bot_token,
        chat_id: channelForm.telegram_chat_id
      }
    case 'webhook':
      return {
        url: channelForm.webhook_url,
        headers: channelForm.webhook_headers || undefined
      }
    default:
      return {}
  }
}

function loadChannelConfig(ch: AlertChannel) {
  const cfg = ch.config || {}
  switch (ch.type) {
    case 'email':
      channelForm.email_recipients = cfg.recipients || ''
      break
    case 'dingtalk':
      channelForm.dingtalk_webhook = cfg.webhook || ''
      channelForm.dingtalk_secret = cfg.secret || ''
      break
    case 'wechat':
      channelForm.wechat_webhook = cfg.webhook || ''
      break
    case 'feishu':
      channelForm.feishu_webhook = cfg.webhook || ''
      break
    case 'telegram':
      channelForm.telegram_bot_token = cfg.bot_token || ''
      channelForm.telegram_chat_id = cfg.chat_id || ''
      break
    case 'webhook':
      channelForm.webhook_url = cfg.url || ''
      channelForm.webhook_headers = cfg.headers || ''
      break
  }
}

function openChannelCreate() {
  channelIsEdit.value = false
  channelEditId.value = null
  resetChannelForm()
  showChannelForm.value = true
}

function openChannelEdit(id: number) {
  const ch = channels.value.find((c) => c.id === id)
  if (!ch) return
  channelIsEdit.value = true
  channelEditId.value = id
  resetChannelForm()
  channelForm.name = ch.name
  channelForm.type = ch.type
  channelForm.enabled = ch.enabled
  loadChannelConfig(ch)
  showChannelForm.value = true
}

async function handleChannelFormSubmit() {
  try {
    await channelFormRef.value?.validate()
  } catch {
    return
  }

  const data: CreateAlertChannelData = {
    name: channelForm.name,
    type: channelForm.type,
    config: buildChannelConfig(),
    enabled: channelForm.enabled
  }

  channelFormLoading.value = true
  try {
    if (channelIsEdit.value && channelEditId.value !== null) {
      await updateAlertChannel(channelEditId.value, data)
      message.success('渠道已更新')
    } else {
      await createAlertChannel(data)
      message.success('渠道已创建')
    }
    showChannelForm.value = false
    loadChannelsList()
    loadChannels()
  } catch (err: any) {
    message.error(err.message || '保存失败')
  } finally {
    channelFormLoading.value = false
  }
}

/* --- 渠道删除确认 --- */

const showChannelDelete = ref(false)
const channelDeleteLoading = ref(false)
const channelDeleteId = ref<number | null>(null)
const channelDeleteName = ref('')

function openChannelDelete(id: number) {
  channelDeleteId.value = id
  const ch = channels.value.find((c) => c.id === id)
  channelDeleteName.value = ch?.name || ''
  showChannelDelete.value = true
}

async function handleChannelDeleteConfirm() {
  if (channelDeleteId.value === null) return
  channelDeleteLoading.value = true
  try {
    await deleteAlertChannel(channelDeleteId.value)
    message.success('渠道已删除')
    showChannelDelete.value = false
    loadChannelsList()
    loadChannels()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    channelDeleteLoading.value = false
  }
}

/* ==================== Tab 3: 告警历史 ==================== */

const historyLoading = ref(false)
const history = ref<AlertHistory[]>([])
const historyTotal = ref(0)
const historyPage = ref(1)
const historyPageSize = ref(20)
const historyFilterRule = ref<string | null>('')
const historyFilterStatus = ref<string | null>('')
const expandedHistoryId = ref<number | null>(null)

// 规则下拉选项（供历史筛选）
const ruleOptions = computed(() => [
  { label: '全部规则', value: '' },
  ...rules.value.map((r) => ({ label: r.name, value: String(r.id) }))
])

async function loadHistory() {
  historyLoading.value = true
  try {
    const params: any = {
      page: historyPage.value,
      pageSize: historyPageSize.value
    }
    if (historyFilterRule.value) params.rule_id = Number(historyFilterRule.value)
    if (historyFilterStatus.value) params.status = historyFilterStatus.value
    const res = await getAlertHistory(params)
    if (res.data?.items) {
      history.value = res.data.items
      historyTotal.value = res.data.total
    } else {
      history.value = []
      historyTotal.value = 0
    }
  } catch {
    history.value = []
    historyTotal.value = 0
  } finally {
    historyLoading.value = false
  }
}

function handleHistoryFilter() {
  historyPage.value = 1
  loadHistory()
}

function handleHistoryReset() {
  historyFilterRule.value = ''
  historyFilterStatus.value = ''
  historyPage.value = 1
  loadHistory()
}

const historyPagination = computed(() => ({
  page: historyPage.value,
  pageSize: historyPageSize.value,
  itemCount: historyTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { historyPage.value = p; loadHistory() },
  'onUpdate:page-size': (ps: number) => { historyPageSize.value = ps; historyPage.value = 1; loadHistory() }
}))

const historyColumns = computed<DataTableColumns>(() => [
  {
    title: '',
    key: 'expand',
    width: 40,
    render: (row: any) => {
      const hasError = !!row.notification_error
      if (!hasError) return null
      const isExpanded = expandedHistoryId.value === row.id
      return h(NButton, {
        size: 'tiny',
        quaternary: true,
        circle: true,
        onClick: () => { expandedHistoryId.value = isExpanded ? null : row.id }
      }, {
        icon: () => h(NIcon, null, { default: () => h(isExpanded ? ChevronDownOutline : ChevronForwardOutline) })
      })
    }
  },
  {
    title: '触发时间',
    key: 'triggered_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.triggered_at || '—')
  },
  {
    title: '规则名称',
    key: 'rule_name',
    width: 160,
    render: (row: any) => h('span', { class: 'history-rule-name' }, row.rule_name)
  },
  {
    title: '指标类型',
    key: 'metric_type',
    width: 90,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: metricTagTypeMap[row.metric_type as MetricType],
      bordered: false,
      round: true
    }, { default: () => metricLabelMap[row.metric_type as MetricType] })
  },
  {
    title: '触发值',
    key: 'trigger_value',
    width: 100,
    render: (row: any) => h('span', { class: 'mono history-value' }, `${Number(row.trigger_value).toFixed(1)}${metricUnitMap[row.metric_type as MetricType]}`)
  },
  {
    title: '阈值',
    key: 'threshold',
    width: 90,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, `${row.threshold}${metricUnitMap[row.metric_type as MetricType]}`)
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const cfg = historyStatusMap[row.status as AlertStatus]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '通知状态',
    key: 'notification_status',
    width: 100,
    render: (row: any) => {
      const cfg = notificationStatusMap[row.notification_status as NotificationStatus]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '恢复时间',
    key: 'recovered_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.recovered_at || '—')
  }
])

// 展开行内容（notification_error）
function renderExpandedRow(row: any) {
  if (expandedHistoryId.value !== row.id || !row.notification_error) return null
  return h('div', { class: 'history-expand' }, [
    h('span', { class: 'history-expand__label' }, '通知错误：'),
    h('span', { class: 'history-expand__error mono' }, row.notification_error)
  ])
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadChannels()
  loadRules()
  loadChannelsList()
  loadHistory()
})
</script>

<template>
  <NTabs type="line" animated>
    <!-- ==================== Tab 1: 告警规则 ==================== -->
    <NTabPane name="rules" tab="告警规则">
      <PageCard title="告警规则" subtitle="配置服务器资源指标的告警触发条件">
        <template #extra>
          <NButton type="primary" @click="openRuleCreate">
            <template #icon>
              <NIcon><AddOutline /></NIcon>
            </template>
            创建规则
          </NButton>
        </template>

        <DataTable
          :columns="ruleColumns"
          :data="rules"
          :loading="rulesLoading"
          :pagination="rulesPagination"
        />
      </PageCard>
    </NTabPane>

    <!-- ==================== Tab 2: 通知渠道 ==================== -->
    <NTabPane name="channels" tab="通知渠道">
      <PageCard title="通知渠道" subtitle="管理告警消息的发送渠道（邮件 / 钉钉 / 企微 / 飞书 / Telegram / Webhook）">
        <template #extra>
          <NButton type="primary" @click="openChannelCreate">
            <template #icon>
              <NIcon><AddOutline /></NIcon>
            </template>
            添加渠道
          </NButton>
        </template>

        <DataTable
          :columns="channelColumns"
          :data="channels"
          :loading="channelsLoading"
          :pagination="channelsPagination"
        />
      </PageCard>
    </NTabPane>

    <!-- ==================== Tab 3: 告警历史 ==================== -->
    <NTabPane name="history" tab="告警历史">
      <PageCard title="告警历史" subtitle="查看所有告警触发与恢复记录">
        <!-- 筛选栏 -->
        <div class="history-filter">
          <NSpace size="medium" align="center">
            <div class="history-filter__item">
              <span class="history-filter__label">规则</span>
              <NSelect
                v-model:value="historyFilterRule"
                :options="ruleOptions"
                placeholder="全部规则"
                clearable
                style="width: 200px"
                @update:value="handleHistoryFilter"
              />
            </div>
            <div class="history-filter__item">
              <span class="history-filter__label">状态</span>
              <NSelect
                v-model:value="historyFilterStatus"
                :options="historyStatusOptions"
                placeholder="全部状态"
                clearable
                style="width: 140px"
                @update:value="handleHistoryFilter"
              />
            </div>
            <NButton size="small" @click="handleHistoryReset">重置</NButton>
          </NSpace>
        </div>

        <DataTable
          :columns="historyColumns"
          :data="history"
          :loading="historyLoading"
          :pagination="historyPagination"
        >
          <template #empty>
            <div class="history-empty">
              <p>暂无告警记录</p>
              <span>当配置的规则被触发时，记录会显示在这里</span>
            </div>
          </template>
        </DataTable>

        <!-- 展开行错误信息 -->
        <div v-for="row in history" :key="`expand-${row.id}`">
          <div v-if="expandedHistoryId === row.id && row.notification_error" class="history-expand-row">
            <span class="history-expand__label">通知错误：</span>
            <span class="history-expand__error mono">{{ row.notification_error }}</span>
          </div>
        </div>
      </PageCard>
    </NTabPane>
  </NTabs>

  <!-- ==================== 规则创建/编辑对话框 ==================== -->
  <ModalForm
    :show="showRuleForm"
    :title="ruleIsEdit ? '编辑规则' : '创建规则'"
    :loading="ruleFormLoading"
    :width="560"
    @update:show="(v: boolean) => (showRuleForm = v)"
    @submit="handleRuleFormSubmit"
  >
    <NForm ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-placement="top">
      <NFormItem label="规则名称" path="name">
        <NInput v-model:value="ruleForm.name" placeholder="例如：CPU 使用率过高" clearable />
      </NFormItem>

      <NSpace size="medium" align="end" style="width: 100%">
        <NFormItem label="指标类型" path="metric_type" style="flex: 1">
          <NSelect v-model:value="ruleForm.metric_type" :options="metricOptions" />
        </NFormItem>
        <NFormItem label="条件" path="condition" style="flex: 1">
          <NSelect v-model:value="ruleForm.condition" :options="conditionOptions" />
        </NFormItem>
        <NFormItem label="阈值" path="threshold" style="flex: 1">
          <NInput
            v-model:value="ruleForm.threshold"
            :placeholder="`单位 ${metricUnitMap[ruleForm.metric_type]}`"
          />
        </NFormItem>
      </NSpace>

      <NFormItem label="持续时间（分钟）" path="duration">
        <NInput v-model:value="ruleForm.duration" placeholder="默认 5 分钟" />
      </NFormItem>

      <NFormItem label="通知渠道">
        <NCheckboxGroup v-model:value="ruleForm.channels">
          <NSpace size="medium" wrap>
            <NCheckbox v-for="ch in channelList" :key="ch.id" :value="String(ch.id)">
              {{ ch.name }}
              <NTag size="tiny" type="default" :bordered="false" round style="margin-left: 4px">
                {{ channelLabelMap[ch.type] }}
              </NTag>
            </NCheckbox>
          </NSpace>
        </NCheckboxGroup>
        <div v-if="channelList.length === 0" class="form-hint">
          暂无可用渠道，请先在「通知渠道」中添加
        </div>
      </NFormItem>

      <NFormItem label="备注" path="remark">
        <NInput v-model:value="ruleForm.remark" placeholder="可选备注信息" clearable />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- ==================== 渠道创建/编辑对话框 ==================== -->
  <ModalForm
    :show="showChannelForm"
    :title="channelIsEdit ? '编辑渠道' : '添加渠道'"
    :loading="channelFormLoading"
    :width="560"
    @update:show="(v: boolean) => (showChannelForm = v)"
    @submit="handleChannelFormSubmit"
  >
    <NForm ref="channelFormRef" :model="channelForm" :rules="channelFormRules" label-placement="top">
      <NFormItem label="渠道名称" path="name">
        <NInput v-model:value="channelForm.name" placeholder="例如：运维告警邮件" clearable />
      </NFormItem>

      <NFormItem label="渠道类型" path="type">
        <NSelect v-model:value="channelForm.type" :options="channelTypeOptions" />
      </NFormItem>

      <!-- 邮件 -->
      <template v-if="channelForm.type === 'email'">
        <NFormItem label="收件人（多个邮箱用逗号分隔）">
          <NInput
            v-model:value="channelForm.email_recipients"
            type="textarea"
            placeholder="admin@example.com, ops@example.com"
            :rows="2"
          />
        </NFormItem>
        <div class="form-hint">SMTP 服务器配置在系统设置中管理</div>
      </template>

      <!-- 钉钉 -->
      <template v-if="channelForm.type === 'dingtalk'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.dingtalk_webhook" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
        </NFormItem>
        <NFormItem label="加签密钥（可选）">
          <NInput v-model:value="channelForm.dingtalk_secret" placeholder="SEC..." />
        </NFormItem>
      </template>

      <!-- 企业微信 -->
      <template v-if="channelForm.type === 'wechat'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.wechat_webhook" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." />
        </NFormItem>
      </template>

      <!-- 飞书 -->
      <template v-if="channelForm.type === 'feishu'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.feishu_webhook" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." />
        </NFormItem>
      </template>

      <!-- Telegram -->
      <template v-if="channelForm.type === 'telegram'">
        <NFormItem label="Bot Token">
          <NInput v-model:value="channelForm.telegram_bot_token" type="password" show-password-on="click" placeholder="123456:ABC-DEF..." />
        </NFormItem>
        <NFormItem label="Chat ID">
          <NInput v-model:value="channelForm.telegram_chat_id" placeholder="@channelname 或数字 ID" />
        </NFormItem>
      </template>

      <!-- Webhook -->
      <template v-if="channelForm.type === 'webhook'">
        <NFormItem label="URL">
          <NInput v-model:value="channelForm.webhook_url" placeholder="https://example.com/webhook" />
        </NFormItem>
        <NFormItem label="请求头（JSON 格式，可选）">
          <NInput
            v-model:value="channelForm.webhook_headers"
            type="textarea"
            placeholder='{"Authorization": "Bearer xxx", "X-Custom": "value"}'
            :rows="3"
          />
        </NFormItem>
      </template>

      <NFormItem label="启用">
        <NSwitch v-model:value="channelForm.enabled" />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- ==================== 规则删除确认 ==================== -->
  <ConfirmDialog
    :show="showRuleDelete"
    title="删除规则"
    :content="`确定要删除规则「${ruleDeleteName}」吗？此操作不可撤销。`"
    confirm-text="删除"
    type="error"
    :loading="ruleDeleteLoading"
    @update:show="(v: boolean) => (showRuleDelete = v)"
    @confirm="handleRuleDeleteConfirm"
  />

  <!-- ==================== 渠道删除确认 ==================== -->
  <ConfirmDialog
    :show="showChannelDelete"
    title="删除渠道"
    :content="`确定要删除渠道「${channelDeleteName}」吗？关联的规则将无法通过此渠道发送通知。`"
    confirm-text="删除"
    type="error"
    :loading="channelDeleteLoading"
    @update:show="(v: boolean) => (showChannelDelete = v)"
    @confirm="handleChannelDeleteConfirm"
  />
</template>

<style scoped>
/* 规则名称 */
.rule-name {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.rule-name__main {
  font-weight: 500;
  color: var(--ink);
}

.rule-name__remark {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.rule-condition {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-mono);
}

/* 渠道名称 */
.channel-name {
  font-weight: 500;
  color: var(--ink);
}

/* 历史筛选 */
.history-filter {
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
}

.history-filter__item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.history-filter__label {
  font-size: var(--fs-caption);
  color: var(--sub);
  white-space: nowrap;
}

/* 历史值 */
.history-value {
  font-weight: 600;
  color: var(--ink);
  font-variant-numeric: tabular-nums;
}

.history-rule-name {
  font-weight: 500;
  color: var(--ink);
}

/* 展开行 */
.history-expand-row {
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-left: 3px solid var(--err);
  margin: 0 var(--space-1) var(--space-2);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  display: flex;
  gap: var(--space-2);
  align-items: flex-start;
}

.history-expand__label {
  font-size: var(--fs-caption);
  color: var(--err);
  font-weight: 600;
  flex-shrink: 0;
}

.history-expand__error {
  font-size: var(--fs-caption);
  color: var(--sub);
  word-break: break-all;
}

/* 空态 */
.history-empty {
  padding: var(--space-8) 0;
  text-align: center;
}

.history-empty p {
  color: var(--sub);
  font-size: var(--fs-body);
  margin: 0 0 var(--space-2);
}

.history-empty span {
  color: var(--faint);
  font-size: var(--fs-caption);
}

/* 表单提示 */
.form-hint {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: calc(var(--space-2) * -1);
  margin-bottom: var(--space-4);
  padding-left: 2px;
}
</style>
