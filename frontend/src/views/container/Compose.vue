<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NCard, NSpace, NInput, NTabs, NTabPane
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import type { DataTableColumns } from 'naive-ui'
import { AddOutline, RefreshOutline, PlayOutline, StopOutline,  CreateOutline, TrashOutline, HammerOutline, CloseOutline, ReloadOutline } from '@vicons/ionicons5'
import { StatusTag, ConfirmDialog, ContentModal, DataTable } from '@/components'
import type { StatusType } from '@/components'
import {
  listComposeStacks, getComposeStack, saveComposeStack, deleteComposeStack, composeAction,
  buildImage, type ComposeStack
} from '@/api/features17'

const message = useMessage()
const { t } = useI18n()

/* 两个功能区以标签页组织 */
const activeTab = ref('stacks')

/* ==================== Compose 栈 ==================== */
const stacks = ref<ComposeStack[]>([])
const stackLoading = ref(false)
const actionName = ref<string | null>(null)

async function loadStacks() {
  stackLoading.value = true
  try {
    const res = await listComposeStacks()
    stacks.value = res.items || []
  } catch {
    stacks.value = []
  } finally {
    stackLoading.value = false
  }
}

function stackStatus(s: ComposeStack): { status: StatusType; text: string } {
  if (s.total > 0 && s.running === s.total) return { status: 'running', text: t('compose.status.running') }
  if (s.running === 0) return { status: 'stopped', text: t('compose.status.stopped') }
  return { status: 'warning', text: t('compose.status.partial') }
}

const stackColumns = computed<DataTableColumns<ComposeStack>>(() => [
  {
    title: t('compose.columns.name'),
    key: 'name',
    render: (row) => row.name
  },
  {
    title: t('compose.columns.containers'),
    key: 'running',
    width: 110,
    render: (row) => `${row.running} / ${row.total}`
  },
  {
    title: t('compose.columns.status'),
    key: 'status',
    width: 100,
    render: (row) => {
      const cfg = stackStatus(row)
      return StatusTagC(cfg)
    }
  },
  {
    title: t('compose.columns.actions'),
    key: 'actions',
    width: 260,
    render: (row) => {
      const busy = actionName.value === row.name
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy, disabled: busy,
            onClick: () => handleStackAction(row.name, 'up')
          }, { icon: () => h(NIcon, null, { default: () => h(PlayOutline) }), default: () => t('compose.actions.start') }),
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy, disabled: busy,
            onClick: () => handleStackAction(row.name, 'down')
          }, { icon: () => h(NIcon, null, { default: () => h(StopOutline) }), default: () => t('compose.actions.stop') }),
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy, disabled: busy,
            onClick: () => handleStackAction(row.name, 'restart')
          }, { icon: () => h(NIcon, null, { default: () => h(ReloadOutline) }), default: () => t('compose.actions.restart') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openEditStack(row.name)
          }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => t('compose.actions.edit') }),
          h(NButton, {
            size: 'tiny', quaternary: true, type: 'error',
            onClick: () => openDeleteStack(row)
          }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }), default: () => t('compose.actions.delete') })
        ]
      })
    }
  }
])

// 模板中使用 StatusTag 组件渲染状态
function StatusTagC(cfg: { status: StatusType; text: string }) {
  return h(StatusTag, { status: cfg.status, text: cfg.text })
}

async function handleStackAction(name: string, action: 'up' | 'down' | 'restart') {
  actionName.value = name
  try {
    await composeAction(name, action)
    message.success(t('compose.messages.actionDone'))
    loadStacks()
  } catch (err: any) {
    message.error(err.message || t('compose.messages.actionFailed'))
  } finally {
    actionName.value = null
  }
}

/* ==================== 栈编辑 ==================== */
const showEditor = ref(false)
const editorLoading = ref(false)
const editorName = ref('')
const editorNameDisabled = ref(false)
const editorContent = ref('')

async function openCreateStack() {
  editorName.value = ''
  editorNameDisabled.value = false
  editorContent.value = ''
  showEditor.value = true
}

