import { defineAsyncComponent } from 'vue'

export { default as PageCard } from './PageCard.vue'
export { default as DataTable } from './DataTable.vue'
export { default as SearchForm } from './SearchForm.vue'
export { default as ModalForm } from './ModalForm.vue'
export { default as StatusTag } from './StatusTag.vue'
export type { StatusType } from './StatusTag.vue'
export { default as StatCard } from './StatCard.vue'
export { default as ConfirmDialog } from './ConfirmDialog.vue'

// CodeEditor 依赖 monaco-editor（体积大），改为异步加载：
// 只有真正渲染 <CodeEditor> 的页面才会下载 monaco chunk
export const CodeEditor = defineAsyncComponent(() => import('./CodeEditor.vue'))
