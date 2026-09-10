<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NEmpty, NTag, NInput, NSelect,
  NFormItem, NSwitch, NTabs, NTabPane
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  AddOutline, RefreshOutline, PencilOutline, PowerOutline, TrashOutline
} from '@vicons/ionicons5'
import { PageCard, DataTable, StatusTag, ModalForm, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import * as appstoreApi from '@/api/appstore'
import type {
  App, AppSource, AppSourceType, AppSourceInput
} from '@/api/appstore'
import AppStoreTab from './AppStoreTab.vue'
import InstalledTab from './InstalledTab.vue'
import InstallDialog from './InstallDialog.vue'
import {
  monoEl, iconBtn, btnGroup, formatTime,
  sourceTypeLabel, sourceTypeTagType
} from './shared'

const message = useMessage()

/* ==================== Tab 状态 ==================== */

const activeTab = ref<string>('store')

/* ==================== 源列表（共享：筛选 + 标签 + 源管理） ==================== */

const sources = ref<AppSource[]>([])
const sourcesLoading = ref(false)

async function loadSources() {
  sourcesLoading.value = true
  try {
    const res = await appstoreApi.getAppSources()
    sources.value = Array.isArray(res.data) ? res.data : []
  } catch (e: any) {
    message.error(e.message || '加载源列表失败')
    sources.value = []
  } finally {
    sourcesLoading.value = false
  }
}

/* ==================== 安装对话框 ==================== */

const installDialogVisible = ref(false)
const installingApp = ref<App | null>(null)
const appStoreTabRef = ref<InstanceType<typeof AppStoreTab> | null>(null)
const installedTabRef = ref<InstanceType<typeof InstalledTab> | null>(null)

function handleInstall(app: App) {
  installingApp.value = app
  installDialogVisible.value = true
}

function handleInstallSuccess() {
  installedTabRef.value?.loadInstalledApps()
}

/* ==================== 源管理 ==================== */

const syncingId = ref<string | null>(null)

/* --- 添加/编辑源对话框 --- */

const sourceModalVisible = ref(false)
const sourceModalLoading = ref(false)
const editingSource = ref<AppSource | null>(null)
const sourceForm = ref<AppSourceInput>({
  name: '',
  type: 'community',
  url: '',
  path: '',
  enabled: true
})

const sourceTypeOptions = [
  { label: '官方', value: 'official' },
  { label: '社区', value: 'community' },
  { label: '私有', value: 'private' }
]

const hasOfficialSource = computed(() =>
  sources.value.some((s) => s.type === 'official' && s.id !== editingSource.value?.id)
)

function openAddSource() {
  editingSource.value = null
  sourceForm.value = { name: '', type: 'community', url: '', path: '', enabled: true }
  sourceModalVisible.value = true
}

function openEditSource(row: AppSource) {
  editingSource.value = row
  sourceForm.value = {
    name: row.name,
    type: row.type,
    url: row.url || '',
    path: row.path || '',
    enabled: row.enabled
  }
  sourceModalVisible.value = true
}

async function submitSource() {
  if (!sourceForm.value.name.trim()) {
    message.warning('请输入源名称')
    return
  }
  if (sourceForm.value.type === 'community' && !sourceForm.value.url?.trim()) {
    message.warning('请输入社区源 URL')
    return
  }
  if (sourceForm.value.type === 'private' && !sourceForm.value.path?.trim()) {
    message.warning('请输入私有源本地路径')
    return
  }
  if (sourceForm.value.type === 'official' && hasOfficialSource.value) {
    message.warning('官方源只能存在一个，请先修改或删除现有官方源')
    return
  }

  sourceModalLoading.value = true
  try {
    const data: AppSourceInput = {
      name: sourceForm.value.name.trim(),
      type: sourceForm.value.type,
      enabled: sourceForm.value.enabled
    }
    if (sourceForm.value.type === 'community') data.url = sourceForm.value.url?.trim()
    if (sourceForm.value.type === 'private') data.path = sourceForm.value.path?.trim()

    if (editingSource.value) {
      await appstoreApi.updateAppSource(editingSource.value.id, data)
      message.success('源已更新')
    } else {
      await appstoreApi.createAppSource(data)
      message.success('源已添加')
    }
    sourceModalVisible.value = false
    loadSources()
  } catch (e: any) {
    message.error(e.message || '操作失败')
  } finally {
    sourceModalLoading.value = false
  }
}

/* --- 启用/禁用切换 --- */

async function toggleSource(row: AppSource) {
  try {
    await appstoreApi.toggleAppSource(row.id)
    message.success(row.enabled ? `已禁用「${row.name}」` : `已启用「${row.name}」`)
    loadSources()
  } catch (e: any) {
    message.error(e.message || '操作失败')
  }
}

/* --- 同步 --- */

async function syncSource(row: AppSource) {
  syncingId.value = row.id
  try {
    const res = await appstoreApi.syncAppSource(row.id)
    const r = res.data
    message.success(`同步完成：新增 ${r.added} / 更新 ${r.updated} / 删除 ${r.deleted}`)
    loadSources()
    if (activeTab.value === 'store') appStoreTabRef.value?.loadApps()
  } catch (e: any) {
    message.error(e.message || '同步失败')
  } finally {
    syncingId.value = null
  }
}

/* --- 删除确认 --- */

const deleteConfirm = ref(false)
const deleteLoading = ref(false)
const targetSource = ref<AppSource | null>(null)

function confirmDeleteSource(row: AppSource) {
  if (row.type === 'official') {
    message.warning('官方源不可删除')
    return
  }
  targetSource.value = row
  deleteConfirm.value = true
}

async function doDeleteSource() {
  if (!targetSource.value) return
  deleteLoading.value = true
  try {
    await appstoreApi.deleteAppSource(targetSource.value.id)
    message.success(`「${targetSource.value.name}」已删除`)
    deleteConfirm.value = false
    loadSources()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

/* --- 源列表 DataTable 列 --- */

const sourceColumns: DataTableColumns = [
  { title: '源名称', key: 'name', width: 180 },
  {
    title: '类型', key: 'type', width: 100,
    render: (row: any) =>
      h(
        NTag,
        {
          size: 'small',
          bordered: false,
          type: sourceTypeTagType(row.type as AppSourceType) as any
        },
        { default: () => sourceTypeLabel(row.type as AppSourceType) }
      )
  },
  {
    title: 'URL / 路径', key: 'url',
    render: (row: any) =>
      h(
        'span',
        { class: 'mono source-url', title: row.url || row.path || '' },
        row.url || row.path || '-'
      )
  },
  {
    title: '状态', key: 'enabled', width: 100,
    render: (row: any) =>
      h(StatusTag, {
        status: (row.enabled ? 'running' : 'stopped') as StatusType,
        text: row.enabled ? '启用' : '禁用'
      })
  },
  {
    title: '最后同步', key: 'last_sync_at', width: 170,
    render: (row: any) => formatTime(row.last_sync_at)
  },
  {
    title: '操作', key: 'actions', width: 280,
    render: (row: any) =>
      btnGroup([
        iconBtn({
          text: '同步', icon: RefreshOutline,
          loading: syncingId.value === row.id,
          onClick: () => syncSource(row)
        }),
        iconBtn({
          text: row.enabled ? '禁用' : '启用',
          icon: PowerOutline,
          type: row.enabled ? 'warning' : 'success',
          onClick: () => toggleSource(row)
        }),
        iconBtn({
          text: '编辑', icon: PencilOutline,
          onClick: () => openEditSource(row)
        }),
        iconBtn({
          text: '删除', icon: TrashOutline, type: 'error',
          disabled: row.type === 'official',
          onClick: () => confirmDeleteSource(row)
        })
      ])
  }
]

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadSources()
  // 默认显示商店 Tab，需要加载应用列表
  setTimeout(() => appStoreTabRef.value?.loadApps(), 0)
})

// nextTick：等新 tab 的面板挂载完成后再触发加载，避免 ref 为空被静默跳过
watch(activeTab, async (tab) => {
  await nextTick()
  if (tab === 'installed') {
    installedTabRef.value?.loadInstalledApps()
  }
  if (tab === 'sources') {
    loadSources()
  }
  if (tab === 'store') {
    appStoreTabRef.value?.loadApps()
  }
})
</script>

<template>
  <PageCard title="应用商店" subtitle="一键部署常用 Web 应用 · 容器化运行 · 环境参数可视化配置">
    <NTabs v-model:value="activeTab" type="line" animated>
      <!-- Tab 1: 应用商店 -->
      <NTabPane name="store" tab="应用商店">
        <AppStoreTab
          ref="appStoreTabRef"
          :sources="sources"
          :sources-loading="sourcesLoading"
          @install="handleInstall"
        />
      </NTabPane>

      <!-- Tab 2: 已安装 -->
      <NTabPane name="installed" tab="已安装">
        <InstalledTab
          ref="installedTabRef"
          :active="activeTab === 'installed'"
          @go-store="activeTab = 'store'"
        />
      </NTabPane>

      <!-- Tab 3: 源管理 -->
      <NTabPane name="sources" tab="源管理">
        <div class="tab-toolbar tab-toolbar--end">
          <NButton type="primary" @click="openAddSource">
            <template #icon><NIcon><AddOutline /></NIcon></template>添加源
          </NButton>
        </div>

        <DataTable
          :columns="sourceColumns"
          :data="sources"
          :loading="sourcesLoading"
          :pagination="false"
          :row-key="(row: any) => row.id"
        />
      </NTabPane>
    </NTabs>

    <!-- 安装对话框 -->
    <InstallDialog
      v-model:visible="installDialogVisible"
      :app="installingApp"
      @success="handleInstallSuccess"
    />

    <!-- 添加/编辑源对话框 -->
    <ModalForm
      v-model:show="sourceModalVisible"
      :title="editingSource ? '编辑源' : '添加源'"
      :loading="sourceModalLoading"
      :width="520"
      @submit="submitSource"
    >
      <NFormItem label="源名称" required>
        <NInput v-model:value="sourceForm.name" placeholder="请输入源名称，如：官方应用源" />
      </NFormItem>

      <NFormItem label="类型" required>
        <NSelect
          v-model:value="sourceForm.type"
          :options="sourceTypeOptions"
          placeholder="请选择源类型"
        />
      </NFormItem>

      <NFormItem v-if="sourceForm.type === 'community'" label="URL" required>
        <NInput v-model:value="sourceForm.url" placeholder="https://example.com/apps.json" />
      </NFormItem>

      <NFormItem v-if="sourceForm.type === 'private'" label="本地路径" required>
        <NInput v-model:value="sourceForm.path" placeholder="/path/to/apps/directory" />
      </NFormItem>

      <NFormItem label="启用">
        <NSwitch v-model:value="sourceForm.enabled" />
      </NFormItem>
    </ModalForm>

    <!-- 删除源确认 -->
    <ConfirmDialog
      v-model:show="deleteConfirm"
      title="删除源"
      :content="`确定要删除源「${targetSource?.name}」吗？该源下的应用将不再可用，此操作不可恢复。`"
      type="error"
      confirm-text="删除"
      :loading="deleteLoading"
      @confirm="doDeleteSource"
    />
  </PageCard>
</template>

<style scoped>
.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

.tab-toolbar {
  margin-bottom: var(--space-4);
  display: flex;
  align-items: center;
}

.tab-toolbar--end {
  justify-content: flex-end;
}

/* 源列表 URL 截断 */
.source-url {
  display: inline-block;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

/* NTabs 样式微调 */
:deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}

:deep(.n-tab) {
  font-size: var(--fs-body);
}
</style>