async function openEditStack(name: string) {
  editorName.value = name
  editorNameDisabled.value = true
  editorContent.value = ''
  showEditor.value = true
  try {
    const detail = await getComposeStack(name)
    editorContent.value = detail.content || ''
  } catch (err: any) {
    message.error(err.message || t('compose.messages.loadContentFailed'))
  }
}

async function handleSaveStack() {
  const name = editorName.value.trim()
  if (!name) {
    message.error(t('compose.rules.nameRequired'))
    return
  }
  editorLoading.value = true
  try {
    await saveComposeStack(name, editorContent.value)
    message.success(t('compose.messages.saved'))
    showEditor.value = false
    loadStacks()
  } catch (err: any) {
    message.error(err.message || t('compose.messages.saveFailed'))
  } finally {
    editorLoading.value = false
  }
}

/* ==================== 栈删除 ==================== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteTarget = ref<ComposeStack | null>(null)

function openDeleteStack(row: ComposeStack) {
  deleteTarget.value = row
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  try {
    await deleteComposeStack(deleteTarget.value.name)
    message.success(t('compose.messages.deleted'))
    showDelete.value = false
    loadStacks()
  } catch (err: any) {
    message.error(err.message || t('compose.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 镜像构建 ==================== */
const buildTag = ref('')
const buildDockerfile = ref('')
const buildLines = ref<string[]>([])
const building = ref(false)
let abortController: AbortController | null = null
const logRef = ref<HTMLElement | null>(null)

async function handleBuild() {
  const tag = buildTag.value.trim()
  if (!tag) {
    message.error(t('compose.rules.tagRequired'))
    return
  }
  if (!buildDockerfile.value.trim()) {
    message.error(t('compose.rules.dockerfileRequired'))
    return
  }
  building.value = true
  buildLines.value = []
  abortController = new AbortController()
  try {
    await buildImage(tag, buildDockerfile.value, appendLine, abortController.signal)
  } catch (err: any) {
    if (err?.name === 'AbortError') {
      appendLine(t('compose.messages.buildAborted'))
    } else {
      appendLine(`[ERROR] ${err.message || t('compose.messages.buildFailed')}`)
      message.error(err.message || t('compose.messages.buildFailed'))
    }
  } finally {
    building.value = false
    abortController = null
  }
}

function handleAbort() {
  abortController?.abort()
}

