<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio,
  NDynamicTags, NDrawer, NDrawerContent, NTag, NSpace, NCheckbox, NEmpty
  } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, PlayOutline, StopOutline, RefreshOutline,
  SettingsOutline, DocumentTextOutline, TrashOutline, CreateOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, SearchForm, ModalForm, ContentModal, StatusTag,
  ConfirmDialog, CodeEditor
} from '@/components'
import { useQueryState } from '@/composables/useQueryState'
import type { StatusType } from '@/components'
import {
  getSites, createSite, updateSite, deleteSite, startSite, stopSite, restartSite,
  getNginxConfig, updateNginxConfig, getSiteLogs,
  type Site, type SiteType, type SiteStatus, type SiteListParams, type CreateSiteData
} from '@/api/site'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const message = useMessage()

/* ========== 列表状态 ========== */
// 分页 / 筛选 URL 化（D19）：刷新与深链可还原现场
const loading = ref(false)
const sites = ref<Site[]>([])
const total = ref(0)
const page = useQueryState('page', 1)
const pageSize = ref(20)

const searchDomain = useQueryState('kw', '')
const filterType = useQueryState<SiteType | ''>('type', '')
const filterStatus = useQueryState<SiteStatus | ''>('status', '')

const typeOptions = computed(() => [
  { label: t('website.types.static'), value: 'static' },
  { label: t('website.types.runtime'), value: 'runtime' },
  { label: t('website.types.reverseProxy'), value: 'reverse_proxy' }
])

const statusOptions = computed(() => [
  { label: t('website.status.running'), value: 'running' },
  { label: t('website.status.stopped'), value: 'stopped' },
  { label: t('website.status.error'), value: 'error' }
])

