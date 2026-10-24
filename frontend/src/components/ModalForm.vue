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

function handleClose() {
  emit('update:show', false)
  emit('cancel')
}

function handleSubmit() {
  emit('submit')
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
    <NForm label-placement="top">
      <slot />
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">{{ resolvedCancelText }}</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          {{ resolvedConfirmText }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