function appendLine(line: string) {
  buildLines.value.push(line)
  if (line.includes('[ERROR]')) {
    message.error(t('compose.messages.buildError'))
  }
  if (line.includes('[DONE]')) {
    message.success(t('compose.messages.buildDone'))
  }
  nextTick(() => {
    const el = logRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

defineExpose({ loadStacks })
void nextTick

// 页面挂载后由父组件决定是否调用 loadStacks；此处保证首次渲染也有数据
loadStacks()
</script>

<script lang="ts">
import { h } from 'vue'
</script>

<template>
  <div class="compose-page">
    <NTabs v-model:value="activeTab" type="line" animated class="compose-tabs">
    <NTabPane name="stacks" :tab="t('compose.cardTitle')">
    <!-- 卡片一：Compose 栈 -->
    <NCard :title="t('compose.cardTitle')" size="small" class="block-card">
      <template #header-extra>
        <NSpace>
          <NButton size="small" @click="loadStacks">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>
            {{ t('compose.refresh') }}
          </NButton>
          <NButton size="small" type="primary" @click="openCreateStack">
            <template #icon><NIcon><AddOutline /></NIcon></template>
            {{ t('compose.createStack') }}
          </NButton>
        </NSpace>
      </template>

      <DataTable
        :columns="stackColumns"
        :data="stacks"
        :loading="stackLoading"
        :bordered="false"
        :pagination="false"
        size="small"
      >
        <template #empty>
          <div class="empty-state">
            <p>{{ t('compose.empty.noStack') }}</p>
            <span>{{ t('compose.empty.noStackHint') }}</span>
          </div>
        </template>
      </DataTable>
    </NCard>
    </NTabPane>

    <NTabPane name="build" :tab="t('compose.buildCardTitle')">
    <!-- 卡片二：镜像构建 -->
    <NCard :title="t('compose.buildCardTitle')" size="small" class="block-card">
      <div class="build-form">
        <div class="build-row">
          <span class="field-label">{{ t('compose.build.tagLabel') }}</span>
          <NInput v-model:value="buildTag" placeholder="myapp:latest" :disabled="building" />
        </div>
        <div class="build-row">
          <span class="field-label">Dockerfile</span>
          <NInput
            v-model:value="buildDockerfile"
            type="textarea"
            :rows="12"
            class="mono-input"
            placeholder="FROM nginx:alpine&#10;COPY . /usr/share/nginx/html"
            :disabled="building"
          />
        </div>
        <NSpace>
          <NButton type="primary" :loading="building" :disabled="building" @click="handleBuild">
            <template #icon><NIcon><HammerOutline /></NIcon></template>
            {{ t('compose.build.start') }}
          </NButton>
          <NButton v-if="building" type="error" @click="handleAbort">
            <template #icon><NIcon><CloseOutline /></NIcon></template>
            {{ t('compose.build.abort') }}
          </NButton>
        </NSpace>
      </div>

      <pre v-if="buildLines.length" ref="logRef" class="build-log">{{ buildLines.join('\n') }}</pre>
      <div v-else class="empty-state">
        <p>{{ t('compose.empty.noBuildLog') }}</p>
        <span>{{ t('compose.empty.noBuildLogHint') }}</span>
      </div>
    </NCard>
    </NTabPane>
    </NTabs>

    <!-- 栈编辑 -->
    <ContentModal
      v-model:show="showEditor"
      :title="editorNameDisabled ? t('compose.editor.editTitle', { name: editorName }) : t('compose.editor.createTitle')"
      size="lg"
    >
      <div class="editor-form">
        <div class="build-row">
          <span class="field-label">{{ t('compose.editor.nameLabel') }}</span>
          <NInput
            v-model:value="editorName"
            :placeholder="t('compose.editor.namePlaceholder')"
            :disabled="editorNameDisabled"
          />
        </div>
        <div class="build-row">
          <span class="field-label">docker-compose.yml</span>
          <NInput
            v-model:value="editorContent"
            type="textarea"
            :rows="16"
            class="mono-input"
            placeholder="services:&#10;  web:&#10;    image: nginx:alpine"
          />
        </div>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditor = false">{{ t('compose.editor.cancel') }}</NButton>
          <NButton type="primary" :loading="editorLoading" @click="handleSaveStack">{{ t('compose.editor.save') }}</NButton>
        </NSpace>
      </template>
    </ContentModal>

    <!-- 删除确认 -->
    <ConfirmDialog
      :show="showDelete"
      :title="t('compose.deleteDialog.title')"
      :content="t('compose.deleteDialog.content', { name: deleteTarget?.name || '' })"
      :confirm-text="t('compose.deleteDialog.confirm')"
      type="error"
      :loading="deleteLoading"
      @update:show="(v: boolean) => (showDelete = v)"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>

<style scoped>
.compose-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}
.compose-tabs :deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
.block-card {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
}
.build-form,
.editor-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.build-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.field-label {
  font-size: var(--fs-caption);
  color: var(--sub);
  font-weight: 600;
}
.mono-input :deep(textarea) {
  font-family: 'JetBrains Mono', 'Cascadia Code', Consolas, monospace;
  font-size: var(--fs-caption);
}
.build-log {
  margin-top: var(--space-4);
  max-height: 320px;
  overflow: auto;
  background: var(--surface2);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  font-family: 'JetBrains Mono', 'Cascadia Code', Consolas, monospace;
  font-size: var(--fs-caption);
  color: var(--sub);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
.empty-state {
  padding: var(--space-8) 0;
  text-align: center;
}
.empty-state p {
  color: var(--sub);
  font-size: var(--fs-body);
  margin: 0 0 var(--space-2);
}
.empty-state span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
</style>
