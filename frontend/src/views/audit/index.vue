<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import { NTabs, NTabPane, NButton, NIcon, NInput, NSelect, NDatePicker, NTag } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { DownloadOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, SearchForm, StatusTag } from '@/components'
import { listAnomalies, type AnomalyEvent } from '@/api/newfeatures'
import type { DataTableColumns } from 'naive-ui'
import {
  getAuditLogs,
  getLoginLogs,
  exportAuditLogs,
  type AuditLog,
  type LoginLog
} from '@/api/audit'
import { useI18n } from 'vue-i18n'
import { useQueryState } from '@/composables/useQueryState'
import { formatDate } from '@/utils/format'

const message = useMessage()
const { t } = useI18n()

/* ==================== URL 状态（D19：Tab/分页/筛选刷新可还原） ==================== */

const activeTab = useQueryState('tab', 'operations')

/* ==================== Tab 1: 操作日志 ==================== */

const auditLoading = ref(false)
/** 各表加载失败信息：非空时对应表格区渲染错误面板（含重试） */
const auditError = ref<string | null>(null)
const auditLogs = ref<AuditLog[]>([])
const auditTotal = ref(0)
const auditPage = useQueryState('page', 1)
const auditPageSize = ref(20)

const searchForm = ref({
  user_id: '',
  action: '',
  result: null as 'success' | 'failed' | null,
  time_range: null as [number, number] | null
})

// 筛选条件 URL 化：初始化从 query 恢复，变化时写回（time_range 为复杂对象，不入 URL）
const qUser = useQueryState('user', '')
const qAction = useQueryState('action', '')
const qResult = useQueryState('result', '')
searchForm.value.user_id = qUser.value
searchForm.value.action = qAction.value
searchForm.value.result = (qResult.value || null) as 'success' | 'failed' | null
watch(() => searchForm.value.user_id, (v) => (qUser.value = v || ''))
watch(() => searchForm.value.action, (v) => (qAction.value = v || ''))
watch(() => searchForm.value.result, (v) => (qResult.value = v || ''))

const resultOptions = computed(() => [
  { label: t('audit.resultOptions.all'), value: '' },
  { label: t('audit.resultOptions.success'), value: 'success' },
  { label: t('audit.resultOptions.failed'), value: 'failed' }
])

const actionOptions = computed(() => [
  { label: t('audit.actionOptions.all'), value: '' },
  { label: t('audit.actionOptions.login'), value: 'login' },
  { label: t('audit.actionOptions.createSite'), value: 'create_site' },
  { label: t('audit.actionOptions.deleteSite'), value: 'delete_site' },
  { label: t('audit.actionOptions.createDatabase'), value: 'create_database' },
  { label: t('audit.actionOptions.installApp'), value: 'install_app' },
  { label: t('audit.actionOptions.restartService'), value: 'restart_service' },
  { label: t('audit.actionOptions.fileOperation'), value: 'file_operation' },
  { label: t('audit.actionOptions.systemSettings'), value: 'system_settings' },
  { label: t('audit.actionOptions.firewallRule'), value: 'firewall_rule' }
])

const auditColumns = computed(() => [
  {
    title: t('audit.columns.time'), key: 'created_at', width: 170,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  { title: t('audit.columns.operator'), key: 'username', width: 110 },
  {
    title: t('audit.columns.actionType'), key: 'action', width: 130,
    render: (row: any) => h('span', { class: 'audit__action' }, row.action)
  },
  { title: t('audit.columns.resource'), key: 'resource', width: 160, ellipsis: { tooltip: true } },
  {
    title: 'IP', key: 'ip', width: 130,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: 'User-Agent', key: 'user_agent', ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'audit__ua' }, row.user_agent)
  },
  {
    title: t('audit.columns.result'), key: 'result', width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.result === 'success' ? 'running' : 'error',
        text: row.result === 'success' ? t('audit.resultOptions.success') : t('audit.resultOptions.failed')
      })
  }
] as DataTableColumns)

