<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NCard, NForm, NFormItem, NButton, NSpace } from 'naive-ui'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    loading?: boolean
    width?: number
  }>(),
  {
    title: '',
    loading: false,
    width: 520
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
        <NButton @click="handleClose">{{ t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          {{ t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
