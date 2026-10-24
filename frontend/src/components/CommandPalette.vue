<script setup lang="ts">
/**
 * 命令面板（审计 D13）：Cmd/Ctrl+K 全局唤起，键盘优先的功能跳转入口。
 * 菜单项由 DefaultLayout 传入（与侧栏分组同源，单一事实源），
 * 本组件只负责：过滤 / 键盘导航 / 触发 select。
 */
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NIcon } from 'naive-ui'
import { SearchOutline } from '@vicons/ionicons5'

export interface CommandItem {
  key: string
  label: string
  icon?: any
  /** 分组名（概览/应用/运维…），仅用于列表内弱化展示 */
  group?: string
}

const props = defineProps<{ items: CommandItem[]; show?: boolean }>()
const emit = defineEmits<{ (e: 'select', key: string): void; (e: 'update:show', v: boolean): void }>()

const { t } = useI18n()

const innerShow = ref(false)
// 外部可受控（顶栏入口），未受控时用内部状态（Ctrl+K）
const show = computed({
  get: () => props.show ?? innerShow.value,
  set: (v: boolean) => {
    innerShow.value = v
    emit('update:show', v)
  }
})
const query = ref('')
const activeIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

const modalStyle = { width: 'var(--modal-width-sm)', marginTop: '12vh' }

// 过滤：label 与路由 key 双匹配（label 跟随当前语言，key 恒定可盲打）
const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  const list = !q
    ? props.items
    : props.items.filter(
        (i) => i.label.toLowerCase().includes(q) || i.key.toLowerCase().includes(q)
      )
  return list.slice(0, 30)
})

watch(filtered, () => {
  activeIndex.value = 0
})

watch(show, (v) => {
  if (v) {
    query.value = ''
    activeIndex.value = 0
    nextTick(() => inputRef.value?.focus())
  }
})

function move(delta: number) {
  const len = filtered.value.length
  if (!len) return
  activeIndex.value = (activeIndex.value + delta + len) % len
}

function run(key: string) {
  show.value = false
  emit('select', key)
}

function runActive() {
  const item = filtered.value[activeIndex.value]
  if (item) run(item.key)
}

// 全局快捷键：Ctrl+K（Win/Linux）/ Cmd+K（macOS）
function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    show.value = !show.value
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <NModal
    v-model:show="show"
    preset="card"
    :style="modalStyle"
    :bordered="false"
    :mask-closable="true"
    :close-on-esc="true"
    class="command-palette"
  >
    <div class="palette-search">
      <NIcon class="palette-search__icon"><SearchOutline /></NIcon>
      <input
        ref="inputRef"
        v-model="query"
        class="palette-search__input"
        :placeholder="t('common.commandPlaceholder')"
        @keydown.down.prevent="move(1)"
        @keydown.up.prevent="move(-1)"
        @keydown.enter.prevent="runActive"
      />
      <kbd class="palette-search__kbd">esc</kbd>
    </div>

    <div class="palette-list" @mouseleave="activeIndex = activeIndex">
      <template v-if="filtered.length">
        <button
          v-for="(item, i) in filtered"
          :key="item.key"
          type="button"
          class="palette-item"
          :class="{ 'palette-item--active': i === activeIndex }"
          @mousemove="activeIndex = i"
          @click="run(item.key)"
        >
          <span v-if="item.icon" class="palette-item__icon">
            <component :is="item.icon" />
          </span>
          <span class="palette-item__label">{{ item.label }}</span>
          <span v-if="item.group" class="palette-item__group">{{ item.group }}</span>
        </button>
      </template>
      <div v-else class="palette-empty">{{ t('common.noResults') }}</div>
    </div>
  </NModal>
</template>

<style scoped>
.palette-search {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-bottom: 1px solid var(--line);
}

.palette-search__icon {
  color: var(--text-quaternary);
  flex-shrink: 0;
}

.palette-search__input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: var(--fs-body);
  font-family: inherit;
  color: var(--ink);
}

.palette-search__input::placeholder {
  color: var(--text-quaternary);
}

.palette-search__kbd {
  font-family: var(--font-mono);
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  padding: var(--space-0) var(--space-2);
  flex-shrink: 0;
}

.palette-list {
  max-height: 40vh;
  overflow-y: auto;
  padding: var(--space-2);
}

.palette-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-2) var(--space-3);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
}

.palette-item--active {
  background: var(--acc-soft);
}

.palette-item__icon {
  display: inline-flex;
  color: var(--sub);
  flex-shrink: 0;
}

.palette-item--active .palette-item__icon {
  color: var(--acc);
}

.palette-item__label {
  flex: 1;
  font-size: var(--fs-body);
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.palette-item__group {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  flex-shrink: 0;
}

.palette-empty {
  padding: var(--space-5);
  text-align: center;
  font-size: var(--fs-body);
  color: var(--text-quaternary);
}
</style>