async function loadAuditLogs() {
  auditLoading.value = true
  auditError.value = null
  try {
    const params: any = {
      page: auditPage.value,
      pageSize: auditPageSize.value
    }
    if (searchForm.value.user_id) params.user_id = Number(searchForm.value.user_id)
    if (searchForm.value.action) params.action = searchForm.value.action
    if (searchForm.value.result) params.result = searchForm.value.result
    if (searchForm.value.time_range) {
      params.start_time = new Date(searchForm.value.time_range[0]).toISOString()
      params.end_time = new Date(searchForm.value.time_range[1]).toISOString()
    }
    const res = await getAuditLogs(params)
    if (res.data?.items) {
      auditLogs.value = res.data.items
      auditTotal.value = res.data.total
    } else {
      auditLogs.value = []
      auditTotal.value = 0
    }
  } catch (e: any) {
    auditLogs.value = []
    auditTotal.value = 0
    auditError.value = e.message || t('audit.messages.loadAuditFailed')
    message.error(e.message || t('audit.messages.loadAuditFailed'))
  } finally {
    auditLoading.value = false
  }
}

function handleSearch(values: Record<string, any>) {
  searchForm.value.user_id = values.user_id || ''
  searchForm.value.action = values.action || ''
  auditPage.value = 1
  loadAuditLogs()
}

function handleReset() {
  searchForm.value = { user_id: '', action: '', result: null, time_range: null }
  auditPage.value = 1
  loadAuditLogs()
}

/* ==================== Tab 2: 登录日志 ==================== */

const loginLoading = ref(false)
const loginLogs = ref<LoginLog[]>([])
const loginError = ref<string | null>(null)
const loginTotal = ref(0)
const loginPage = useQueryState('loginPage', 1)
const loginPageSize = ref(20)
const exportLoading = ref(false)

const loginColumns = computed(() => [
  {
    title: t('audit.columns.time'), key: 'created_at', width: 170,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  { title: t('audit.columns.username'), key: 'username', width: 130 },
  {
    title: 'IP', key: 'ip', width: 140,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: 'User-Agent', key: 'user_agent', ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'audit__ua' }, row.user_agent)
  },
  {
    title: t('audit.columns.result'), key: 'result', width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.result === 'success' ? 'running' : 'error',
        text: row.result === 'success' ? t('audit.resultOptions.success') : t('audit.resultOptions.failed')
      })
  }
] as DataTableColumns)

async function loadLoginLogs() {
  loginLoading.value = true
  loginError.value = null
  try {
    const res = await getLoginLogs(loginPage.value, loginPageSize.value)
    if (res.data?.items) {
      loginLogs.value = res.data.items
      loginTotal.value = res.data.total
    } else {
      loginLogs.value = []
      loginTotal.value = 0
    }
  } catch (e: any) {
    loginLogs.value = []
    loginTotal.value = 0
    loginError.value = e.message || t('audit.messages.loadLoginFailed')
    message.error(e.message || t('audit.messages.loadLoginFailed'))
  } finally {
    loginLoading.value = false
  }
}

async function handleExport() {
  exportLoading.value = true
  try {
    const res = await exportAuditLogs()
    if (res.data?.download_url) {
      window.open(res.data.download_url, '_blank')
    }
    message.success(t('audit.messages.exportSuccess'))
  } catch {
    message.success(t('audit.messages.exportMock'))
  } finally {
    exportLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadAuditLogs()
  loadLoginLogs()
})

/* ========== 异常登录检测（S-012） ========== */
const NTagComp = NTag
const anomalies = ref<AnomalyEvent[]>([])
const anomalyLoading = ref(false)
const anomalyError = ref<string | null>(null)

async function loadAnomalies() {
  anomalyLoading.value = true
  anomalyError.value = null
  try {
    const res = await listAnomalies()
    anomalies.value = res.items || []
  } catch (e: any) {
    anomalyError.value = e.message || t('audit.messages.detectFailed')
    message.error(e.message || t('audit.messages.detectFailed'))
  } finally {
    anomalyLoading.value = false
  }
}

const anomalyTypeMap = computed<Record<string, { label: string; type: 'error' | 'warning' | 'info' }>>(() => ({
  brute_force_ip: { label: t('audit.anomalyTypes.bruteForceIp'), type: 'error' },
  brute_force_user: { label: t('audit.anomalyTypes.bruteForceUser'), type: 'error' },
  new_ip_login: { label: t('audit.anomalyTypes.newIpLogin'), type: 'info' }
}))

const anomalyColumns = computed<DataTableColumns>(() => [
  {
    title: t('audit.columns.type'), key: 'type', width: 120,
    render: (row: any) => {
      const cfg = anomalyTypeMap.value[row.type] || { label: row.type, type: 'info' as const }
      return h(NTagComp, { size: 'small', type: cfg.type }, { default: () => cfg.label })
    }
  },
  { title: t('audit.columns.subject'), key: 'subject', width: 160 },
  { title: t('audit.columns.count'), key: 'count', width: 80 },
  { title: t('audit.columns.lastSeen'), key: 'last_seen', width: 180, render: (r: any) => formatDate(r.last_seen) },
  { title: t('audit.columns.detail'), key: 'detail', ellipsis: { tooltip: true } }
])

