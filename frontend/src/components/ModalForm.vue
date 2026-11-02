<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NForm, NButton, NSpace } from 'naive-ui'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    loading?: boolean
    width?: number
    /** 主按钮文案覆盖（默认 common.confirm；新建资源场景传「创建」等词表统一词） */
    confirmText?: string
    /** 取消按钮文案覆盖（默认 common.cancel） */
    cancelText?: string
  }>(),
  {
    title: '',
    loading: false,
    width: 520,
    confirmText: '',
    cancelText: ''
  }
)

const emit = defineEmits<{
  (e: 'update:show', val: boolean): void
  (e: 'submit'): void
  (e: 'cancel'): void
}>()

const visible = ref(props.show)

watch(
  () => props.show,
  (val) => (visible.value = val)
)

const resolvedTitle = computed(() => props.title || t('common.edit'))
const resolvedConfirmText = computed(() => props.confirmText || t('common.confirm'))
const resolvedCancelText = computed(() => props.cancelText || t('common.cancel'))

/* Enter 提交时需要跳过的组件：这些控件内部用 Enter 确认选项（NSelect/NSwitch 等），
   若一并提交表单会造成误触发。 */
const ENTER_SKIP_SELECTOR =
  '.n-base-selection, .n-checkbox, .n-radio, .n-switch, .n-date-picker, .n-cascader, .n-tree-select, .n-time-picker'

function handleClose() {
  emit('update:show', false)
  emit('cancel')
}

function handleSubmit() {
  // 防重复提交：loading 期间忽略后续触发（双击确认 / 回车连击）
  if (props.loading) return
  emit('submit')
}

/**
 * 回车提交。
 * - isComposing：中文输入法选词时的回车不能提交（否则选词即提交）
 * - textarea：多行输入回车应换行
 * - 选择类控件内部：回车用于确认选项
 */
function handleKeydown(e: KeyboardEvent) {
  if (e.key !== 'Enter' || e.isComposing || props.loading) return
  const target = e.target as HTMLElement | null
  if (!target) return
  if (target.tagName === 'TEXTAREA') return
  if (target.closest?.(ENTER_SKIP_SELECTOR)) return
  e.preventDefault()
  handleSubmit()
}
</script>

<template>
  <NModal
    :show="visible"
    :mask-closable="false"
    preset="card"
    :title="resolvedTitle"
    :style="{ width: 'min(' + width + 'px, 92vw)' }"
    @update:show="(v: boolean) => emit('update:show', v)"
    @after-leave="handleClose"
  >
    <NForm label-placement="top" @submit.prevent="handleSubmit" @keydown="handleKeydown">
      <slot />
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">{{ resolvedCancelText }}</NButton>
        <NButton type="primary" :loading="loading" :disabled="loading" @click="handleSubmit">
          {{ resolvedConfirmText }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
