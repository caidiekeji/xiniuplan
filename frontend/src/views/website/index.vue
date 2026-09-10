<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio,
  NDynamicTags, NDrawer, NDrawerContent, NModal, NTag, NSpace, NCheckbox,
  useMessage
} from 'naive-ui'
import {
  AddOutline, PlayOutline, StopOutline, RefreshOutline,
  SettingsOutline, DocumentTextOutline, TrashOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, SearchForm, ModalForm, StatusTag,
  ConfirmDialog, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import {
  getSites, createSite, deleteSite, startSite, stopSite, restartSite,
  getNginxConfig, updateNginxConfig, getSiteLogs,
  type Site, type SiteType, type SiteStatus, type SiteListParams, type CreateSiteData
} from '@/api/site'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()

/* ========== 列表状态 ========== */
const loading = ref(false)
const sites = ref<Site[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const searchDomain = ref('')
const filterType = ref<SiteType | null>(null)
const filterStatus = ref<SiteStatus | null>(null)

const typeOptions = [
  { label: '静态网站', value: 'static' },
  { label: '运行环境', value: 'runtime' },
  { label: '反向代理', value: 'reverse_proxy' }
]

const statusOptions = [
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' },
  { label: '异常', value: 'error' }
]

const typeLabelMap: Record<SiteType, string> = {
  static: '静态网站',
  runtime: '运行环境',
  reverse_proxy: '反向代理'
}

const typeTagTypeMap: Record<SiteType, 'info' | 'success' | 'warning'> = {
  static: 'info',
  runtime: 'success',
  reverse_proxy: 'warning'
}

const runtimeOptions = [
  { label: 'PHP 8.1', value: 'php-8.1' },
  { label: 'Node 18', value: 'node-18' },
  { label: 'Python 3.11', value: 'python-3.11' },
  { label: 'Go', value: 'go' }
]

/* ========== Mock 数据（后端未就绪时兜底） ========== */
function generateMockSites(): { items: Site[]; total: number } {
  const mock: Site[] = [
    {
      id: 1, domain: 'example.com', aliases: ['www.example.com'],
      type: 'static', rootDir: '/www/example.com', status: 'running',
      ssl: true, createdAt: '2026-08-15 10:30:00', remark: '公司官网'
    },
    {
      id: 2, domain: 'api.example.com', aliases: [],
      type: 'reverse_proxy', proxyTarget: 'http://127.0.0.1:3000',
      rootDir: '/www/api.example.com', status: 'running',
      ssl: true, createdAt: '2026-08-20 14:20:00', remark: 'API 网关'
    },
    {
      id: 3, domain: 'blog.example.com', aliases: ['m.example.com'],
      type: 'runtime', runtime: 'php-8.1',
      rootDir: '/www/blog.example.com', status: 'stopped',
      ssl: false, createdAt: '2026-09-01 09:15:00', remark: '博客站点'
    },
    {
      id: 4, domain: 'shop.example.com', aliases: [],
      type: 'runtime', runtime: 'node-18',
      rootDir: '/www/shop.example.com', status: 'error',
      ssl: true, createdAt: '2026-09-05 16:45:00', remark: '电商平台'
    }
  ]
  return { items: mock, total: mock.length }
}

async function loadSites() {
  loading.value = true
  try {
    const params: SiteListParams = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (searchDomain.value) params.domain = searchDomain.value
    if (filterType.value) params.type = filterType.value
    if (filterStatus.value) params.status = filterStatus.value

    const res = await getSites(params)
    if (res.data?.items) {
      sites.value = res.data.items
      total.value = res.data.total
    } else {
      const mock = generateMockSites()
      sites.value = mock.items
      total.value = mock.total
    }
  } catch {
    const mock = generateMockSites()
    sites.value = mock.items
    total.value = mock.total
  } finally {
    loading.value = false
  }
}

function handleSearch(values: Record<string, any>) {
  searchDomain.value = values.domain || ''
  page.value = 1
  loadSites()
}

function handleReset() {
  searchDomain.value = ''
  filterType.value = null
  filterStatus.value = null
  page.value = 1
  loadSites()
}

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { page.value = p; loadSites() },
  'onUpdate:page-size': (ps: number) => { pageSize.value = ps; page.value = 1; loadSites() }
}))

