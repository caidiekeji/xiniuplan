<script setup lang="ts">
import { ref, reactive, computed, watch, h, onMounted, onBeforeUnmount } from 'vue'
import { useMessage } from '@/composables/useMessage'
import { useI18n } from 'vue-i18n'
import {
  NInput, NButton, NSpace, NIcon, NTree, NFormItem, NSelect,
  NDrawer, NDrawerContent, NAlert, NProgress, NDropdown, NTag
} from 'naive-ui'
import type { DataTableColumns, TreeOption } from 'naive-ui'
import {
  FolderOutline, DocumentTextOutline, ImageOutline, CodeSlashOutline, ArchiveOutline,
  AddOutline, RefreshOutline, SearchOutline, CloudUploadOutline, CloudDownloadOutline,
  CreateOutline, TrashOutline, CopyOutline, MoveOutline, EllipsisHorizontalOutline
} from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, ConfirmDialog, CodeEditor } from '@/components'
import { useQueryState } from '@/composables/useQueryState'
import {
  listFiles, createDir, createFile, getFileContent, writeFileContent,
  renameFile, copyFile, moveFile, deleteFiles, compressFiles, decompressFile,
  searchFiles, uploadFile, downloadFileAs,
  type FileItem, type CompressFormat
} from '@/api/file'
import { listRecycle, restoreRecycle, purgeRecycle, type RecycleItem } from '@/api/features17'

const message = useMessage()
const { t } = useI18n()

/* ==================== 移动端适配 ==================== */
const showTree = ref(true)
const isMobile = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value) showTree.value = false
}

