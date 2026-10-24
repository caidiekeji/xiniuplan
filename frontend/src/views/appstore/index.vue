<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import { NButton, NIcon, NTag, NInput, NSelect, NFormItem, NSwitch, NTabs, NTabPane } from 'naive-ui'
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
import { iconBtn, btnGroup, formatTime, sourceTypeLabel, sourceTypeTagType } from './shared'

const { t } = useI18n()
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
    message.error(e.message || t('appstore.loadSourcesFailed'))
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

const sourceTypeOptions = computed(() => [
  { label: t('appstore.sourceTypes.official'), value: 'official' },
  { label: t('appstore.sourceTypes.community'), value: 'community' },
  { label: t('appstore.sourceTypes.private'), value: 'private' }
])

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
    message.warning(t('appstore.validate.nameRequired'))
    return
  }
  if (sourceForm.value.type === 'community' && !sourceForm.value.url?.trim()) {
    message.warning(t('appstore.validate.communityUrlRequired'))
    return
  }
  if (sourceForm.value.type === 'private' && !sourceForm.value.path?.trim()) {
    message.warning(t('appstore.validate.privatePathRequired'))
    return
  }
  if (sourceForm.value.type === 'official' && hasOfficialSource.value) {
    message.warning(t('appstore.validate.officialOnlyOnce'))
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
      message.success(t('appstore.messages.sourceUpdated'))
    } else {
      await appstoreApi.createAppSource(data)
      message.success(t('appstore.messages.sourceAdded'))
    }
    sourceModalVisible.value = false
    loadSources()
  } catch (e: any) {
    message.error(e.message || t('appstore.messages.actionFailed'))
  } finally {
    sourceModalLoading.value = false
  }
}

/* --- 启用/禁用切换 --- */

async function toggleSource(row: AppSource) {
  try {
    await appstoreApi.toggleAppSource(row.id)
    message.success(
      row.enabled
        ? t('appstore.messages.sourceDisabled', { name: row.name })
        : t('appstore.messages.sourceEnabled', { name: row.name })
    )
    loadSources()
  } catch (e: any) {
    message.error(e.message || t('appstore.messages.actionFailed'))
  }
}

/* --- 同步 --- */

