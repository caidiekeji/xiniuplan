<script setup lang="ts">
import { computed } from 'vue'
import { NModal } from 'naive-ui'

/**
 * ContentModal — 内容展示弹窗（base 层原子组件）
 *
 * 与 ModalForm（表单弹窗）/ ConfirmDialog（确认弹窗）构成弹窗三原子：
 * - 表单 → ModalForm；简单确认 → ConfirmDialog；内容展示（日志/终端/向导/详情）→ ContentModal
 *
 * 内建行为（使用方无需重复拼装）：
 * - 宽度走 --modal-width-* token（sm/md/lg/xl 四档，移动端断点自动降为 vw 钳制）
 * - ESC / 遮罩关闭可配置（默认均可关闭；表单向导类传 :mask-closable="false"）
 * - 焦点陷阱 / 关闭后焦点返回由 Naive UI NModal 内建提供
 * - 双主题背景、圆角、标题栏样式随 token / Naive 主题自动切换
 */
const props = withDefaults(
  defineProps<{
    /** 弹窗显隐（支持 v-model:show） */
    show: boolean
    /** 标题栏文本（标题栏始终显示，内容弹窗以 card 预设呈现） */
    title?: string
    /** 宽度档位：sm=480 / md=560 / lg=800 / xl=960，对应 --modal-width-* token */
    size?: 'sm' | 'md' | 'lg' | 'xl'
    /** 点击遮罩是否关闭（默认 true；含未保存输入的场景传 false） */
    maskClosable?: boolean
    /** ESC 是否关闭（默认 true） */
    escClosable?: boolean
  }>(),
  {
    title: '',
    size: 'md',
    maskClosable: true,
    escClosable: true
  }
)

const emit = defineEmits<{
  (e: 'update:show', val: boolean): void
}>()

const widthStyle = computed(() => ({ width: `var(--modal-width-${props.size})` }))
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    :title="title"
    :style="widthStyle"
    :mask-closable="maskClosable"
    :esc-closable="escClosable"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <slot />
    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </NModal>
</template>