/* ==================== 工具函数 ==================== */
function dirname(p: string) {
  const i = p.lastIndexOf('/')
  return i <= 0 ? '/' : p.slice(0, i)
}
function basename(p: string) {
  return p.split('/').pop() || p
}
function joinPath(a: string, b: string) {
  const r = (a.replace(/\/+$/, '') + '/' + b.replace(/^\/+/, ''))
  return r || '/'
}
function normalizePath(p: string) {
  if (!p) return '/'
  if (!p.startsWith('/')) p = '/' + p
  p = p.replace(/\/+$/, '')
  return p || '/'
}
function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return (bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
}
function formatTime(tk: string) {
  if (!tk) return '-'
  const d = new Date(tk)
  if (isNaN(d.getTime())) return tk
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
const TEXT_EXT = /\.(js|ts|vue|jsx|tsx|py|go|rs|java|c|cpp|h|cc|sh|bash|json|yaml|yml|xml|html|htm|css|scss|less|sql|md|markdown|conf|ini|env|txt|log|toml|cfg)$/i
const ARCHIVE_EXT = /\.(zip|tar|gz|tgz|bz2|xz|7z|rar)$/i
function isTextFile(name: string) { return TEXT_EXT.test(name) }
function isArchive(name: string) { return ARCHIVE_EXT.test(name) }
function getFileIcon(row: FileItem) {
  if (row.is_dir) return FolderOutline
  const n = row.name.toLowerCase()
  if (/\.(png|jpe?g|gif|svg|webp|bmp|ico)$/.test(n)) return ImageOutline
  if (ARCHIVE_EXT.test(n)) return ArchiveOutline
  if (TEXT_EXT.test(n)) return CodeSlashOutline
  return DocumentTextOutline
}
function detectLanguage(name: string) {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const map: Record<string, string> = {
    js: 'javascript', ts: 'typescript', vue: 'html', jsx: 'javascript', tsx: 'typescript',
    py: 'python', go: 'go', rs: 'rust', java: 'java', c: 'c', cpp: 'cpp', cc: 'cpp', h: 'cpp',
    sh: 'shell', bash: 'shell', json: 'json', yaml: 'yaml', yml: 'yaml', xml: 'xml',
    html: 'html', htm: 'html', css: 'css', scss: 'scss', less: 'less', sql: 'sql',
    md: 'markdown', markdown: 'markdown', conf: 'ini', ini: 'ini', env: 'ini', txt: 'plaintext',
    log: 'plaintext', toml: 'ini', cfg: 'ini'
  }
  return map[ext] || 'plaintext'
}

/* ==================== 状态 ==================== */
// 路径 / 搜索词 URL 化（D19）：目录深链可直达、刷新不丢位置
const currentPath = useQueryState('path', '/')
const files = ref<FileItem[]>([])
const loading = ref(false)
/** 目录加载失败信息：非空时表格区渲染错误面板（含重试） */
const loadError = ref<string | null>(null)
const checkedKeys = ref<(string | number)[]>([])

const searchKeyword = useQueryState('kw', '')
const searchResults = ref<any[]>([])
const searching = ref(false)
const inSearchMode = computed(() => searchKeyword.value.trim().length > 0)

const treeData = ref<TreeOption[]>([{ key: '/', label: '/', isLeaf: false }])

const dialog = reactive({
  mkdir: false, mkfile: false, rename: false, move: false, copy: false,
  compress: false, decompress: false,
  form: {
    dirName: '', fileName: '', fileContent: '',
    renameNewName: '', targetPath: '',
    compressFormat: 'zip' as CompressFormat, compressDest: '',
    decompressDest: ''
  },
  target: null as FileItem | null,
  srcPaths: [] as string[]
})

const editor = reactive({
  show: false, path: '', name: '', content: '', language: 'plaintext',
  loading: false, saving: false, size: 0
})

interface UploadItem {
  file: File
  name: string
  size: number
  progress: number
  status: 'pending' | 'uploading' | 'done' | 'error'
  error?: string
}
const upload = reactive({
  show: false, items: [] as UploadItem[], uploading: false, dragOver: false
})
const fileInput = ref<HTMLInputElement | null>(null)

const del = reactive({
  show: false, paths: [] as string[], loading: false
})

/* ==================== 面包屑 / 目录树 ==================== */
const breadcrumbParts = computed(() => {
  const p = currentPath.value
  const parts = p === '/' ? [] : p.split('/').filter(Boolean)
  const result: { label: string; path: string }[] = [{ label: '/', path: '/' }]
  let acc = ''
  for (const part of parts) {
    acc += '/' + part
    result.push({ label: part, path: acc })
  }
  return result
})

async function loadTreeNode(node: TreeOption) {
  try {
    const res = await listFiles(node.key as string)
    node.children = res.data.items
      .filter((f) => f.is_dir)
      .map((f) => ({ key: f.path, label: f.name, isLeaf: false }))
  } catch {
    node.children = []
  }
}
function handleTreeSelect(keys: (string | number)[]) {
  if (keys.length) navigateTo(String(keys[0]))
}

/* ==================== 列表加载 ==================== */
async function loadList() {
  loading.value = true
  loadError.value = null
  try {
    const res = await listFiles(currentPath.value)
    files.value = (res.data?.items || []).map((f) => ({ ...f, key: f.path }))
  } catch (e: any) {
    loadError.value = e.message || t('file.messages.loadDirFailed')
    message.error(e.message || t('file.messages.loadDirFailed'))
    files.value = []
  } finally {
    loading.value = false
  }
}

function navigateTo(path: string) {
  currentPath.value = normalizePath(path)
  checkedKeys.value = []
  searchKeyword.value = ''
  searchResults.value = []
  loadList()
}

/* ==================== 搜索 ==================== */
let searchTimer: number | null = null
watch(searchKeyword, (val) => {
  if (searchTimer) clearTimeout(searchTimer)
  const kw = val.trim()
  if (!kw) {
    searchResults.value = []
    return
  }
  searchTimer = window.setTimeout(() => doSearch(kw), 300)
})
async function doSearch(kw: string) {
  searching.value = true
  try {
    const res = await searchFiles(currentPath.value, kw, false)
    searchResults.value = (res.data?.items || []).map((i) => ({
      ...i, key: i.path, mode: '-', extension: ''
    }))
  } catch (e: any) {
    message.error(e.message || t('file.messages.searchFailed'))
  } finally {
    searching.value = false
  }
}

const displayData = computed(() => (inSearchMode.value ? searchResults.value : files.value))

/* ==================== 表格列 ==================== */
function renderName(row: FileItem) {
  return h(
    'div',
    {
      class: ['file-name', row.is_dir && 'file-name--dir'],
      onClick: () => row.is_dir && navigateTo(row.path)
    },
    [
      h(NIcon, { size: 15, class: 'file-name__icon' }, () => h(getFileIcon(row))),
      h('span', { class: 'file-name__text' }, row.name)
    ]
  )
}
function renderActions(row: FileItem) {
  const options = [
    { label: t('file.menus.move'), key: 'move' },
    { label: t('file.menus.copy'), key: 'copy' },
    { label: t('file.menus.compress'), key: 'compress' }
  ]
  if (!row.is_dir && isArchive(row.name)) {
    options.push({ label: t('file.menus.decompress'), key: 'decompress' })
  }
  function handleMore(key: string | number) {
    if (key === 'move') openMove([row.path])
    else if (key === 'copy') openCopy([row.path])
    else if (key === 'compress') openCompress([row.path])
    else if (key === 'decompress') openDecompress(row)
  }
  return h(NSpace, { size: 2, align: 'center' }, () => [
    h(NButton, { size: 'tiny', quaternary: true, title: t('file.actions.rename'), onClick: () => openRename(row) },
      { default: () => h(NIcon, { size: 14 }, () => h(CreateOutline)) }),
    h(NButton, { size: 'tiny', quaternary: true, title: t('file.actions.download'), onClick: () => handleDownload(row) },
      { default: () => h(NIcon, { size: 14 }, () => h(CloudDownloadOutline)) }),
    h(NButton, { size: 'tiny', quaternary: true, title: t('file.actions.delete'), onClick: () => openDelete([row]) },
      { default: () => h(NIcon, { size: 14 }, () => h(TrashOutline)) }),
    h(NDropdown, { options, trigger: 'click', onSelect: handleMore },
      { default: () => h(NButton, { size: 'tiny', quaternary: true, title: t('file.actions.more') },
        { default: () => h(NIcon, { size: 14 }, () => h(EllipsisHorizontalOutline)) }) })
  ])
}
const columns = computed<DataTableColumns>(() => [
  { type: 'selection', width: 40 },
  { title: t('file.columns.name'), key: 'name', render: renderName as any },
  {
    title: t('file.columns.size'), key: 'size', width: 110,
    render: (row: FileItem) => row.is_dir ? '—' : h('span', { class: 'mono' }, formatSize(row.size))
  },
  {
    title: t('file.columns.mode'), key: 'mode', width: 100,
    render: (row: FileItem) => h('span', { class: 'mono' }, row.mode || '-')
  },
  {
    title: t('file.columns.modTime'), key: 'mod_time', width: 170,
    render: (row: FileItem) => h('span', { class: 'mono' }, formatTime(row.mod_time))
  },
  { title: t('file.columns.actions'), key: 'actions', width: 200, render: renderActions as any }
])

function handleRowDblclick(row: FileItem) {
  if (row.is_dir) {
    navigateTo(row.path)
  } else if (isTextFile(row.name)) {
    openEditor(row)
  } else {
    message.info(t('file.messages.binaryDownload'))
  }
}

/* ==================== 新建 ==================== */
function openMkdir() {
  dialog.form.dirName = ''
  dialog.mkdir = true
}
const dialogLoading = reactive({
  mkdir: false, mkfile: false, rename: false, move: false,
  copy: false, compress: false, decompress: false
})

async function submitMkdir() {
  const name = dialog.form.dirName.trim()
  if (!name) { message.warning(t('file.messages.inputDirName')); return }
  dialogLoading.mkdir = true
  try {
    await createDir(joinPath(currentPath.value, name))
    message.success(t('file.messages.dirCreated'))
    dialog.mkdir = false
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.createFailed')) }
  finally { dialogLoading.mkdir = false }
}
function openMkfile() {
  dialog.form.fileName = ''
  dialog.form.fileContent = ''
  dialog.mkfile = true
}
async function submitMkfile() {
  const name = dialog.form.fileName.trim()
  if (!name) { message.warning(t('file.messages.inputFileName')); return }
  dialogLoading.mkfile = true
  try {
    await createFile(joinPath(currentPath.value, name), dialog.form.fileContent)
    message.success(t('file.messages.fileCreated'))
    dialog.mkfile = false
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.createFailed')) }
  finally { dialogLoading.mkfile = false }
}

/* ==================== 重命名 / 移动 / 复制 ==================== */
function openRename(row: FileItem) {
  dialog.target = row
  dialog.form.renameNewName = row.name
  dialog.rename = true
}
async function submitRename() {
  if (!dialog.target) return
  const name = dialog.form.renameNewName.trim()
  if (!name) { message.warning(t('file.messages.inputNewName')); return }
  const newPath = joinPath(dirname(dialog.target.path), name)
  dialogLoading.rename = true
  try {
    await renameFile(dialog.target.path, newPath)
    message.success(t('file.messages.renameSuccess'))
    dialog.rename = false
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.renameFailed')) }
  finally { dialogLoading.rename = false }
}
function openMove(paths: string[]) {
  dialog.srcPaths = paths
  dialog.form.targetPath = currentPath.value
  dialog.move = true
}
async function submitMove() {
  const dstDir = normalizePath(dialog.form.targetPath)
  dialogLoading.move = true
  try {
    await Promise.all(dialog.srcPaths.map((src) => moveFile(src, joinPath(dstDir, basename(src)))))
    message.success(t('file.messages.moveSuccess'))
    dialog.move = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.moveFailed')) }
  finally { dialogLoading.move = false }
}
function openCopy(paths: string[]) {
  dialog.srcPaths = paths
  dialog.form.targetPath = currentPath.value
  dialog.copy = true
}
async function submitCopy() {
  const dstDir = normalizePath(dialog.form.targetPath)
  dialogLoading.copy = true
  try {
    await Promise.all(dialog.srcPaths.map((src) => copyFile(src, joinPath(dstDir, basename(src)))))
    message.success(t('file.messages.copySuccess'))
    dialog.copy = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.copyFailed')) }
  finally { dialogLoading.copy = false }
}

/* ==================== 压缩 / 解压 ==================== */
function openCompress(paths: string[]) {
  dialog.srcPaths = paths
  dialog.form.compressFormat = 'zip'
  dialog.form.compressDest = joinPath(currentPath.value, 'archive.zip')
  dialog.compress = true
}
async function submitCompress() {
  const dest = dialog.form.compressDest.trim()
  if (!dest) { message.warning(t('file.messages.inputTargetPath')); return }
  dialogLoading.compress = true
  try {
    await compressFiles(dialog.srcPaths, dest, dialog.form.compressFormat)
    message.success(t('file.messages.compressSuccess'))
    dialog.compress = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.compressFailed')) }
  finally { dialogLoading.compress = false }
}
function openDecompress(row: FileItem) {
  dialog.target = row
  dialog.form.decompressDest = dirname(row.path)
  dialog.decompress = true
}
async function submitDecompress() {
  if (!dialog.target) return
  const dest = normalizePath(dialog.form.decompressDest)
  dialogLoading.decompress = true
  try {
    await decompressFile(dialog.target.path, dest)
    message.success(t('file.messages.decompressSuccess'))
    dialog.decompress = false
    loadList()
  } catch (e: any) { message.error(e.message || t('file.messages.decompressFailed')) }
  finally { dialogLoading.decompress = false }
}

/* ==================== 删除 ==================== */
function openDelete(items: FileItem[]) {
  del.paths = items.map((i) => i.path)
  del.show = true
}
function openDeleteByChecked() {
  if (!checkedKeys.value.length) return
  del.paths = checkedKeys.value.map(String)
  del.show = true
}
async function submitDelete() {
  del.loading = true
  try {
    await deleteFiles(del.paths)
    message.success(t('file.messages.deleteSuccess'))
    del.show = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) {
    message.error(e.message || t('file.messages.deleteFailed'))
  } finally {
    del.loading = false
  }
}

/* ==================== 下载 ==================== */
async function handleDownload(row: FileItem) {
  try {
    await downloadFileAs(row.path, row.name)
    message.success(t('file.messages.downloadSuccess'))
  } catch (e: any) {
    message.error(e.message || t('file.messages.downloadFailed'))
  }
}

/* ==================== 上传 ==================== */
function openUpload() {
  upload.items = []
  upload.show = true
}
function handleDrop(e: DragEvent) {
  upload.dragOver = false
  const fl = e.dataTransfer?.files
  if (fl && fl.length) addUploadFiles(Array.from(fl))
}
function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length) addUploadFiles(Array.from(input.files))
  input.value = ''
}
function addUploadFiles(fs: File[]) {
  for (const f of fs) {
    upload.items.push({ file: f, name: f.name, size: f.size, progress: 0, status: 'pending' })
  }
}
function removeUploadItem(idx: number) {
  upload.items.splice(idx, 1)
}
async function startUpload() {
  if (!upload.items.length) { message.warning(t('file.messages.selectFileFirst')); return }
  upload.uploading = true
  try {
    for (const item of upload.items) {
      if (item.status === 'done') continue
      item.status = 'uploading'
      item.progress = 0
      const fd = new FormData()
      fd.append('path', currentPath.value)
      fd.append('files', item.file)
      try {
        await uploadFile(fd, (p) => { item.progress = p })
        item.status = 'done'
        item.progress = 100
      } catch (err: any) {
        item.status = 'error'
        item.error = err.message || t('file.messages.uploadFailed')
      }
    }
    const ok = upload.items.filter((i) => i.status === 'done').length
    message.success(t('file.messages.uploadDone', { done: ok, total: upload.items.length }))
    upload.show = false
    upload.items = []
    loadList()
  } finally {
    upload.uploading = false
  }
}

/* ==================== 编辑器 ==================== */
async function openEditor(row: FileItem) {
  editor.show = true
  editor.path = row.path
  editor.name = row.name
  editor.language = detectLanguage(row.name)
  editor.content = ''
  editor.size = row.size
  editor.loading = true
  try {
    const res = await getFileContent(row.path)
    editor.content = res.data.content
    editor.size = res.data.size ?? row.size
  } catch (e: any) {
    message.error(e.message || t('file.messages.readFileFailed'))
  } finally {
    editor.loading = false
  }
}
async function saveEditor() {
  if (!editor.path) return
  editor.saving = true
  try {
    await writeFileContent(editor.path, editor.content)
    message.success(t('file.messages.saveSuccess'))
  } catch (e: any) {
    message.error(e.message || t('file.messages.saveFailed'))
  } finally {
    editor.saving = false
  }
}
function onKeydown(e: KeyboardEvent) {
  if (editor.show && (e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault()
    saveEditor()
  }
}

/* ==================== 批量 ==================== */
const checkedPaths = computed(() => checkedKeys.value.map(String))

/* ==================== 回收站 ==================== */
const showRecycle = ref(false)
const recycleLoading = ref(false)
const recycleItems = ref<RecycleItem[]>([])
const recycleActionId = ref<number | null>(null)

async function openRecycle() {
  showRecycle.value = true
  await loadRecycle()
}

async function loadRecycle() {
  recycleLoading.value = true
  try {
    const data = await listRecycle()
    recycleItems.value = data?.items || []
  } catch (e: any) {
    recycleItems.value = []
    message.error(e?.message || t('file.messages.loadRecycleFailed'))
  } finally {
    recycleLoading.value = false
  }
}

async function doRestore(item: RecycleItem) {
  recycleActionId.value = item.id
  try {
    await restoreRecycle(item.id)
    message.success(t('file.messages.restored'))
    await loadRecycle()
    loadList()
  } catch (e: any) {
    message.error(e?.message || t('file.messages.restoreFailed'))
  } finally {
    recycleActionId.value = null
  }
}

async function doPurge(item: RecycleItem) {
  recycleActionId.value = item.id
  try {
    await purgeRecycle(item.id)
    message.success(t('file.messages.purged'))
    await loadRecycle()
  } catch (e: any) {
    message.error(e?.message || t('file.messages.deleteFailed'))
  } finally {
    recycleActionId.value = null
  }
}

const recycleColumns = computed<DataTableColumns>(() => [
  {
    title: t('file.recycle.originalPath'),
    key: 'original_path',
    render: (row: any) => h('span', { class: 'mono', style: 'font-size:var(--fs-caption);word-break:break-all' }, row.original_path)
  },
  {
    title: t('file.recycle.size'),
    key: 'size',
    width: 90,
    render: (row: any) => h('span', { class: 'mono' }, formatSize(row.size))
  },
  {
    title: t('file.recycle.deletedAt'),
    key: 'deleted_at',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.deleted_at)
  },
  {
    title: t('file.columns.actions'),
    key: 'actions',
    width: 150,
    render: (row: any) => h(NSpace, { size: 'small', wrap: false }, () => [
      h(NButton, {
        size: 'tiny',
        type: 'primary',
        quaternary: true,
        loading: recycleActionId.value === row.id,
        onClick: () => doRestore(row as RecycleItem)
      }, { default: () => t('file.actions.restore') }),
      h(NButton, {
        size: 'tiny',
        type: 'error',
        quaternary: true,
        loading: recycleActionId.value === row.id,
        onClick: () => doPurge(row as RecycleItem)
      }, { default: () => t('file.actions.purge') })
    ])
  }
])

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  window.addEventListener('keydown', onKeydown)
  loadList()
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', checkMobile)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <PageCard class="file-page" padding="0">
    <!-- 顶部工具栏 -->
    <div class="fm-toolbar">
      <div class="fm-breadcrumb">
        <template v-for="(p, i) in breadcrumbParts" :key="p.path">
          <span v-if="i > 0" class="crumb-sep">/</span>
          <a class="crumb" :class="{ 'crumb--active': i === breadcrumbParts.length - 1 }" @click="navigateTo(p.path)">
            {{ p.label }}
          </a>
        </template>
      </div>
      <NSpace class="fm-actions" size="small">
        <NButton v-if="isMobile" size="small" quaternary circle @click="showTree = !showTree" :title="t('file.treeTitle')">
          <template #icon><NIcon size="14"><FolderOutline /></NIcon></template>
        </NButton>
        <NInput
          v-model:value="searchKeyword"
          :placeholder="t('file.search.placeholder')"
          clearable
          size="small"
          class="fm-search"
        >
          <template #prefix>
            <NIcon size="14"><SearchOutline /></NIcon>
          </template>
        </NInput>
        <NButton size="small" @click="openMkdir">
          <template #icon><NIcon size="14"><FolderOutline /></NIcon></template>
          {{ t('file.buttons.mkdir') }}
        </NButton>
        <NButton size="small" @click="openMkfile">
          <template #icon><NIcon size="14"><AddOutline /></NIcon></template>
          {{ t('file.buttons.mkfile') }}
        </NButton>
        <NButton size="small" @click="openUpload">
          <template #icon><NIcon size="14"><CloudUploadOutline /></NIcon></template>
          {{ t('file.buttons.upload') }}
        </NButton>
        <NButton size="small" @click="openRecycle">
          <template #icon><NIcon size="14"><TrashOutline /></NIcon></template>
          {{ t('file.recycleTitle') }}
        </NButton>
        <NButton size="small" quaternary circle @click="loadList" :title="t('file.buttons.refresh')">
          <template #icon><NIcon size="14"><RefreshOutline /></NIcon></template>
        </NButton>
      </NSpace>
    </div>

    <!-- 批量操作条 -->
    <div v-if="checkedKeys.length" class="fm-batchbar">
      <span class="fm-batchbar__count">{{ t('file.batch.selectedCount', { n: checkedKeys.length }) }}</span>
      <NSpace size="small">
        <NButton size="tiny" @click="openCopy(checkedPaths)">
          <template #icon><NIcon size="12"><CopyOutline /></NIcon></template>{{ t('file.batch.copy') }}
        </NButton>
        <NButton size="tiny" @click="openMove(checkedPaths)">
          <template #icon><NIcon size="12"><MoveOutline /></NIcon></template>{{ t('file.batch.move') }}
        </NButton>
        <NButton size="tiny" @click="openCompress(checkedPaths)">{{ t('file.batch.compress') }}</NButton>
        <NButton size="tiny" type="error" @click="openDeleteByChecked">
          <template #icon><NIcon size="12"><TrashOutline /></NIcon></template>{{ t('file.actions.delete') }}
        </NButton>
        <NButton size="tiny" quaternary @click="checkedKeys = []">{{ t('file.batch.clear') }}</NButton>
      </NSpace>
    </div>

    <!-- 搜索提示 -->
    <div v-if="inSearchMode" class="fm-search-hint">
      <NTag size="small" type="info" :bordered="false">{{ t('file.search.searching') }}</NTag>
      <span>{{ t('file.search.hint', { keyword: searchKeyword.trim(), count: searchResults.length }) }}</span>
      <a class="fm-search-hint__exit" @click="searchKeyword = ''">{{ t('file.search.exit') }}</a>
    </div>

    <!-- 主体：左树右表 -->
    <div class="fm-body">
      <div class="fm-tree" :class="{ 'fm-tree--hidden': isMobile && !showTree }">
        <div class="fm-tree__title">{{ t('file.treeTitle') }}</div>
        <NTree
          :data="treeData"
          :load-data="loadTreeNode"
          :selected-keys="[currentPath]"
          block-line
          @update:selected-keys="handleTreeSelect"
        />
      </div>
      <div class="fm-list">
        <DataTable
          :columns="columns"
          :data="displayData"
          :loading="loading || searching"
          :pagination="false"
          :row-key="(row: FileItem) => row.path"
          :error="loadError"
          @retry="loadList"
          v-model:checkedRowKeys="checkedKeys"
          @rowDblclick="handleRowDblclick"
        />
      </div>
    </div>
  </PageCard>

  <!-- 回收站 -->
  <NDrawer :show="showRecycle" :width="'var(--drawer-width)'" placement="right" @update:show="(v: boolean) => (showRecycle = v)">
    <NDrawerContent :title="t('file.recycleTitle')" closable>
      <NAlert type="info" :bordered="false" style="margin-bottom: var(--space-3)">
        {{ t('file.recycle.alert') }}
      </NAlert>
      <DataTable
        :columns="recycleColumns"
        :data="recycleItems"
        :loading="recycleLoading"
        :pagination="false"
        size="small"
      />
    </NDrawerContent>
  </NDrawer>

  <!-- 新建文件夹 -->
  <ModalForm
    v-model:show="dialog.mkdir"
    :title="t('file.dialogs.mkdir.title')"
    :loading="dialogLoading.mkdir"
    @submit="submitMkdir"
  >
    <NFormItem :label="t('file.dialogs.mkdir.name')" required>
      <NInput v-model:value="dialog.form.dirName" :placeholder="t('file.dialogs.mkdir.namePlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.mkdir.path')">
      <NInput :value="currentPath" readonly />
    </NFormItem>
  </ModalForm>

  <!-- 新建文件 -->
  <ModalForm
    v-model:show="dialog.mkfile"
    :title="t('file.dialogs.mkfile.title')"
    :loading="dialogLoading.mkfile"
    @submit="submitMkfile"
    :width="560"
  >
    <NFormItem :label="t('file.dialogs.mkfile.name')" required>
      <NInput v-model:value="dialog.form.fileName" :placeholder="t('file.dialogs.mkfile.namePlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.mkfile.content')">
      <NInput v-model:value="dialog.form.fileContent" type="textarea" :rows="6" :placeholder="t('file.dialogs.mkfile.contentPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 重命名 -->
  <ModalForm
    v-model:show="dialog.rename"
    :title="t('file.dialogs.rename.title')"
    :loading="dialogLoading.rename"
    @submit="submitRename"
  >
    <NFormItem :label="t('file.dialogs.rename.oldName')">
      <NInput :value="dialog.target?.name" readonly />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.rename.newName')" required>
      <NInput v-model:value="dialog.form.renameNewName" />
    </NFormItem>
  </ModalForm>

  <!-- 移动 -->
  <ModalForm
    v-model:show="dialog.move"
    :title="t('file.dialogs.move.title')"
    :loading="dialogLoading.move"
    @submit="submitMove"
  >
    <NFormItem :label="t('file.dialogs.move.source')">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.move.target')" required>
      <NInput v-model:value="dialog.form.targetPath" :placeholder="t('file.dialogs.move.targetPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 复制 -->
  <ModalForm
    v-model:show="dialog.copy"
    :title="t('file.dialogs.copy.title')"
    :loading="dialogLoading.copy"
    @submit="submitCopy"
  >
    <NFormItem :label="t('file.dialogs.copy.source')">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.copy.target')" required>
      <NInput v-model:value="dialog.form.targetPath" :placeholder="t('file.dialogs.copy.targetPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 压缩 -->
  <ModalForm
    v-model:show="dialog.compress"
    :title="t('file.dialogs.compress.title')"
    :loading="dialogLoading.compress"
    @submit="submitCompress"
  >
    <NFormItem :label="t('file.dialogs.compress.source')">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.compress.format')">
      <NSelect v-model:value="dialog.form.compressFormat" :options="[
        { label: 'zip', value: 'zip' },
        { label: 'tar.gz', value: 'tar.gz' }
      ]" />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.compress.dest')" required>
      <NInput v-model:value="dialog.form.compressDest" :placeholder="t('file.dialogs.compress.destPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 解压 -->
  <ModalForm
    v-model:show="dialog.decompress"
    :title="t('file.dialogs.decompress.title')"
    :loading="dialogLoading.decompress"
    @submit="submitDecompress"
  >
    <NFormItem :label="t('file.dialogs.decompress.archive')">
      <NInput :value="dialog.target?.path" readonly />
    </NFormItem>
    <NFormItem :label="t('file.dialogs.decompress.target')" required>
      <NInput v-model:value="dialog.form.decompressDest" :placeholder="t('file.dialogs.decompress.targetPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 上传 -->
  <ModalForm v-model:show="upload.show" :title="t('file.dialogs.upload.title')" :loading="upload.uploading" @submit="startUpload" :width="600">
    <NFormItem :label="t('file.dialogs.upload.target')">
      <NInput :value="currentPath" readonly />
    </NFormItem>
    <div
      class="dropzone"
      :class="{ 'dropzone--over': upload.dragOver }"
      @dragover.prevent="upload.dragOver = true"
      @dragleave="upload.dragOver = false"
      @drop.prevent="handleDrop"
      @click="fileInput?.click()"
    >
      <input ref="fileInput" type="file" multiple hidden @change="handleFileSelect" />
      <NIcon size="32" class="dropzone__icon"><CloudUploadOutline /></NIcon>
      <p class="dropzone__title">{{ t('file.dialogs.upload.dropTitle') }}</p>
      <p class="dropzone__hint">{{ t('file.dialogs.upload.dropHint') }}</p>
    </div>
    <div v-if="upload.items.length" class="upload-list">
      <div v-for="(item, idx) in upload.items" :key="item.name + idx" class="upload-item">
        <div class="upload-item__head">
          <span class="upload-item__name">{{ item.name }}</span>
          <span class="upload-item__meta mono">{{ formatSize(item.size) }} · {{ item.progress }}%</span>
          <NButton v-if="item.status !== 'uploading'" size="tiny" text @click="removeUploadItem(idx)">{{ t('file.actions.remove') }}</NButton>
        </div>
        <NProgress
          :percentage="item.progress"
          :status="item.status === 'error' ? 'error' : item.status === 'done' ? 'success' : 'default'"
          :indicator-placement="'inside'"
        />
        <div v-if="item.error" class="upload-item__error">{{ item.error }}</div>
      </div>
    </div>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    v-model:show="del.show"
    :title="t('file.dialogs.delete.title')"
    :content="t('file.dialogs.delete.content', { count: del.paths.length })"
    type="error"
    :confirm-text="t('file.actions.delete')"
    :loading="del.loading"
    @confirm="submitDelete"
  />

  <!-- 代码编辑器抽屉 -->
  <NDrawer v-model:show="editor.show" :width="'var(--drawer-width-lg)'" placement="right">
    <NDrawerContent :title="t('file.dialogs.editor.title', { name: editor.name })" :closable="true">
      <div class="editor-meta">
        <span class="mono editor-meta__path">{{ editor.path }}</span>
        <NTag size="small" :bordered="false">{{ editor.language }}</NTag>
      </div>
      <NAlert
        v-if="editor.size > 1048576"
        type="warning"
        style="margin-bottom: var(--space-3)"
        :title="t('file.dialogs.editor.largeFile', { size: formatSize(editor.size) })"
      />
      <CodeEditor
        v-model:modelValue="editor.content"
        :language="editor.language"
        height="calc(100vh - 280px)"
      />
      <template #footer>
        <NSpace justify="end">
          <NButton @click="editor.show = false">{{ t('file.actions.close') }}</NButton>
          <NButton type="primary" :loading="editor.saving" @click="saveEditor">
            {{ t('file.actions.save') }}
          </NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped>
.file-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 104px);
}

.fm-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-5);
  border-bottom: 1px solid var(--line);
  flex-wrap: wrap;
}