async function syncSource(row: AppSource) {
  syncingId.value = row.id
  try {
    const res = await appstoreApi.syncAppSource(row.id)
    const r = res.data
    message.success(
      t('appstore.messages.syncDone', { added: r.added, updated: r.updated, deleted: r.deleted })
    )
    loadSources()
    if (activeTab.value === 'store') appStoreTabRef.value?.loadApps()
  } catch (e: any) {
    message.error(e.message || t('appstore.messages.syncFailed'))
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
    message.warning(t('appstore.messages.officialNoDelete'))
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
    message.success(t('appstore.messages.sourceDeleted', { name: targetSource.value.name }))
    deleteConfirm.value = false
    loadSources()
  } catch (e: any) {
    message.error(e.message || t('appstore.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* --- 源列表 DataTable 列 --- */

const sourceColumns = computed<DataTableColumns>(() => [
  { title: t('appstore.columns.name'), key: 'name', width: 180 },
  {
    title: t('appstore.columns.type'), key: 'type', width: 100,
    render: (row: any) =>
      h(
        NTag,
        {
          size: 'small',
          bordered: false,
          type: sourceTypeTagType(row.type as AppSourceType) as any
        },
        { default: () => sourceTypeLabel(t, row.type as AppSourceType) }
      )
  },
  {
    title: t('appstore.columns.url'), key: 'url',
    render: (row: any) =>
      h(
        'span',
        { class: 'mono source-url', title: row.url || row.path || '' },
        row.url || row.path || '-'
      )
  },
  {
    title: t('appstore.columns.status'), key: 'enabled', width: 100,
    render: (row: any) =>
      h(StatusTag, {
        status: (row.enabled ? 'running' : 'stopped') as StatusType,
        text: row.enabled ? t('appstore.status.enabled') : t('appstore.status.disabled')
      })
  },
  {
    title: t('appstore.columns.lastSync'), key: 'last_sync_at', width: 170,
    render: (row: any) => formatTime(row.last_sync_at)
  },
  {
    title: t('appstore.columns.actions'), key: 'actions', width: 280,
    render: (row: any) =>
      btnGroup([
        iconBtn({
          text: t('appstore.actions.sync'), icon: RefreshOutline,
          loading: syncingId.value === row.id,
          onClick: () => syncSource(row)
        }),
        iconBtn({
          text: row.enabled ? t('appstore.actions.disable') : t('appstore.actions.enable'),
          icon: PowerOutline,
          type: row.enabled ? 'warning' : 'success',
          onClick: () => toggleSource(row)
        }),
        iconBtn({
          text: t('appstore.actions.edit'), icon: PencilOutline,
          onClick: () => openEditSource(row)
        }),
        iconBtn({
          text: t('appstore.actions.delete'), icon: TrashOutline, type: 'error',
          disabled: row.type === 'official',
          onClick: () => confirmDeleteSource(row)
        })
      ])
  }
])

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
  <PageCard :title="t('appstore.pageTitle')" :subtitle="t('appstore.pageSubtitle')">
    <NTabs v-model:value="activeTab" type="line" animated>
      <!-- Tab 1: 应用商店 -->
      <NTabPane name="store" :tab="t('appstore.tabs.store')">
        <AppStoreTab
          ref="appStoreTabRef"
          :sources="sources"
          :sources-loading="sourcesLoading"
          @install="handleInstall"
        />
      </NTabPane>

      <!-- Tab 2: 已安装 -->
      <NTabPane name="installed" :tab="t('appstore.tabs.installed')">
        <InstalledTab
          ref="installedTabRef"
          :active="activeTab === 'installed'"
          @go-store="activeTab = 'store'"
        />
      </NTabPane>

      <!-- Tab 3: 源管理 -->
      <NTabPane name="sources" :tab="t('appstore.tabs.sources')">
        <div class="tab-toolbar tab-toolbar--end">
          <NButton type="primary" @click="openAddSource">
            <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('appstore.addSource') }}
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
      :title="editingSource ? t('appstore.form.editTitle') : t('appstore.form.addTitle')"
      :loading="sourceModalLoading"
      :width="520"
      @submit="submitSource"
    >
      <NFormItem :label="t('appstore.form.nameLabel')" required>
        <NInput v-model:value="sourceForm.name" :placeholder="t('appstore.form.namePlaceholder')" />
      </NFormItem>

      <NFormItem :label="t('appstore.form.typeLabel')" required>
        <NSelect
          v-model:value="sourceForm.type"
          :options="sourceTypeOptions"
          :placeholder="t('appstore.form.typePlaceholder')"
        />
      </NFormItem>

      <NFormItem v-if="sourceForm.type === 'community'" label="URL" required>
        <NInput v-model:value="sourceForm.url" placeholder="https://example.com/apps.json" />
      </NFormItem>

      <NFormItem v-if="sourceForm.type === 'private'" :label="t('appstore.form.pathLabel')" required>
        <NInput v-model:value="sourceForm.path" placeholder="/path/to/apps/directory" />
      </NFormItem>

      <NFormItem :label="t('appstore.form.enabledLabel')">
        <NSwitch v-model:value="sourceForm.enabled" />
      </NFormItem>
    </ModalForm>

    <!-- 删除源确认 -->
    <ConfirmDialog
      v-model:show="deleteConfirm"
      :title="t('appstore.deleteDialog.title')"
      :content="t('appstore.deleteDialog.content', { name: targetSource?.name })"
      type="error"
      :confirm-text="t('appstore.deleteDialog.confirm')"
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
