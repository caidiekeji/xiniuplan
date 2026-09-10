<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NButton, NSpace, NIcon } from 'naive-ui'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    content?: string
    confirmText?: string
    cancelText?: string
    type?: 'warning' | 'error' | 'info'
    loading?: boolean
  }>(),
  {
    title: '',
    confirmText: '',
    cancelText: '',
    type: 'warning',
    loading: false
  }
)

const emit = defineEmits<{
  (e: 'update:show', val: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const visible = ref(props.show)

watch(
  () => props.show,
  (val) => (visible.value = val)
)

const resolvedTitle = computed(() => props.title || t('common.confirmDelete'))
const resolvedConfirmText = computed(() => props.confirmText || t('common.confirm'))
const resolvedCancelText = computed(() => props.cancelText || t('common.cancel'))

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  emit('update:show', false)
  emit('cancel')
}

const typeColor: Record<string, string> = {
  warning: 'var(--warn)',
  error: 'var(--err)',
  info: 'var(--info)'
}
</script>

<template>
  <NModal
    :show="visible"
    :mask-closable="false"
    preset="dialog"
    :title="resolvedTitle"
    :positive-text="resolvedConfirmText"
    :negative-text="resolvedCancelText"
    :loading="loading"
    @positive-click="handleConfirm"
    @negative-click="handleCancel"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <div class="confirm-dialog__content">
      <span class="led" :class="type === 'error' ? 'led--err' : type === 'warning' ? 'led--warn' : 'led--offline'" />
      <span>{{ content }}</span>
    </div>
  </NModal>
</template>

<style scoped>
.confirm-dialog__content {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  color: var(--sub);
  line-height: var(--lh-body);
  padding-top: var(--space-2);
}
</style>