.fm-breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-0);
  font-size: var(--fs-body);
}

.crumb {
  color: var(--sub);
  cursor: pointer;
  padding: var(--space-0) var(--space-2);
  border-radius: var(--radius-sm);
  transition: color var(--t-fast), background var(--t-fast);
}

.crumb:hover {
  color: var(--acc);
  background: var(--acc-soft);
}

.crumb--active {
  color: var(--ink);
  font-weight: 600;
}

.crumb-sep {
  color: var(--text-quaternary);
}

.fm-actions {
  flex-shrink: 0;
}

.fm-search {
  width: var(--input-xl);
}

.fm-batchbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) var(--space-5);
  background: var(--acc-soft);
  border-bottom: 1px solid var(--line);
}

.fm-batchbar__count {
  font-size: var(--fs-caption);
  color: var(--acc);
  font-weight: 600;
}

.fm-search-hint {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-5);
  background: var(--info-soft);
  border-bottom: 1px solid var(--line);
  font-size: var(--fs-caption);
  color: var(--sub);
}

.fm-search-hint__exit {
  color: var(--acc);
  cursor: pointer;
  margin-left: auto;
}

.fm-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.fm-tree {
  width: var(--sidebar-width);
  border-right: 1px solid var(--line);
  overflow-y: auto;
  padding: var(--space-3);
  flex-shrink: 0;
}

