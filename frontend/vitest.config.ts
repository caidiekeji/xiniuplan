import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    // 默认 jsdom（供后续组件测试使用）
    environment: 'jsdom',
    // 纯逻辑测试走 node 环境：不加载 jsdom，避免 undici/jsdom 的
    // webidl.util.markAsUncloneable 兼容性问题（该 API 仅 Node 22+ 提供，
    // Node 20 下加载 jsdom 即抛 TypeError，导致整轮测试无用例执行）。
    environmentMatchGlobs: [
      ['**/__tests__/**/*.test.ts', 'node'],
      ['**/*.spec.ts', 'node']
    ],
    globals: true
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
})
