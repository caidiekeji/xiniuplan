<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag } from 'naive-ui'

export type StatusType = 'running' | 'stopped' | 'error' | 'warning' | 'pending'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    status: StatusType
    text?: string
  }>(),
  {}
)

const statusMap = computed(() => ({
  running: { text: t('common.enabled'), type: 'success' as const, led: 'led--ok' },
  stopped: { text: t('common.disabled'), type: 'default' as const, led: 'led--offline' },
  error: { text: t('common.failed'), type: 'error' as const, led: 'led--err' },
  warning: { text: t('common.warning'), type: 'warning' as const, led: 'led--warn' },
  pending: { text: t('common.loading'), type: 'info' as const, led: 'led--warn' }
}))

const config = computed(() => statusMap.value[props.status] || statusMap.value.stopped)
const displayText = computed(() => props.text || config.value.text)
</script>

<template>
  <span class="status-tag">
    <span class="led" :class="config.led" />
    <NTag :type="config.type" size="small" :bordered="false" round>
      {{ displayText }}
    </NTag>
  </span>
</template>

<style scoped>
.status-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}
</style>