.fm-tree__title {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-2);
  padding-left: var(--space-2);
}

.fm-list {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: var(--space-3);
}

.file-name {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.file-name--dir {
  cursor: pointer;
  color: var(--acc);
}

.file-name--dir:hover .file-name__text {
  text-decoration: underline;
}

.file-name__icon {
  flex-shrink: 0;
}

.file-name__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 上传拖拽区 */
.dropzone {
  border: 2px dashed var(--line-strong);
  border-radius: var(--radius-md);
  padding: var(--space-6) var(--space-4);
  text-align: center;
  cursor: pointer;
  transition: border-color var(--t-fast), background var(--t-fast);
}

.dropzone:hover,
.dropzone--over {
  border-color: var(--acc);
  background: var(--acc-soft);
}

.dropzone__icon {
  color: var(--acc);
  margin-bottom: var(--space-2);
}

.dropzone__title {
  color: var(--ink);
  font-size: var(--fs-body);
  margin-bottom: var(--space-1);
}

.dropzone__hint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}

.upload-list {
  margin-top: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.upload-item__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-1);
}

.upload-item__name {
  flex: 1;
  font-size: var(--fs-body);
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-item__meta {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  flex-shrink: 0;
}

.upload-item__error {
  font-size: var(--fs-caption);
  color: var(--err);
  margin-top: var(--space-1);
}

/* 编辑器 */
.editor-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.editor-meta__path {
  flex: 1;
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .file-page {
    height: calc(100vh - 96px);
  }
  .fm-toolbar {
    padding: var(--space-2) var(--space-3);
    gap: var(--space-2);
  }
  .fm-search {
    width: 100% !important;
    order: -1;
  }
  .fm-actions {
    width: 100%;
    flex-wrap: wrap;
  }
  .fm-tree {
    position: absolute;
    z-index: 10;
    height: 100%;
    background: var(--surface);
    box-shadow: var(--shadow-lg);
    transition: transform var(--t-normal) var(--ease);
  }
  .fm-tree--hidden {
    transform: translateX(-100%);
  }
  .fm-batchbar {
    padding: var(--space-2) var(--space-3);
    flex-wrap: wrap;
    gap: var(--space-2);
  }
}
</style>
