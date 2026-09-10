<script setup lang="ts">
import { ref, reactive, computed, watch, h, onMounted, onBeforeUnmount } from 'vue'
import { useMessage } from 'naive-ui'
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
import {
  listFiles, createDir, createFile, getFileContent, writeFileContent,
  renameFile, copyFile, moveFile, deleteFiles, compressFiles, decompressFile,
  searchFiles, uploadFile, downloadFileAs,
  type FileItem, type CompressFormat
} from '@/api/file'

const message = useMessage()

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
function formatTime(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return t
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
const currentPath = ref('/')
const files = ref<FileItem[]>([])
const loading = ref(false)
const checkedKeys = ref<(string | number)[]>([])

const searchKeyword = ref('')
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
  try {
    const res = await listFiles(currentPath.value)
    files.value = (res.data?.items || []).map((f) => ({ ...f, key: f.path }))
  } catch (e: any) {
    message.error(e.message || '加载目录失败')
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
    message.error(e.message || '搜索失败')
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
    { label: '移动', key: 'move' },
    { label: '复制', key: 'copy' },
    { label: '压缩', key: 'compress' }
  ]
  if (!row.is_dir && isArchive(row.name)) {
    options.push({ label: '解压到…', key: 'decompress' })
  }
  function handleMore(key: string | number) {
    if (key === 'move') openMove([row.path])
    else if (key === 'copy') openCopy([row.path])
    else if (key === 'compress') openCompress([row.path])
    else if (key === 'decompress') openDecompress(row)
  }
  return h(NSpace, { size: 2, align: 'center' }, () => [
    h(NButton, { size: 'tiny', quaternary: true, title: '重命名', onClick: () => openRename(row) },
      { default: () => h(NIcon, { size: 14 }, () => h(CreateOutline)) }),
    h(NButton, { size: 'tiny', quaternary: true, title: '下载', onClick: () => handleDownload(row) },
      { default: () => h(NIcon, { size: 14 }, () => h(CloudDownloadOutline)) }),
    h(NButton, { size: 'tiny', quaternary: true, title: '删除', onClick: () => openDelete([row]) },
      { default: () => h(NIcon, { size: 14 }, () => h(TrashOutline)) }),
    h(NDropdown, { options, trigger: 'click', onSelect: handleMore },
      { default: () => h(NButton, { size: 'tiny', quaternary: true, title: '更多' },
        { default: () => h(NIcon, { size: 14 }, () => h(EllipsisHorizontalOutline)) }) })
  ])
}
const columns: DataTableColumns = [
  { type: 'selection', width: 40 },
  { title: '名称', key: 'name', render: renderName as any },
  {
    title: '大小', key: 'size', width: 110,
    render: (row: FileItem) => row.is_dir ? '—' : h('span', { class: 'mono' }, formatSize(row.size))
  },
  {
    title: '权限', key: 'mode', width: 100,
    render: (row: FileItem) => h('span', { class: 'mono' }, row.mode || '-')
  },
  {
    title: '修改时间', key: 'mod_time', width: 170,
    render: (row: FileItem) => h('span', { class: 'mono' }, formatTime(row.mod_time))
  },
  { title: '操作', key: 'actions', width: 200, render: renderActions as any }
]

function handleRowDblclick(row: FileItem) {
  if (row.is_dir) {
    navigateTo(row.path)
  } else if (isTextFile(row.name)) {
    openEditor(row)
  } else {
    message.info('二进制文件请下载后查看')
  }
}

/* ==================== 新建 ==================== */
function openMkdir() {
  dialog.form.dirName = ''
  dialog.mkdir = true
}
async function submitMkdir() {
  const name = dialog.form.dirName.trim()
  if (!name) { message.warning('请输入文件夹名称'); return }
  try {
    await createDir(joinPath(currentPath.value, name))
    message.success('文件夹已创建')
    dialog.mkdir = false
    loadList()
  } catch (e: any) { message.error(e.message || '创建失败') }
}
function openMkfile() {
  dialog.form.fileName = ''
  dialog.form.fileContent = ''
  dialog.mkfile = true
}
async function submitMkfile() {
  const name = dialog.form.fileName.trim()
  if (!name) { message.warning('请输入文件名'); return }
  try {
    await createFile(joinPath(currentPath.value, name), dialog.form.fileContent)
    message.success('文件已创建')
    dialog.mkfile = false
    loadList()
  } catch (e: any) { message.error(e.message || '创建失败') }
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
  if (!name) { message.warning('请输入新名称'); return }
  const newPath = joinPath(dirname(dialog.target.path), name)
  try {
    await renameFile(dialog.target.path, newPath)
    message.success('重命名成功')
    dialog.rename = false
    loadList()
  } catch (e: any) { message.error(e.message || '重命名失败') }
}
function openMove(paths: string[]) {
  dialog.srcPaths = paths
  dialog.form.targetPath = currentPath.value
  dialog.move = true
}
async function submitMove() {
  const dstDir = normalizePath(dialog.form.targetPath)
  try {
    await Promise.all(dialog.srcPaths.map((src) => moveFile(src, joinPath(dstDir, basename(src)))))
    message.success('移动成功')
    dialog.move = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || '移动失败') }
}
function openCopy(paths: string[]) {
  dialog.srcPaths = paths
  dialog.form.targetPath = currentPath.value
  dialog.copy = true
}
async function submitCopy() {
  const dstDir = normalizePath(dialog.form.targetPath)
  try {
    await Promise.all(dialog.srcPaths.map((src) => copyFile(src, joinPath(dstDir, basename(src)))))
    message.success('复制成功')
    dialog.copy = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || '复制失败') }
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
  if (!dest) { message.warning('请输入目标路径'); return }
  try {
    await compressFiles(dialog.srcPaths, dest, dialog.form.compressFormat)
    message.success('压缩任务已提交')
    dialog.compress = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) { message.error(e.message || '压缩失败') }
}
function openDecompress(row: FileItem) {
  dialog.target = row
  dialog.form.decompressDest = dirname(row.path)
  dialog.decompress = true
}
async function submitDecompress() {
  if (!dialog.target) return
  const dest = normalizePath(dialog.form.decompressDest)
  try {
    await decompressFile(dialog.target.path, dest)
    message.success('解压任务已提交')
    dialog.decompress = false
    loadList()
  } catch (e: any) { message.error(e.message || '解压失败') }
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
    message.success('删除成功')
    del.show = false
    checkedKeys.value = []
    loadList()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    del.loading = false
  }
}