/* ========== 表格列 ========== */
const columns = computed<DataTableColumns>(() => [
  {
    title: '域名',
    key: 'domain',
    width: 240,
    render: (row: any) => h('div', { class: 'site-domain' }, [
      h('span', { class: 'site-domain__main mono' }, row.domain),
      row.aliases.length > 0
        ? h('div', { class: 'site-domain__aliases' },
            row.aliases.map((a: any) => h(NTag, { size: 'tiny', bordered: false, style: 'margin-right:4px;margin-top:2px' }, { default: () => a }))
          )
        : null
    ])
  },
  {
    title: '类型',
    key: 'type',
    width: 110,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as keyof typeof typeTagTypeMap],
      bordered: false,
      round: true
    }, { default: () => typeLabelMap[row.type as keyof typeof typeLabelMap] })
  },
  {
    title: '运行环境',
    key: 'runtime',
    width: 160,
    render: (row: any) => {
      if (row.type === 'runtime' && row.runtime) {
        return h('span', { class: 'mono' }, row.runtime)
      }
      if (row.type === 'reverse_proxy' && row.proxyTarget) {
        return h('span', { class: 'mono', style: 'color:var(--sub)' }, row.proxyTarget)
      }
      return h('span', { style: 'color:var(--faint)' }, '—')
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => h(StatusTag, { status: row.status as StatusType })
  },
  {
    title: 'SSL',
    key: 'ssl',
    width: 90,
    render: (row: any) => row.ssl
      ? h(StatusTag, { status: 'running', text: '已启用' })
      : h(StatusTag, { status: 'stopped', text: '未启用' })
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.createdAt)
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row: any) => {
      const isRunning = row.status === 'running'
      const actionLoading = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: isRunning ? 'warning' : 'success',
          loading: actionLoading,
          onClick: () => isRunning ? handleStop(row.id) : handleStart(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(isRunning ? StopOutline : PlayOutline) }),
          default: () => isRunning ? '停止' : '启动'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          loading: actionLoading,
          onClick: () => handleRestart(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
          default: () => '重启'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => openConfig(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(SettingsOutline) }),
          default: () => '配置'
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
          type: 'error',
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

/* ========== 启停操作 ========== */
const actionLoadingId = ref<number | null>(null)

async function handleStart(id: number) {
  actionLoadingId.value = id
  try {
    await startSite(id)
    message.success('站点已启动')
    loadSites()
  } catch (err: any) {
    message.error(err.message || '启动失败')
  } finally {
    actionLoadingId.value = null
  }
}

async function handleStop(id: number) {
  actionLoadingId.value = id
  try {
    await stopSite(id)
    message.success('站点已停止')
    loadSites()
  } catch (err: any) {
    message.error(err.message || '停止失败')
  } finally {
    actionLoadingId.value = null
  }
}

