<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
// 按需引入 monaco：editor.api 仅含编辑器内核（约 1/3 体积），
// 不打包全部语言服务（json/css/html/ts worker 各自数 MB）。
// Monarch 语法高亮内置，nginx/yaml/shell 等高亮不受影响。
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

// 仅注册基础 worker（取消令牌/word-based 建议所需），语言级智能提示按需再加
self.MonacoEnvironment = {
  getWorker() {
    return new EditorWorker()
  }
}

const props = withDefaults(
  defineProps<{
    modelValue?: string
    language?: string
    theme?: string
    readOnly?: boolean
    height?: string
    minimap?: boolean
  }>(),
  {
    modelValue: '',
    language: 'javascript',
    theme: 'vs-dark',
    readOnly: false,
    height: 'var(--editor-height)',
    minimap: true
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const editorRef = ref<HTMLDivElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

onMounted(() => {
  if (!editorRef.value) return
  editor = monaco.editor.create(editorRef.value, {
    value: props.modelValue,
    language: props.language,
    theme: props.theme,
    readOnly: props.readOnly,
    minimap: { enabled: props.minimap },
    fontSize: 13,
    lineNumbers: 'on',
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 2,
    wordWrap: 'on',
    padding: { top: 8, bottom: 8 }
  })

  editor.onDidChangeModelContent(() => {
    emit('update:modelValue', editor?.getValue() || '')
  })
})

watch(
  () => props.modelValue,
  (val) => {
    if (editor && val !== editor.getValue()) {
      editor.setValue(val)
    }
  }
)

watch(
  () => props.language,
  (val) => {
    if (editor) {
      monaco.editor.setModelLanguage(editor.getModel()!, val)
    }
  }
)

onBeforeUnmount(() => {
  editor?.dispose()
  editor = null
})
</script>

<template>
  <div ref="editorRef" class="code-editor" :style="{ height }" />
</template>

<style scoped>
.code-editor {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}
</style>
