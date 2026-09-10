<script setup lang="ts">
import { NForm, NFormItem, NButton, NSpace } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'search', values: Record<string, any>): void
  (e: 'reset'): void
}>()

function handleSearch(e: Event) {
  e.preventDefault()
  const formData = new FormData(e.target as HTMLFormElement)
  const values: Record<string, any> = {}
  formData.forEach((v, k) => (values[k] = v))
  emit('search', values)
}

function handleReset() {
  emit('reset')
}
</script>

<template>
  <form class="search-form" @submit="handleSearch">
    <NForm inline :label-width="80">
      <slot />
      <NFormItem>
        <NSpace>
          <NButton type="primary" html-type="submit" :loading="loading">
            {{ t('common.search') }}
          </NButton>
          <NButton @click="handleReset">{{ t('common.reset') }}</NButton>
        </NSpace>
      </NFormItem>
    </NForm>
  </form>
</template>

<style scoped>
.search-form {
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
}

/* 移动端：表单项垂直排列 */
@media (max-width: 768px) {
  .search-form :deep(.n-form) {
    flex-direction: column;
    align-items: stretch;
  }
  .search-form :deep(.n-form-item) {
    margin-right: 0;
    margin-bottom: var(--space-2);
  }
  .search-form :deep(.n-form-item-label) {
    width: 70px !important;
  }
  .search-form :deep(.n-input),
  .search-form :deep(.n-select),
  .search-form :deep(.n-date-picker) {
    width: 100% !important;
  }
}
</style>