async function handleRestart(id: number) {
  actionLoadingId.value = id
  try {
    await restartSite(id)
    message.success('站点已重启')
    loadSites()
  } catch (err: any) {
    message.error(err.message || '重启失败')
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 创建网站对话框 ========== */
const showCreate = ref(false)
const createLoading = ref(false)
const createFormRef = ref()
const createForm = reactive<CreateSiteData>({
  domain: '',
  aliases: [],
  type: 'static',
  runtime: undefined,
  proxyTarget: undefined,
  rootDir: '',
  remark: ''
})

const createRules = {
  domain: {
    required: true,
    message: '请输入主域名',
    trigger: 'blur'
  },
  rootDir: {
    required: true,
    message: '请输入网站根目录',
    trigger: 'blur'
  },
  runtime: {
    required: true,
    message: '请选择运行环境',
    trigger: 'change'
  },
  proxyTarget: {
    required: true,
    message: '请输入反向代理目标',
    trigger: 'blur'
  }
}

function openCreate() {
  createForm.domain = ''
  createForm.aliases = []
  createForm.type = 'static'
  createForm.runtime = undefined
  createForm.proxyTarget = undefined
  createForm.rootDir = ''
  createForm.remark = ''
  showCreate.value = true
}

function handleDomainInput() {
  if (createForm.domain && !createForm.rootDir) {
    createForm.rootDir = `/www/${createForm.domain}`
  }
}

async function handleCreateSubmit() {
  try {
    await createFormRef.value?.validate()
  } catch {
    return
  }

  const data: CreateSiteData = {
    domain: createForm.domain,
    aliases: createForm.aliases,
    type: createForm.type,
    rootDir: createForm.rootDir,
    remark: createForm.remark
  }
  if (createForm.type === 'runtime') data.runtime = createForm.runtime
  if (createForm.type === 'reverse_proxy') data.proxyTarget = createForm.proxyTarget

  createLoading.value = true
  try {
    await createSite(data)
    message.success('网站创建成功')
    showCreate.value = false
    loadSites()
  } catch (err: any) {
    message.error(err.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

/* ========== Nginx 配置抽屉 ========== */
const showConfig = ref(false)
const configLoading = ref(false)
const configSaving = ref(false)
const configSiteId = ref<number | null>(null)
const configSiteDomain = ref('')
const nginxConfig = ref('')

async function openConfig(id: number) {
  configSiteId.value = id
  const site = sites.value.find((s) => s.id === id)
  configSiteDomain.value = site?.domain || ''
  showConfig.value = true
  configLoading.value = true
  try {
    const res = await getNginxConfig(id)
    nginxConfig.value = res.data?.content || generateMockNginxConfig(configSiteDomain.value)
  } catch {
    nginxConfig.value = generateMockNginxConfig(configSiteDomain.value)
  } finally {
    configLoading.value = false
  }
}

function generateMockNginxConfig(domain: string): string {
  return `server {
    listen 80;
    server_name ${domain || 'example.com'};
    root /www/${domain || 'example.com'};
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }

    location ~ /\\. {
        deny all;
    }

    access_log /var/log/nginx/${domain || 'example'}_access.log;
    error_log /var/log/nginx/${domain || 'example'}_error.log;
}
`
}

async function saveConfig() {
  if (configSiteId.value === null) return
  configSaving.value = true
  try {
    await updateNginxConfig(configSiteId.value, nginxConfig.value)
    message.success('Nginx 配置已保存')
    showConfig.value = false
  } catch (err: any) {
    message.error(err.message || '保存失败')
  } finally {
    configSaving.value = false
  }
}

/* ========== 日志对话框 ========== */
const showLogs = ref(false)
const logsLoading = ref(false)
const logsSiteId = ref<number | null>(null)
const logsSiteDomain = ref('')
const logsType = ref<'access' | 'error'>('access')
const logsContent = ref('')

const logTypeOptions = [
  { label: '访问日志 (access)', value: 'access' },
  { label: '错误日志 (error)', value: 'error' }
]

async function openLogs(id: number) {
  logsSiteId.value = id
  const site = sites.value.find((s) => s.id === id)
  logsSiteDomain.value = site?.domain || ''
  logsType.value = 'access'
  showLogs.value = true
  await loadLogs()
}

async function loadLogs() {
  if (logsSiteId.value === null) return
  logsLoading.value = true
  try {
    const res = await getSiteLogs(logsSiteId.value, logsType.value, 200)
    logsContent.value = res.data?.content || generateMockLogs(logsType.value)
  } catch {
    logsContent.value = generateMockLogs(logsType.value)
  } finally {
    logsLoading.value = false
  }
}

function generateMockLogs(type: string): string {
  if (type === 'access') {
    return `192.168.1.100 - - [08/Sep/2026:10:30:15 +0800] "GET / HTTP/1.1" 200 5432 "-" "Mozilla/5.0"
192.168.1.101 - - [08/Sep/2026:10:30:16 +0800] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"
192.168.1.102 - - [08/Sep/2026:10:30:18 +0800] "POST /api/login HTTP/1.1" 200 567 "-" "curl/7.68.0"
192.168.1.103 - - [08/Sep/2026:10:30:20 +0800] "GET /static/css/main.css HTTP/1.1" 304 0 "https://example.com" "Mozilla/5.0"
192.168.1.104 - - [08/Sep/2026:10:30:22 +0800] "GET /favicon.ico HTTP/1.1" 404 153 "-" "Mozilla/5.0"
`
  }
  return `2026/09/08 10:25:30 [error] 1234#0: *5678 open() "/www/example.com/missing.html" failed (2: No such file or directory)
2026/09/08 10:26:15 [warn] 1234#0: *5679 using uninitialized "foo" variable
2026/09/08 10:27:00 [error] 1234#0: *5680 upstream timed out (110: Connection timed out)
`
}

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteSiteId = ref<number | null>(null)
const deleteSiteDomain = ref('')
const deleteFiles = ref(false)

function openDelete(id: number) {
  deleteSiteId.value = id
  const site = sites.value.find((s) => s.id === id)
  deleteSiteDomain.value = site?.domain || ''
  deleteFiles.value = false
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteSiteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteSite(deleteSiteId.value, deleteFiles.value)
    message.success('网站已删除')
    showDelete.value = false
    loadSites()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadSites()
})
</script>

<template>
  <PageCard title="网站管理" subtitle="Web 站点全生命周期管理">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        创建网站
      </NButton>
    </template>

    <!-- 搜索筛选 -->
    <SearchForm @search="handleSearch" @reset="handleReset">
      <NFormItem label="域名" path="domain">
        <NInput
          v-model:value="searchDomain"
          name="domain"
          placeholder="输入域名搜索"
          clearable
          style="width: var(--input-lg)"
        />
      </NFormItem>
      <NFormItem label="类型">
        <NSelect
          v-model:value="filterType"
          :options="typeOptions"
          placeholder="全部类型"
          clearable
          style="width: var(--input-md)"
        />
      </NFormItem>
      <NFormItem label="状态">
        <NSelect
          v-model:value="filterStatus"
          :options="statusOptions"
          placeholder="全部状态"
          clearable
          style="width: var(--input-sm)"
        />
      </NFormItem>
    </SearchForm>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="sites"
      :loading="loading"
      :pagination="pagination"
    />
  </PageCard>

  <!-- 创建网站对话框 -->
  <ModalForm
    :show="showCreate"
    title="创建网站"
    :loading="createLoading"
    :width="640"
    @update:show="(v: boolean) => (showCreate = v)"
    @submit="handleCreateSubmit"
  >
    <NForm ref="createFormRef" :model="createForm" :rules="createRules" label-placement="top">
      <NFormItem label="网站类型" path="type">
        <NRadioGroup v-model:value="createForm.type" class="type-radio-group">
          <NSpace size="medium">
            <NRadio value="static" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">静态网站</span>
                <span class="type-card__desc">HTML / CSS / JS</span>
              </div>
            </NRadio>
            <NRadio value="runtime" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">运行环境</span>
                <span class="type-card__desc">PHP / Node / Python</span>
              </div>
            </NRadio>
            <NRadio value="reverse_proxy" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">反向代理</span>
                <span class="type-card__desc">转发到后端服务</span>
              </div>
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <NFormItem label="主域名" path="domain">
        <NInput
          v-model:value="createForm.domain"
          placeholder="例如 example.com"
          @input="handleDomainInput"
        />
      </NFormItem>

      <NFormItem label="绑定域名" path="aliases">
        <NDynamicTags
          v-model:value="createForm.aliases"
          placeholder="输入副域名后回车添加"
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'runtime'" label="运行环境" path="runtime">
        <NSelect
          v-model:value="createForm.runtime"
          :options="runtimeOptions"
          placeholder="选择运行环境"
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'reverse_proxy'" label="反向代理目标" path="proxyTarget">
        <NInput
          v-model:value="createForm.proxyTarget"
          placeholder="例如 http://127.0.0.1:3000"
        />
      </NFormItem>

      <NFormItem label="网站根目录" path="rootDir">
        <NInput v-model:value="createForm.rootDir" placeholder="/www/{domain}" />
      </NFormItem>

      <NFormItem label="备注" path="remark">
        <NInput
          v-model:value="createForm.remark"
          type="textarea"
          placeholder="可选备注信息"
          :rows="2"
        />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- Nginx 配置抽屉 -->
  <NDrawer
    :show="showConfig"
    :width="720"
    placement="right"
    @update:show="(v: boolean) => (showConfig = v)"
  >
    <NDrawerContent
      :title="`Nginx 配置 — ${configSiteDomain}`"
      closable
    >
      <div class="config-drawer">
        <div v-if="configLoading" class="config-loading">加载中...</div>
        <CodeEditor
          v-else
          v-model="nginxConfig"
          language="nginx"
          height="calc(100vh - 200px)"
        />
        <div class="config-drawer__footer">
          <NSpace justify="end">
            <NButton @click="showConfig = false">取消</NButton>
            <NButton type="primary" :loading="configSaving" @click="saveConfig">
              保存配置
            </NButton>
          </NSpace>
        </div>
      </div>
    </NDrawerContent>
  </NDrawer>

  <!-- 日志对话框 -->
  <NModal
    :show="showLogs"
    :mask-closable="false"
    preset="card"
    :title="`站点日志 — ${logsSiteDomain}`"
    :style="{ width: 'var(--modal-width-lg)' }"
    @update:show="(v: boolean) => (showLogs = v)"
  >
    <div class="logs-modal">
      <div class="logs-modal__toolbar">
        <NSelect
          v-model:value="logsType"
          :options="logTypeOptions"
          style="width: var(--input-xl)"
          @update:value="loadLogs"
        />
        <NButton size="small" quaternary @click="loadLogs">
          <template #icon>
            <NIcon><RefreshOutline /></NIcon>
          </template>
          刷新
        </NButton>
      </div>
      <CodeEditor
        v-model="logsContent"
        language="log"
        :read-only="true"
        height="420px"
        :minimap="false"
      />
    </div>
  </NModal>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    title="删除网站"
    :content="`确定要删除网站「${deleteSiteDomain}」吗？此操作不可撤销。`"
    confirm-text="删除"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  >
    <div class="delete-extra">
      <NCheckbox v-model:checked="deleteFiles">
        同时删除网站根目录文件
      </NCheckbox>
    </div>
  </ConfirmDialog>
</template>

<style scoped>
.site-domain__main {
  font-weight: 500;
  color: var(--ink);
}

.site-domain__aliases {
  margin-top: 4px;
  display: flex;
  flex-wrap: wrap;
}

.type-radio-group {
  width: 100%;
}

.type-radio-card {
  margin: 0;
  padding: 0;
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  transition: border-color var(--t-fast) var(--ease);
}

.type-radio-card :deep(.n-radio) {
  display: none;
}

.type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-4) var(--space-5);
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background var(--t-fast) var(--ease);
}

.type-card__title {
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
}

.type-card__desc {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.type-radio-card:hover {
  border-color: var(--acc);
}

.config-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.config-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: var(--chart-height);
  color: var(--faint);
}

.config-drawer__footer {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line);
}

.logs-modal__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.delete-extra {
  margin-top: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid var(--line);
}
</style>
