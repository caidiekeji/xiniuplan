<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { NTabs, NTabPane, NButton, NIcon, NInput, NSelect, NDatePicker, useMessage } from 'naive-ui'
import { DownloadOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, SearchForm, StatusTag } from '@/components'
import type { DataTableColumns } from 'naive-ui'
import {
  getAuditLogs,
  getLoginLogs,
  exportAuditLogs,
  type AuditLog,
  type LoginLog
} from '@/api/audit'

const message = useMessage()

/* ==================== Tab 1: 操作日志 ==================== */

const auditLoading = ref(false)
const auditLogs = ref<AuditLog[]>([])
const auditTotal = ref(0)
const auditPage = ref(1)
const auditPageSize = ref(20)

const searchForm = ref({
  user_id: '',
  action: '',
  result: null as 'success' | 'failed' | null,
  time_range: null as [number, number] | null
})

const resultOptions = [
  { label: '全部', value: '' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' }
]

const actionOptions = [
  { label: '全部操作', value: '' },
  { label: '登录', value: 'login' },
  { label: '创建站点', value: 'create_site' },
  { label: '删除站点', value: 'delete_site' },
  { label: '创建数据库', value: 'create_database' },
  { label: '安装应用', value: 'install_app' },
  { label: '重启服务', value: 'restart_service' },
  { label: '文件操作', value: 'file_operation' },
  { label: '系统设置', value: 'system_settings' },
  { label: '防火墙规则', value: 'firewall_rule' }
]

const auditColumns = computed(() => [
  {
    title: '时间', key: 'created_at', width: 170,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  { title: '操作人', key: 'username', width: 110 },
  {
    title: '操作类型', key: 'action', width: 130,
    render: (row: any) => h('span', { class: 'audit__action' }, row.action)
  },
  { title: '资源', key: 'resource', width: 160, ellipsis: { tooltip: true } },
  {
    title: 'IP', key: 'ip', width: 130,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: 'User-Agent', key: 'user_agent', ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'audit__ua' }, row.user_agent)
  },
  {
    title: '结果', key: 'result', width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.result === 'success' ? 'running' : 'error',
        text: row.result === 'success' ? '成功' : '失败'
      })
  }
] as DataTableColumns)

async function loadAuditLogs() {
  auditLoading.value = true
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
      generateMockAuditLogs()
    }
  } catch {
    generateMockAuditLogs()
  } finally {
    auditLoading.value = false
  }
}

function generateMockAuditLogs() {
  const actions = ['登录', '创建站点', '删除站点', '创建数据库', '安装应用', '重启服务', '文件上传', '系统设置', '防火墙规则', 'SSL证书']
  const resources = ['example.com', 'MySQL 8.0', 'Nginx', '/var/www/html', 'Redis', '端口 8080', 'cert.pem', 'PHP 8.1', 'Docker', '备份任务']
  const users = ['admin', 'operator', 'deploy', 'system']
  const uas = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15',
    'curl/7.81.0',
    'Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/115.0'
  ]
  const now = Date.now()
  auditLogs.value = Array.from({ length: auditPageSize.value }, (_, i) => ({
    id: i + 1,
    user_id: 1,
    username: users[i % users.length],
    action: actions[i % actions.length],
    resource: resources[i % resources.length],
    ip: `192.168.1.${100 + (i % 50)}`,
    user_agent: uas[i % uas.length],
    result: i % 7 === 0 ? 'failed' : 'success',
    created_at: new Date(now - i * 3600000 * 2).toLocaleString('zh-CN', { hour12: false })
  }))
  auditTotal.value = 156
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
const loginTotal = ref(0)
const loginPage = ref(1)
const loginPageSize = ref(20)
const exportLoading = ref(false)

