<script setup lang="ts">
// 服务端目录浏览器弹窗：用于"网站根目录"等路径字段的点选式输入。
// 复用文件服务的 GET /api/v1/files（自带 JWT + 权限 + safePath 防穿越），
// 只展示子目录，点击进入、面包屑回退，底部确认选中当前目录。
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NButton, NSpace, NIcon, NEmpty, NSpin } from 'naive-ui'
import { FolderOutline, ChevronUpOutline, RefreshOutline } from '@vicons/ionicons5'
import { listFiles, type FileItem } from '@/api/file'
import { useMessage } from '@/composables/useMessage'

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    /** 打开时定位到的初始目录（默认根目录） */
    initialPath?: string
  }>(),
  {
    title: '',
    initialPath: '/'
  }
)

const emit = defineEmits<{
  (e: 'update:show', val: boolean): void
  (e: 'select', path: string): void
}>()

const { t } = useI18n()
const message = useMessage()

const currentPath = ref('/')
const entries = ref<FileItem[]>([])
const loading = ref(false)

const dirs = computed(() => entries.value.filter((e) => e.is_dir))

/** 面包屑：'/' + 逐级累积路径 */
const crumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const list = [{ label: '/', path: '/' }]
  let acc = ''
  for (const p of parts) {
    acc += '/' + p
    list.push({ label: p, path: acc })
  }
  return list
})

/** 当前目录是否存在上一级（根目录时不可再向上） */
const canGoUp = computed(() => currentPath.value !== '/')

function parentPath(path: string): string {
  const idx = path.replace(/\/+$/, '').lastIndexOf('/')
  if (idx <= 0) return '/'
  return path.slice(0, idx)
}

async function navigate(path: string) {
  loading.value = true
  try {
    const res = await listFiles(path)
    currentPath.value = res.data?.path || path
    entries.value = res.data?.items || []
  } catch (err: any) {
    entries.value = []
    message.error(err?.message || t('common.dirPicker.loadFailed'))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      const init = props.initialPath?.trim()
      // 初始路径若指向文件/不存在，后端会报错；定位到其父目录更稳妥
      navigate(init && init !== '/' ? init : '/')
    }
  }
)

function handleSelect() {
  emit('select', currentPath.value)
  emit('update:show', false)
}

function handleClose() {
  emit('update:show', false)
}
</script>

<template>
  <NModal
    :show="show"
    :mask-closable="false"
    preset="card"
    :title="title || t('common.dirPicker.title')"
    :style="{ width: 'min(560px, 92vw)' }"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <!-- 工具栏：上一级 / 刷新 -->
    <div class="dir-picker__toolbar">
      <NSpace size="small">
        <NButton size="small" quaternary :disabled="!canGoUp || loading" @click="navigate(parentPath(currentPath))">
          <template #icon>
            <NIcon><ChevronUpOutline /></NIcon>
          </template>
          {{ t('common.dirPicker.parent') }}
        </NButton>
        <NButton size="small" quaternary :loading="loading" @click="navigate(currentPath)">
          <template #icon>
            <NIcon><RefreshOutline /></NIcon>
          </template>
          {{ t('common.dirPicker.refresh') }}
        </NButton>
      </NSpace>
    </div>

    <!-- 面包屑路径 -->
    <div class="dir-picker__crumbs mono">
      <template v-for="(c, i) in crumbs" :key="c.path">
        <span v-if="i > 0" class="dir-picker__crumb-sep">/</span>
        <span
          class="dir-picker__crumb"
          :class="{ 'dir-picker__crumb--active': i === crumbs.length - 1 }"
          @click="!loading && navigate(c.path)"
        >{{ c.label }}</span>
      </template>
    </div>

    <!-- 子目录列表 -->
    <div class="dir-picker__list">
      <NSpin :show="loading">
        <NEmpty
          v-if="!loading && dirs.length === 0"
          :description="t('common.dirPicker.empty')"
          style="padding: var(--space-6) 0"
        />
        <div
          v-for="d in dirs"
          :key="d.path"
          class="dir-picker__row"
          @click="navigate(d.path)"
        >
          <NIcon class="dir-picker__row-icon" :size="18"><FolderOutline /></NIcon>
          <span class="dir-picker__row-name mono">{{ d.name }}</span>
        </div>
      </NSpin>
    </div>

    <template #footer>
      <div class="dir-picker__footer">
        <span class="dir-picker__current mono" :title="currentPath">{{ currentPath }}</span>
        <NSpace>
          <NButton @click="handleClose">{{ t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="loading" @click="handleSelect">
            {{ t('common.dirPicker.selectCurrent') }}
          </NButton>
        </NSpace>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.dir-picker__toolbar {
  margin-bottom: var(--space-2);
}

.dir-picker__crumbs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
  padding: var(--space-2) var(--space-3);
  margin-bottom: var(--space-2);
  background: var(--bg-subtle, rgba(0, 0, 0, 0.03));
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  font-size: var(--fs-caption);
  word-break: break-all;
}

.dir-picker__crumb {
  cursor: pointer;
  color: var(--sub);
  padding: 0 2px;
  border-radius: 4px;
  transition: color var(--t-fast) var(--ease);
}

.dir-picker__crumb:hover {
  color: var(--acc);
}

.dir-picker__crumb--active {
  color: var(--ink);
  font-weight: 600;
  cursor: default;
}

.dir-picker__crumb-sep {
  color: var(--text-quaternary);
}

.dir-picker__list {
  min-height: 200px;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
}

.dir-picker__row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
}

.dir-picker__row:hover {
  background: var(--bg-hover, rgba(0, 0, 0, 0.04));
}

.dir-picker__row-icon {
  color: var(--acc);
  flex-shrink: 0;
}

.dir-picker__row-name {
  font-size: var(--fs-caption);
  color: var(--ink);
  word-break: break-all;
}

.dir-picker__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.dir-picker__current {
  font-size: var(--fs-caption);
  color: var(--sub);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
