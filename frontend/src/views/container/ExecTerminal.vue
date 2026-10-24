<template>
  <ContentModal v-model:show="visible" :title="t('container.execTitle', { name })" size="xl">
    <div ref="termEl" class="exec-term" @keydown.ctrl.c.prevent></div>
  </ContentModal>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ContentModal } from '@/components'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { useThemeTokens } from '@/composables/useThemeTokens'

const props = defineProps<{ show: boolean; containerId: string; name: string }>()
const emit = defineEmits<{ (e: 'update:show', v: boolean): void }>()

const { t } = useI18n()

const visible = ref(props.show)
const termEl = ref<HTMLElement | null>(null)

let term: Terminal | null = null
let ws: WebSocket | null = null
let fitAddon: FitAddon | null = null

const authStore = useAuthStore()
const { tokens } = useThemeTokens()

watch(() => props.show, (v) => {
  visible.value = v
  if (v) {
    setTimeout(openTerminal, 100)
  } else {
    cleanup()
  }
})
watch(visible, (v) => emit('update:show', v))

function openTerminal() {
  cleanup()
  if (!termEl.value || !props.containerId) return

  // 字号走 token（--term-font-size），与主终端 Terminal/index.vue 保持一档
  term = new Terminal({
    cursorBlink: true,
    fontSize: parseInt(tokens.value.termFontSize) || 13,
    theme: { background: '#141a24', foreground: '#d8dee9' }
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(termEl.value)
  fitAddon.fit()

  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const base = (window as any).__PANEL_BASE__ || ''
  ws = new WebSocket(`${proto}://${location.host}${base}/api/v1/ws/exec?cid=${encodeURIComponent(props.containerId)}&token=${authStore.token}`)
  ws.binaryType = 'arraybuffer'

  ws.onmessage = (ev) => {
    const data = typeof ev.data === 'string' ? ev.data : new TextDecoder().decode(ev.data)
    try {
      const msg = JSON.parse(data)
      if (msg.type === 'error') {
        term?.writeln('\r\n\x1b[31m' + msg.message + '\x1b[0m')
      }
      // connected 消息静默
      return
    } catch {
      // 原始输出
    }
    term?.write(data)
  }

  ws.onclose = () => {
    term?.writeln('\r\n\x1b[90m' + t('container.execDisconnected') + '\x1b[0m')
  }

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })
}

function cleanup() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (term) {
    term.dispose()
    term = null
  }
}

onBeforeUnmount(cleanup)
</script>

<style scoped>
.exec-term {
  height: 480px;
  background: #141a24;
  border-radius: 6px;
  padding: var(--space-2);
}
</style>