const typeLabelMap = computed<Record<SiteType, string>>(() => ({
  static: t('website.types.static'),
  runtime: t('website.types.runtime'),
  reverse_proxy: t('website.types.reverseProxy')
}))

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
async function loadSites() {
  loading.value = true
  try {
    const params: SiteListParams = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (searchDomain.value) params.domain = searchDomain.value
    if (filterType.value) params.type = filterType.value as SiteType
    if (filterStatus.value) params.status = filterStatus.value as SiteStatus

    const res = await getSites(params)
    if (res.data?.items) {
      sites.value = res.data.items
      total.value = res.data.total
    } else {
      sites.value = []
      total.value = 0
    }
  } catch (e: any) {
    sites.value = []
    total.value = 0
    message.error(e.message || t('website.messages.loadListFailed'))
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
  filterType.value = ''
  filterStatus.value = ''
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
    title: t('website.columns.domain'),
    key: 'domain',
    width: 240,
    render: (row: any) => h('div', { class: 'site-domain' }, [
      h('span', { class: 'site-domain__main mono' }, row.domain),
      (row.aliases || []).length > 0
        ? h('div', { class: 'site-domain__aliases' },
            (row.aliases || []).map((a: any) => h(NTag, { size: 'tiny', bordered: false, style: 'margin-right:var(--space-1);margin-top:var(--space-0)' }, { default: () => a }))
          )
        : null
    ])
  },
  {
    title: t('website.columns.type'),
    key: 'type',
    width: 110,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as keyof typeof typeTagTypeMap],
      bordered: false,
      round: true
    }, { default: () => typeLabelMap.value[row.type as SiteType] })
  },
  {
    title: t('website.columns.runtime'),
    key: 'runtime',
    width: 160,
    render: (row: any) => {
      if (row.type === 'runtime' && row.runtime) {
        return h('span', { class: 'mono' }, row.runtime)
      }
      if (row.type === 'reverse_proxy' && row.proxyTarget) {
        return h('span', { class: 'mono', style: 'color:var(--sub)' }, row.proxyTarget)
      }
      return h('span', { style: 'color: var(--text-quaternary)' }, '—')
    }
  },
  {
    title: t('website.columns.status'),
    key: 'status',
    width: 100,
    render: (row: any) => h(StatusTag, { status: row.status as StatusType })
  },
  {
    title: 'SSL',
    key: 'ssl',
    width: 90,
    render: (row: any) => row.ssl
      ? h(StatusTag, { status: 'running', text: t('website.ssl.enabled') })
      : h(StatusTag, { status: 'stopped', text: t('website.ssl.disabled') })
  },
  {
    title: t('website.columns.createdAt'),
    key: 'createdAt',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.createdAt)
  },
  {
    title: t('website.columns.actions'),
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
          default: () => isRunning ? t('website.actions.stop') : t('website.actions.start')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => openEdit(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
          default: () => t('website.actions.edit')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          loading: actionLoading,
          onClick: () => handleRestart(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
          default: () => t('website.actions.restart')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => openConfig(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(SettingsOutline) }),
          default: () => t('website.actions.config')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => openLogs(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }),
          default: () => t('website.actions.logs')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'error',
          onClick: () => openDelete(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          default: () => t('website.actions.delete')
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
    message.success(t('website.messages.startSuccess'))
    loadSites()
  } catch (err: any) {
    message.error(err.message || t('website.messages.startFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleStop(id: number) {
  actionLoadingId.value = id
  try {
    await stopSite(id)
    message.success(t('website.messages.stopSuccess'))
    loadSites()
  } catch (err: any) {
    message.error(err.message || t('website.messages.stopFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleRestart(id: number) {
  actionLoadingId.value = id
  try {
    await restartSite(id)
    message.success(t('website.messages.restartSuccess'))
    loadSites()
  } catch (err: any) {
    message.error(err.message || t('website.messages.restartFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 创建/编辑网站对话框 ========== */
// 同一个弹窗承担创建与编辑两种模式：
// 编辑模式下域名不可改（后端会拒绝换主域名，域名变更应走删除重建），
// 其余字段（别名、类型、运行时、目录、备注）可改。
const showCreate = ref(false)
const createLoading = ref(false)
const createFormRef = ref()
const isEditMode = ref(false)
const editSiteId = ref<number | null>(null)
const createForm = reactive<CreateSiteData>({
  domain: '',
  aliases: [],
  type: 'static',
  runtime: undefined,
  proxyTarget: undefined,
  rootDir: '',
  remark: ''
})

const createRules = computed(() => ({
  domain: {
    required: true,
    message: t('website.rules.domainRequired'),
    trigger: 'blur'
  },
  rootDir: {
    required: true,
    message: t('website.rules.rootDirRequired'),
    trigger: 'blur'
  },
  runtime: {
    required: true,
    message: t('website.rules.runtimeRequired'),
    trigger: 'change'
  },
  proxyTarget: {
    required: true,
    message: t('website.rules.proxyTargetRequired'),
    trigger: 'blur'
  }
}))

function openCreate() {
  isEditMode.value = false
  editSiteId.value = null
  createForm.domain = ''
  createForm.aliases = []
  createForm.type = 'static'
  createForm.runtime = undefined
  createForm.proxyTarget = undefined
  createForm.rootDir = ''
  createForm.remark = ''
  showCreate.value = true
}

/** 打开编辑对话框：预填现有站点数据 */
function openEdit(row: any) {
  isEditMode.value = true
  editSiteId.value = row.id
  createForm.domain = row.domain || ''
  createForm.aliases = Array.isArray(row.aliases) ? [...row.aliases] : []
  createForm.type = (row.type || 'static') as SiteType
  createForm.runtime = row.runtime
  createForm.proxyTarget = row.proxyTarget
  createForm.rootDir = row.rootDir || ''
  createForm.remark = row.remark || ''
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
    if (isEditMode.value && editSiteId.value !== null) {
      await updateSite(editSiteId.value, data)
      message.success(t('website.messages.updateSuccess'))
    } else {
      await createSite(data)
      message.success(t('website.messages.createSuccess'))
    }
    showCreate.value = false
    loadSites()
  } catch (err: any) {
    message.error(err.message || t(isEditMode.value ? 'website.messages.updateFailed' : 'website.messages.createFailed'))
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
    nginxConfig.value = res.data?.content || ''
  } catch {
    nginxConfig.value = ''
    message.error(t('website.messages.readConfigFailed'))
  } finally {
    configLoading.value = false
  }
}

async function saveConfig() {
  if (configSiteId.value === null) return
  configSaving.value = true
  try {
    await updateNginxConfig(configSiteId.value, nginxConfig.value)
    message.success(t('website.messages.configSaved'))
    showConfig.value = false
  } catch (err: any) {
    message.error(err.message || t('website.messages.saveFailed'))
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

const logTypeOptions = computed(() => [
  { label: t('website.logTypes.access'), value: 'access' },
  { label: t('website.logTypes.error'), value: 'error' }
])

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
    logsContent.value = res.data?.content || ''
    if (!res.data?.content) message.error(t('website.messages.logsFailed'))
  } catch {
    logsContent.value = ''
    message.error(t('website.messages.logsFailed'))
  } finally {
    logsLoading.value = false
  }
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
    message.success(t('website.messages.deleteSuccess'))
    showDelete.value = false
    loadSites()
  } catch (err: any) {
    message.error(err.message || t('website.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadSites()
})
</script>

<template>
  <PageCard :title="t('website.title')" :subtitle="t('website.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('website.createSite') }}
      </NButton>
    </template>

    <!-- 搜索筛选 -->
    <SearchForm @search="handleSearch" @reset="handleReset">
      <NFormItem :label="t('website.search.domainLabel')" path="domain">
        <NInput
          v-model:value="searchDomain"
          name="domain"
          :placeholder="t('website.search.domainPlaceholder')"
          clearable
          style="width: var(--input-lg)"
        />
      </NFormItem>
      <NFormItem :label="t('website.search.typeLabel')">
        <NSelect
          v-model:value="filterType"
          :options="typeOptions"
          :placeholder="t('website.search.typePlaceholder')"
          clearable
          style="width: var(--input-md)"
        />
      </NFormItem>
      <NFormItem :label="t('website.search.statusLabel')">
        <NSelect
          v-model:value="filterStatus"
          :options="statusOptions"
          :placeholder="t('website.search.statusPlaceholder')"
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
    >
      <!-- 空态引导（词表 D15）：无筛选条件的首次空态给创建入口；筛选空结果只显示「暂无数据」 -->
      <template #empty>
        <NEmpty :description="t('common.noData')">
          <template v-if="!searchDomain && !filterType && !filterStatus" #extra>
            <NButton size="small" @click="openCreate">{{ t('common.create') }}</NButton>
          </template>
        </NEmpty>
      </template>
    </DataTable>
  </PageCard>

  <!-- 创建/编辑网站对话框 -->
  <ModalForm
    :show="showCreate"
    :title="isEditMode ? t('website.editModal.title') : t('website.createModal.title')"
    :loading="createLoading"
    :width="640"
    @update:show="(v: boolean) => (showCreate = v)"
    @submit="handleCreateSubmit"
  >
    <NForm ref="createFormRef" :model="createForm" :rules="createRules" label-placement="top">
      <NFormItem :label="t('website.createModal.typeLabel')" path="type">
        <NRadioGroup v-model:value="createForm.type" class="type-radio-group">
          <NSpace size="medium">
            <NRadio value="static" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">{{ t('website.types.static') }}</span>
                <span class="type-card__desc">HTML / CSS / JS</span>
              </div>
            </NRadio>
            <NRadio value="runtime" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">{{ t('website.types.runtime') }}</span>
                <span class="type-card__desc">PHP / Node / Python</span>
              </div>
            </NRadio>
            <NRadio value="reverse_proxy" class="type-radio-card">
              <div class="type-card">
                <span class="type-card__title">{{ t('website.types.reverseProxy') }}</span>
                <span class="type-card__desc">{{ t('website.createModal.reverseProxyDesc') }}</span>
              </div>
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <NFormItem :label="t('website.createModal.domainLabel')" path="domain">
        <NInput
          v-model:value="createForm.domain"
          :placeholder="t('website.createModal.domainPlaceholder')"
          :readonly="isEditMode"
          :disabled="isEditMode"
          @input="handleDomainInput"
        />
        <div v-if="isEditMode" class="form-hint">{{ t('website.editModal.domainLocked') }}</div>
      </NFormItem>

      <NFormItem :label="t('website.createModal.aliasesLabel')" path="aliases">
        <NDynamicTags
          v-model:value="createForm.aliases"
          :placeholder="t('website.createModal.aliasesPlaceholder')"
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'runtime'" :label="t('website.createModal.runtimeLabel')" path="runtime">
        <NSelect
          v-model:value="createForm.runtime"
          :options="runtimeOptions"
          :placeholder="t('website.createModal.runtimePlaceholder')"
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'reverse_proxy'" :label="t('website.createModal.proxyTargetLabel')" path="proxyTarget">
        <NInput
          v-model:value="createForm.proxyTarget"
          :placeholder="t('website.createModal.proxyTargetPlaceholder')"
        />
      </NFormItem>

      <NFormItem :label="t('website.createModal.rootDirLabel')" path="rootDir">
        <NInput v-model:value="createForm.rootDir" placeholder="/www/{domain}" />
      </NFormItem>

      <NFormItem :label="t('website.createModal.remarkLabel')" path="remark">
        <NInput
          v-model:value="createForm.remark"
          type="textarea"
          :placeholder="t('website.createModal.remarkPlaceholder')"
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
      :title="t('website.configDrawer.title', { domain: configSiteDomain })"
      closable
    >
      <div class="config-drawer">
        <div v-if="configLoading" class="config-loading">{{ t('website.configDrawer.loading') }}</div>
        <CodeEditor
          v-else
          v-model="nginxConfig"
          language="nginx"
          height="calc(100vh - 200px)"
        />
        <div class="config-drawer__footer">
          <NSpace justify="end">
            <NButton @click="showConfig = false">{{ t('website.configDrawer.cancel') }}</NButton>
            <NButton type="primary" :loading="configSaving" @click="saveConfig">
              {{ t('website.configDrawer.save') }}
            </NButton>
          </NSpace>
        </div>
      </div>
    </NDrawerContent>
  </NDrawer>

  <!-- 日志对话框 -->
  <ContentModal
    :show="showLogs"
    :mask-closable="false"
    :title="t('website.logsModal.title', { domain: logsSiteDomain })"
    size="lg"
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
          {{ t('website.logsModal.refresh') }}
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
  </ContentModal>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('website.deleteDialog.title')"
    :content="t('website.deleteDialog.content', { domain: deleteSiteDomain })"
    :confirm-text="t('website.actions.delete')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  >
    <div class="delete-extra">
      <NCheckbox v-model:checked="deleteFiles">
        {{ t('website.deleteDialog.deleteFiles') }}
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
  margin-top: var(--space-1);
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
  color: var(--text-quaternary);
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
  color: var(--text-quaternary);
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
