<script setup lang="ts">
import { ref, onErrorCaptured, h, defineComponent } from 'vue'
import { NAlert, NButton, NSpace, NThing } from 'naive-ui'

interface Props {
  /** 错误时显示的标题 */
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '页面渲染出错'
})

const hasError = ref(false)
const errorMessage = ref('')
const errorStack = ref('')
const retryKey = ref(0)

onErrorCaptured((err, instance, info) => {
  window.console.error('[ErrorBoundary] Captured error:', err, info)
  hasError.value = true
  errorMessage.value = err instanceof Error ? err.message : String(err)
  errorStack.value = err instanceof Error ? err.stack || '' : ''
  // 返回 false 阻止错误继续向上传播，避免白屏
  return false
})

function handleRetry() {
  hasError.value = false
  errorMessage.value = ''
  errorStack.value = ''
  retryKey.value += 1
}

// 用一个内部组件来承载默认插槽，便于通过 key 强制重挂载
const SlotHost = defineComponent({
  setup(_, { slots }) {
    return () => slots.default?.()
  }
})
</script>

<template>
  <div class="error-boundary">
    <template v-if="hasError">
      <NAlert type="error" class="error-boundary__alert">
        <template #header>
          <div class="error-boundary__header">
            <span>{{ title }}</span>
            <NButton size="small" type="error" @click="handleRetry">重试</NButton>
          </div>
        </template>
        <NSpace vertical size="small">
          <div class="error-boundary__msg">{{ errorMessage }}</div>
          <NThing v-if="errorStack" class="error-boundary__stack">
            <pre>{{ errorStack }}</pre>
          </NThing>
        </NSpace>
      </NAlert>
    </template>
    <SlotHost v-else :key="retryKey">
      <slot />
    </SlotHost>
  </div>
</template>

<style scoped>
.error-boundary {
  width: 100%;
  height: 100%;
}

.error-boundary__alert {
  margin: var(--space-4);
}

.error-boundary__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.error-boundary__msg {
  font-size: var(--fs-body);
  color: var(--ink);
  word-break: break-all;
}

.error-boundary__stack pre {
  font-family: var(--font-mono);
  font-size: var(--fs-caption);
  color: var(--sub);
  background: var(--surface2);
  padding: var(--space-3);
  border-radius: var(--radius-sm);
  overflow: auto;
  max-height: 300px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