onMounted(loadAnomalies)
</script>

<template>
  <div class="audit-page">
    <NTabs v-model:value="activeTab" type="line" animated>
      <!-- Tab 1: 操作日志 -->
      <NTabPane name="operations" :tab="t('audit.tabs.operations')">
        <PageCard :title="t('audit.audit.title')" :subtitle="t('audit.audit.subtitle')">
          <SearchForm @search="handleSearch" @reset="handleReset">
            <div class="audit__search-row">
              <div class="audit__search-item">
                <label class="audit__search-label">{{ t('audit.audit.userId') }}</label>
                <NInput v-model:value="searchForm.user_id" name="user_id" :placeholder="t('audit.audit.userIdPlaceholder')" size="small" style="width: var(--input-sm)" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">{{ t('audit.audit.actionType') }}</label>
                <NSelect v-model:value="searchForm.action" name="action" :options="actionOptions" size="small" style="width: var(--input-md)" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">{{ t('audit.audit.result') }}</label>
                <NSelect v-model:value="searchForm.result" name="result" :options="resultOptions as any" size="small" style="width: var(--input-sm)" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">{{ t('audit.audit.timeRange') }}</label>
                <NDatePicker
                  v-model:value="searchForm.time_range"
                  type="datetimerange"
                  size="small"
                  style="width: var(--input-3xl); max-width: 100%"
                />
              </div>
            </div>
          </SearchForm>

          <DataTable
            :columns="auditColumns"
            :data="auditLogs"
            :loading="auditLoading"
            :row-key="(row: any) => row.id"
            :error="auditError"
            @retry="loadAuditLogs"
            :pagination="{
              page: auditPage,
              pageSize: auditPageSize,
              itemCount: auditTotal,
              'onUpdate:page': (p: number) => { auditPage = p; loadAuditLogs() },
              'onUpdate:pageSize': (s: number) => { auditPageSize = s; auditPage = 1; loadAuditLogs() }
            }"
          />
        </PageCard>
      </NTabPane>

      <!-- Tab: 异常检测 -->
      <NTabPane name="anomalies" :tab="t('audit.tabs.anomalies')">
        <PageCard :title="t('audit.anomalies.title')" :subtitle="t('audit.anomalies.subtitle')">
          <template #extra>
            <NButton size="small" quaternary :loading="anomalyLoading" @click="loadAnomalies">
              {{ t('audit.anomalies.refresh') }}
            </NButton>
          </template>
          <DataTable
            :columns="anomalyColumns"
            :data="anomalies"
            :loading="anomalyLoading"
            :row-key="(row: any) => row.type + row.subject"
            :error="anomalyError"
            @retry="loadAnomalies"
          />
        </PageCard>
      </NTabPane>

      <!-- Tab 2: 登录日志 -->
      <NTabPane name="logins" :tab="t('audit.tabs.logins')">
        <PageCard :title="t('audit.logins.title')" :subtitle="t('audit.logins.subtitle')">
          <template #extra>
            <NButton type="primary" size="small" :loading="exportLoading" @click="handleExport">
              <template #icon>
                <NIcon><DownloadOutline /></NIcon>
              </template>
              {{ t('audit.logins.exportCsv') }}
            </NButton>
          </template>

          <DataTable
            :columns="loginColumns"
            :data="loginLogs"
            :loading="loginLoading"
            :row-key="(row: any) => row.id"
            :error="loginError"
            @retry="loadLoginLogs"
            :pagination="{
              page: loginPage,
              pageSize: loginPageSize,
              itemCount: loginTotal,
              'onUpdate:page': (p: number) => { loginPage = p; loadLoginLogs() },
              'onUpdate:pageSize': (s: number) => { loginPageSize = s; loginPage = 1; loadLoginLogs() }
            }"
          />
        </PageCard>
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.audit-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.audit__search-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  align-items: flex-end;
}

.audit__search-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.audit__search-label {
  font-size: var(--fs-caption);
  color: var(--sub);
  font-weight: 500;
}

.audit__action {
  font-weight: 500;
  color: var(--ink);
}

.audit__ua {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  font-family: var(--font-mono);
}
</style>
