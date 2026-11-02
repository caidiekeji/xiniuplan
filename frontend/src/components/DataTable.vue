<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NDataTable, NSpin, NEmpty, NButton, NResult } from 'naive-ui'
import type { DataTableColumns, DataTableProps } from 'naive-ui'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    columns: DataTableColumns<any>
    data: any[]
    loading?: boolean
    pagination?: DataTableProps['pagination'] | false
    bordered?: boolean
    striped?: boolean
    size?: 'small' | 'medium' | 'large'
    rowKey?: (row: any) => string | number
    checkedRowKeys?: (string | number)[]
    virtualScroll?: boolean
    maxHeight?: number | string
    scrollX?: number | string
    /** 加载失败时的错误文案；非空则渲染错误态（含重试按钮），优先于空态 */
    error?: string | null
  }>(),
  {
    loading: false,
    bordered: false,
    striped: true,
    size: 'medium',
    checkedRowKeys: undefined,
    virtualScroll: false,
    maxHeight: undefined,
    scrollX: 700
  }
)

const emit = defineEmits<{
  (e: 'update:checkedRowKeys', val: (string | number)[]): void
  (e: 'rowDblclick', row: any, index: number): void
  (e: 'rowContextmenu', row: any, index: number, event: MouseEvent): void
  (e: 'retry'): void
}>()

// 后端空列表可能序列化为 null（Go nil slice），兜底为空数组防止渲染崩溃
const safeData = computed(() => (Array.isArray(props.data) ? props.data : []))
const safeColumns = computed(() => (Array.isArray(props.columns) ? props.columns : []))

const internalPagination = computed(() => {
  if (props.pagination === false) return false
  const base = {
    pageSize: 20,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100]
  }
  if (typeof props.pagination === 'object') {
    return Object.assign({}, base, props.pagination)
  }
  return base
})

const resolvedRowKey = computed(() => {
  return props.rowKey || ((row: any) => row.id || row.key || Math.random())
})

function handleCheckedRowKeysChange(val: (string | number)[]) {
  emit('update:checkedRowKeys', val)
}
</script>

<template>
  <div class="data-table">
    <NSpin :show="loading && !error">
      <!--
        错误态独立成面板（而非塞进空态插槽）：
        1. 空表格无法区分「没有数据」与「加载失败」，用户只能靠一闪而过的 toast 判断；
        2. 空态插槽常被页面自定义覆盖，错误态放在插槽里会被吞掉；
        3. 独立面板可统一承载「重试」入口，让失败可恢复。
      -->
      <div v-if="error" class="data-table__error">
        <NResult status="error" :title="t('common.loadFailed')" :description="error">
          <template #footer>
            <NButton size="small" type="primary" @click="emit('retry')">
              {{ t('common.retry') }}
            </NButton>
          </template>
        </NResult>
      </div>

      <NDataTable
        v-else
        :columns="safeColumns"
        :data="safeData"
        :pagination="internalPagination"
        :bordered="bordered"
        :striped="striped"
        :size="size"
        :row-key="resolvedRowKey"
        :checked-row-keys="checkedRowKeys"
        :virtual-scroll="virtualScroll"
        :max-height="maxHeight"
        :scroll-x="scrollX"
        @update:checked-row-keys="handleCheckedRowKeysChange"
        @row-dblclick="(row: any, index: number) => emit('rowDblclick', row, index)"
        @row-contextmenu="(row: any, index: number, e: MouseEvent) => emit('rowContextmenu', row, index, e)"
      >
        <template #empty>
          <!-- 无错误时的常规空态（错误态已在上方独立处理） -->
          <slot name="empty" :error="null">
            <NEmpty :description="t('common.noData')" />
          </slot>
        </template>
      </NDataTable>
    </NSpin>
  </div>
</template>

<style scoped>
.data-table {
  width: 100%;
}

/* 错误态面板：占位高度接近表格首屏，避免加载失败时页面塌陷 */
.data-table__error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 220px;
  padding: var(--space-4) var(--space-3);
}

.data-table__error :deep(.n-result) {
  padding: var(--space-4) 0;
}

.data-table :deep(.n-data-table) {
  background: transparent;
  font-size: var(--fs-body);
}

.data-table :deep(.n-data-table-th) {
  background: var(--surface2);
  color: var(--sub);
  font-weight: 600;
  font-size: var(--fs-caption);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.data-table :deep(.n-data-table-td) {
  color: var(--ink);
  white-space: nowrap;
}

.data-table :deep(.n-data-table-tr:hover .n-data-table-td) {
  background: var(--surface2);
}

/* 移动端：表格横向滚动；压缩样式单点定义于此（审计 D25，勿在 main.css 重复） */
@media (max-width: 768px) {
  .data-table :deep(.n-data-table-th),
  .data-table :deep(.n-data-table-td) {
    padding: var(--space-2) var(--space-3) !important;
    font-size: 12px !important;
  }
}

@media (max-width: 480px) {
  .data-table :deep(.n-data-table-th),
  .data-table :deep(.n-data-table-td) {
    padding: var(--space-2) var(--space-2) !important;
    font-size: 12px !important;
  }
}
</style>