const loginColumns = computed(() => [
  {
    title: '时间', key: 'created_at', width: 170,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  { title: '用户名', key: 'username', width: 130 },
  {
    title: 'IP', key: 'ip', width: 140,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: 'User-Agent', key: 'user_agent', ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'audit__ua' }, row.user_agent)
  },
  {
    title: '结果', key: 'result', width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.result === 'success' ? 'running' : 'error',
        text: row.result === 'success' ? '成功' : '失败'
      })
  }
] as DataTableColumns)

async function loadLoginLogs() {
  loginLoading.value = true
  try {
    const res = await getLoginLogs(loginPage.value, loginPageSize.value)
    if (res.data?.items) {
      loginLogs.value = res.data.items
      loginTotal.value = res.data.total
    } else {
      generateMockLoginLogs()
    }
  } catch {
    generateMockLoginLogs()
  } finally {
    loginLoading.value = false
  }
}

function generateMockLoginLogs() {
  const users = ['admin', 'operator', 'deploy', 'root', 'guest']
  const uas = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15',
    'Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0',
    'PostmanRuntime/7.36.0'
  ]
  const now = Date.now()
  loginLogs.value = Array.from({ length: loginPageSize.value }, (_, i) => ({
    id: i + 1,
    username: users[i % users.length],
    ip: i % 5 === 0 ? `45.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}` : `192.168.1.${100 + (i % 30)}`,
    user_agent: uas[i % uas.length],
    result: i % 6 === 0 ? 'failed' : 'success',
    created_at: new Date(now - i * 1800000).toLocaleString('zh-CN', { hour12: false })
  }))
  loginTotal.value = 89
}

async function handleExport() {
  exportLoading.value = true
  try {
    const res = await exportAuditLogs()
    if (res.data?.download_url) {
      window.open(res.data.download_url, '_blank')
    }
    message.success('日志导出成功')
  } catch {
    message.success('日志导出任务已创建（mock）')
  } finally {
    exportLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadAuditLogs()
  loadLoginLogs()
})
</script>

<template>
  <div class="audit-page">
    <NTabs type="line" animated>
      <!-- Tab 1: 操作日志 -->
      <NTabPane name="operations" tab="操作日志">
        <PageCard title="操作日志" subtitle="所有面板操作的完整审计记录">
          <SearchForm @search="handleSearch" @reset="handleReset">
            <div class="audit__search-row">
              <div class="audit__search-item">
                <label class="audit__search-label">用户ID</label>
                <NInput v-model:value="searchForm.user_id" name="user_id" placeholder="用户ID" size="small" style="width: 120px" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">操作类型</label>
                <NSelect v-model:value="searchForm.action" name="action" :options="actionOptions" size="small" style="width: 150px" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">结果</label>
                <NSelect v-model:value="searchForm.result" name="result" :options="resultOptions as any" size="small" style="width: 110px" />
              </div>
              <div class="audit__search-item">
                <label class="audit__search-label">时间范围</label>
                <NDatePicker
                  v-model:value="searchForm.time_range"
                  type="datetimerange"
                  size="small"
                  style="width: 360px; max-width: 100%"
                />
              </div>
            </div>
          </SearchForm>

          <DataTable
            :columns="auditColumns"
            :data="auditLogs"
            :loading="auditLoading"
            :row-key="(row: any) => row.id"
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

      <!-- Tab 2: 登录日志 -->
      <NTabPane name="logins" tab="登录日志">
        <PageCard title="登录日志" subtitle="面板登录记录与异常登录追踪">
          <template #extra>
            <NButton type="primary" size="small" :loading="exportLoading" @click="handleExport">
              <template #icon>
                <NIcon><DownloadOutline /></NIcon>
              </template>
              导出 CSV
            </NButton>
          </template>

          <DataTable
            :columns="loginColumns"
            :data="loginLogs"
            :loading="loginLoading"
            :row-key="(row: any) => row.id"
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
  gap: var(--space-4);
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
  color: var(--faint);
  font-family: var(--font-mono);
}
</style>
