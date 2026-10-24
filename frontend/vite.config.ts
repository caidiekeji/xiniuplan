import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg'],
      manifest: {
        name: '犀牛面板 RhinoPanel',
        short_name: 'RhinoPanel',
        description: '服务器管理面板',
        theme_color: '#0e1117',
        icons: [
          { src: '/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icon-512.png', sizes: '512x512', type: 'image/png' }
        ]
      },
      workbox: {
        maximumFileSizeToCacheInBytes: 5 * 1024 * 1024,
        // 只预缓存静态壳资源（几 KB）：默认行为会把全部 134 个构建产物
        // （含 monaco 2.3MB、echarts、xterm、数十个图标 chunk）在 SW install
        // 阶段全部 precache，首次访问时与首屏资源抢带宽——实测把 naive-ui
        // 从 240ms 拖到 2.5s（2026-10-22）。
        globPatterns: ['manifest.webmanifest', 'favicon.svg', 'icons/*'],
        runtimeCaching: [
          {
            // 导航请求：NetworkFirst，保证 HTML 新鲜，离线时回退缓存
            urlPattern: ({ request }) => request.mode === 'navigate',
            handler: 'NetworkFirst',
            options: {
              cacheName: 'panel-pages',
              networkTimeoutSeconds: 3,
              expiration: { maxEntries: 10, maxAgeSeconds: 60 * 60 * 24 * 7 }
            }
          },
          {
            // JS/CSS 资产（含所有路由级懒加载 chunk）：StaleWhileRevalidate，
            // 首次访问网络下载并后台入库，二次访问直接命中缓存
            urlPattern: /\/assets\/.*\.(?:js|css|woff2?)$/,
            handler: 'StaleWhileRevalidate',
            options: {
              cacheName: 'panel-assets',
              cacheableResponse: { statuses: [0, 200] },
              expiration: { maxEntries: 300, maxAgeSeconds: 60 * 60 * 24 * 30 }
            }
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8888',
        changeOrigin: true,
        ws: true
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        // manualChunks 的坑（2026-10-22 / 10-23 两次踩中）：
        // 显式列出包名会把该包的**全部模块**强行打进同一个 chunk，
        // 从而破坏按需引入与 tree-shaking：
        //   1. monaco-editor → 入口静态依赖 monaco，index.html 被注入
        //      2.3MB modulepreload + 133KB 同步 CSS（首屏被编辑器占满）
        //   2. echarts → 页面只用 LineChart，却被打进整个 1.1MB 全量包，
        //      初始化执行耗时数秒（仪表盘图表迟迟不出）
        //   3. naive-ui → 同理，整包强制合并，摇树失效
        // 结论：只保留确实需要整体分组的 vue 运行时，其余交给 Rollup 按
        // 实际 import 关系自然分包（大数据依赖会自动拆成独立懒加载 chunk）。
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router', 'pinia']
        }
      }
    }
  }
})