/* ==================== 下载 ==================== */
async function handleDownload(row: FileItem) {
  try {
    await downloadFileAs(row.path, row.name)
    message.success('开始下载')
  } catch (e: any) {
    message.error(e.message || '下载失败')
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
  if (!upload.items.length) { message.warning('请先选择文件'); return }
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
        item.error = err.message || '上传失败'
      }
    }
    const ok = upload.items.filter((i) => i.status === 'done').length
    message.success(`上传完成 ${ok}/${upload.items.length}`)
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
    message.error(e.message || '读取文件失败')
  } finally {
    editor.loading = false
  }
}
async function saveEditor() {
  if (!editor.path) return
  editor.saving = true
  try {
    await writeFileContent(editor.path, editor.content)
    message.success('保存成功')
  } catch (e: any) {
    message.error(e.message || '保存失败')
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
        <NButton v-if="isMobile" size="small" quaternary circle @click="showTree = !showTree" title="目录树">
          <template #icon><NIcon size="14"><FolderOutline /></NIcon></template>
        </NButton>
        <NInput
          v-model:value="searchKeyword"
          placeholder="搜索当前目录…"
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
          新建文件夹
        </NButton>
        <NButton size="small" @click="openMkfile">
          <template #icon><NIcon size="14"><AddOutline /></NIcon></template>
          新建文件
        </NButton>
        <NButton size="small" @click="openUpload">
          <template #icon><NIcon size="14"><CloudUploadOutline /></NIcon></template>
          上传
        </NButton>
        <NButton size="small" quaternary circle @click="loadList" title="刷新">
          <template #icon><NIcon size="14"><RefreshOutline /></NIcon></template>
        </NButton>
      </NSpace>
    </div>

    <!-- 批量操作条 -->
    <div v-if="checkedKeys.length" class="fm-batchbar">
      <span class="fm-batchbar__count">已选 {{ checkedKeys.length }} 项</span>
      <NSpace size="small">
        <NButton size="tiny" @click="openCopy(checkedPaths)">
          <template #icon><NIcon size="12"><CopyOutline /></NIcon></template>复制
        </NButton>
        <NButton size="tiny" @click="openMove(checkedPaths)">
          <template #icon><NIcon size="12"><MoveOutline /></NIcon></template>移动
        </NButton>
        <NButton size="tiny" @click="openCompress(checkedPaths)">压缩</NButton>
        <NButton size="tiny" type="error" @click="openDeleteByChecked">
          <template #icon><NIcon size="12"><TrashOutline /></NIcon></template>删除
        </NButton>
        <NButton size="tiny" quaternary @click="checkedKeys = []">取消选择</NButton>
      </NSpace>
    </div>

    <!-- 搜索提示 -->
    <div v-if="inSearchMode" class="fm-search-hint">
      <NTag size="small" type="info" :bordered="false">搜索中</NTag>
      <span>关键词「{{ searchKeyword.trim() }}」 · 命中 {{ searchResults.length }} 项</span>
      <a class="fm-search-hint__exit" @click="searchKeyword = ''">退出搜索</a>
    </div>

    <!-- 主体：左树右表 -->
    <div class="fm-body">
      <div class="fm-tree" :class="{ 'fm-tree--hidden': isMobile && !showTree }">
        <div class="fm-tree__title">目录树</div>
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
          v-model:checkedRowKeys="checkedKeys"
          @rowDblclick="handleRowDblclick"
        />
      </div>
    </div>
  </PageCard>

  <!-- 新建文件夹 -->
  <ModalForm v-model:show="dialog.mkdir" title="新建文件夹" @submit="submitMkdir">
    <NFormItem label="文件夹名称" required>
      <NInput v-model:value="dialog.form.dirName" placeholder="例如：logs" @keyup.enter="submitMkdir" />
    </NFormItem>
    <NFormItem label="所在路径">
      <NInput :value="currentPath" readonly />
    </NFormItem>
  </ModalForm>

  <!-- 新建文件 -->
  <ModalForm v-model:show="dialog.mkfile" title="新建文件" @submit="submitMkfile" :width="560">
    <NFormItem label="文件名" required>
      <NInput v-model:value="dialog.form.fileName" placeholder="例如：README.md" @keyup.enter="submitMkfile" />
    </NFormItem>
    <NFormItem label="初始内容（可选）">
      <NInput v-model:value="dialog.form.fileContent" type="textarea" :rows="6" placeholder="文件内容" />
    </NFormItem>
  </ModalForm>

  <!-- 重命名 -->
  <ModalForm v-model:show="dialog.rename" title="重命名" @submit="submitRename">
    <NFormItem label="原名称">
      <NInput :value="dialog.target?.name" readonly />
    </NFormItem>
    <NFormItem label="新名称" required>
      <NInput v-model:value="dialog.form.renameNewName" @keyup.enter="submitRename" />
    </NFormItem>
  </ModalForm>

  <!-- 移动 -->
  <ModalForm v-model:show="dialog.move" title="移动" @submit="submitMove">
    <NFormItem label="源路径">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem label="目标目录" required>
      <NInput v-model:value="dialog.form.targetPath" placeholder="/path/to/dest" />
    </NFormItem>
  </ModalForm>

  <!-- 复制 -->
  <ModalForm v-model:show="dialog.copy" title="复制" @submit="submitCopy">
    <NFormItem label="源路径">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem label="目标目录" required>
      <NInput v-model:value="dialog.form.targetPath" placeholder="/path/to/dest" />
    </NFormItem>
  </ModalForm>

  <!-- 压缩 -->
  <ModalForm v-model:show="dialog.compress" title="压缩" @submit="submitCompress">
    <NFormItem label="源文件">
      <NInput :value="dialog.srcPaths.join(', ')" readonly />
    </NFormItem>
    <NFormItem label="压缩格式">
      <NSelect v-model:value="dialog.form.compressFormat" :options="[
        { label: 'zip', value: 'zip' },
        { label: 'tar.gz', value: 'tar.gz' }
      ]" />
    </NFormItem>
    <NFormItem label="目标文件路径" required>
      <NInput v-model:value="dialog.form.compressDest" placeholder="/path/to/archive.zip" />
    </NFormItem>
  </ModalForm>

  <!-- 解压 -->
  <ModalForm v-model:show="dialog.decompress" title="解压" @submit="submitDecompress">
    <NFormItem label="压缩包">
      <NInput :value="dialog.target?.path" readonly />
    </NFormItem>
    <NFormItem label="目标目录" required>
      <NInput v-model:value="dialog.form.decompressDest" placeholder="/path/to/extract" />
    </NFormItem>
  </ModalForm>

  <!-- 上传 -->
  <ModalForm v-model:show="upload.show" title="上传文件" :loading="upload.uploading" @submit="startUpload" :width="600">
    <NFormItem label="目标目录">
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
      <p class="dropzone__title">拖拽文件到此处，或点击选择</p>
      <p class="dropzone__hint">支持多文件同时上传到当前目录</p>
    </div>
    <div v-if="upload.items.length" class="upload-list">
      <div v-for="(item, idx) in upload.items" :key="item.name + idx" class="upload-item">
        <div class="upload-item__head">
          <span class="upload-item__name">{{ item.name }}</span>
          <span class="upload-item__meta mono">{{ formatSize(item.size) }} · {{ item.progress }}%</span>
          <NButton v-if="item.status !== 'uploading'" size="tiny" text @click="removeUploadItem(idx)">移除</NButton>
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
    title="确认删除"
    :content="`确定删除选中的 ${del.paths.length} 项？此操作不可恢复。`"
    type="error"
    confirm-text="删除"
    :loading="del.loading"
    @confirm="submitDelete"
  />

  <!-- 代码编辑器抽屉 -->
  <NDrawer v-model:show="editor.show" :width="820" placement="right">
    <NDrawerContent :title="`编辑：${editor.name}`" :closable="true">
      <div class="editor-meta">
        <span class="mono editor-meta__path">{{ editor.path }}</span>
        <NTag size="small" :bordered="false">{{ editor.language }}</NTag>
      </div>
      <NAlert
        v-if="editor.size > 1048576"
        type="warning"
        style="margin-bottom: var(--space-3)"
        :title="`文件较大（${formatSize(editor.size)}），编辑可能较慢`"
      />
      <CodeEditor
        v-model:modelValue="editor.content"
        :language="editor.language"
        height="calc(100vh - 280px)"
      />
      <template #footer>
        <NSpace justify="end">
          <NButton @click="editor.show = false">关闭</NButton>
          <NButton type="primary" :loading="editor.saving" @click="saveEditor">
            保存 (Ctrl+S)
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
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
