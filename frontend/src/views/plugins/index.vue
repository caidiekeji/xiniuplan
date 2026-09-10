<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NIcon, NEmpty, NSpin, NGrid, NGi,
  NCard, NFormItem, NSwitch
} from 'naive-ui'
import {
  SettingsOutline, PlayOutline, StopOutline,
  ExtensionPuzzleOutline
} from '@vicons/ionicons5'
import {
  PageCard, StatusTag, ModalForm, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import * as pluginApi from '@/api/plugin'
import type { PluginItem } from '@/api/plugin'

const message = useMessage()

/* ==================== 状态 ==================== */

const plugins = ref<PluginItem[]>([])
const loading = ref(false)

/* ==================== 配置对话框 ==================== */

const configModalVisible = ref(false)
const configLoading = ref(false)
const currentPlugin = ref<PluginItem | null>(null)
const configText = ref('{}')

/* ==================== 工具函数 ==================== */

/** 根据插件名生成确定性颜色 */
const iconColors = [
  '#1BA784', '#2DD4A8', '#60A5FA', '#E85D4D', '#A78BFA',
  '#F472B6', '#34D399', '#4AC7A5', '#38BDF8', '#FB7185'
]

function getIconColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return iconColors[Math.abs(hash) % iconColors.length]
}

function mapStatus(p: PluginItem): StatusType {
  return p.enabled ? 'running' : 'stopped'
}

function statusText(p: PluginItem): string {
  if (p.running) return '运行中'
  return p.enabled ? '已启用' : '已禁用'
}

/* ==================== 数据加载 ==================== */

async function loadPlugins() {
  loading.value = true
  try {
    const res = await pluginApi.getPlugins()
    plugins.value = res.data.items
  } catch (e: any) {
    message.error(e.message || '加载插件列表失败')
    plugins.value = []
  } finally {
    loading.value = false
  }
}

/* ==================== 启用/禁用 ==================== */

const toggling = ref<Set<string>>(new Set())

async function togglePlugin(p: PluginItem) {
  if (toggling.value.has(p.name)) return
  toggling.value.add(p.name)
  try {
    if (p.enabled) {
      await pluginApi.disablePlugin(p.name)
      message.success(`插件「${p.name}」已禁用`)
    } else {
      await pluginApi.enablePlugin(p.name)
      message.success(`插件「${p.name}」已启用`)
    }
    await loadPlugins()
  } catch (e: any) {
    message.error(e.message || '操作失败')
  } finally {
    toggling.value.delete(p.name)
  }
}

/* ==================== 配置对话框 ==================== */

async function openConfig(p: PluginItem) {
  currentPlugin.value = p
  configText.value = '{}'
  configModalVisible.value = true
  try {
    const res = await pluginApi.getPluginConfig(p.name)
    configText.value = res.data.raw || '{}'
  } catch (e: any) {
    message.error(e.message || '加载配置失败')
  }
}

async function submitConfig() {
  if (!currentPlugin.value) return
  configLoading.value = true
  try {
    // 校验 JSON 格式
    let configObj: Record<string, any>
    try {
      configObj = JSON.parse(configText.value)
    } catch {
      message.error('JSON 格式错误，请检查配置内容')
      return
    }
    await pluginApi.updatePluginConfig(currentPlugin.value.name, configObj)
    message.success('配置已保存')
    configModalVisible.value = false
  } catch (e: any) {
    message.error(e.message || '保存配置失败')
  } finally {
    configLoading.value = false
  }
}

/* ==================== 计算属性 ==================== */

const enabledCount = computed(() => plugins.value.filter(p => p.enabled).length)

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadPlugins()
})
</script>

<template>
  <PageCard
    title="插件管理"
    :subtitle="`已安装 ${plugins.length} 个插件 · 已启用 ${enabledCount} 个`"
  >
    <template #extra>
      <NButton @click="loadPlugins" :loading="loading" size="small">
        <template #icon>
          <NIcon><ExtensionPuzzleOutline /></NIcon>
        </template>
        刷新
      </NButton>
    </template>

    <NSpin :show="loading">
      <!-- 插件卡片网格 -->
      <div v-if="plugins.length > 0" class="plugin-grid">
        <NCard
          v-for="p in plugins"
          :key="p.name"
          class="plugin-card"
          hoverable
        >
          <!-- 卡片头部：图标 + 名称 + 状态 -->
          <div class="plugin-card__header">
            <div
              class="plugin-card__icon"
              :style="{ background: getIconColor(p.name) }"
            >
              {{ (p.icon || p.name.charAt(0)).toUpperCase() }}
            </div>
            <div class="plugin-card__meta">
              <div class="plugin-card__name">{{ p.name }}</div>
              <div class="plugin-card__version mono">v{{ p.version }}</div>
            </div>
            <StatusTag :status="mapStatus(p)" :text="statusText(p)" />
          </div>

          <!-- 描述 -->
          <p class="plugin-card__desc">{{ p.description }}</p>

          <!-- 作者 -->
          <div class="plugin-card__author">
            <NIcon :size="12"><ExtensionPuzzleOutline /></NIcon>
            <span>{{ p.author || 'unknown' }}</span>
          </div>

          <!-- 操作区 -->
          <div class="plugin-card__footer">
            <div class="plugin-card__toggle">
              <span class="plugin-card__toggle-label">
                {{ p.enabled ? '已启用' : '已禁用' }}
              </span>
              <div :title="p.enabled ? '点击禁用' : '点击启用'">
                <NSwitch
                  :value="p.enabled"
                  :loading="toggling.has(p.name)"
                  @update:value="() => togglePlugin(p)"
                />
              </div>
            </div>
            <NButton
              size="small"
              quaternary
              @click="openConfig(p)"
            >
              <template #icon>
                <NIcon><SettingsOutline /></NIcon>
              </template>
              配置
            </NButton>
          </div>
        </NCard>
      </div>

      <!-- 空态 -->
      <div v-else-if="!loading" class="empty-hint">
        <NEmpty description="暂无已安装插件">
          <template #icon>
            <NIcon :size="48"><ExtensionPuzzleOutline /></NIcon>
          </template>
        </NEmpty>
      </div>
    </NSpin>

    <!-- 配置对话框 -->
    <ModalForm
      v-model:show="configModalVisible"
      :title="currentPlugin ? `配置插件 - ${currentPlugin.name}` : '插件配置'"
      :loading="configLoading"
      :width="640"
      @submit="submitConfig"
    >
      <NFormItem label="插件配置 (JSON)">
        <CodeEditor
          v-model="configText"
          language="json"
          height="320px"
          :minimap="false"
        />
      </NFormItem>
      <div class="config-hint">
        配置以 JSON 格式存储，保存后若插件正在运行将热更新配置。
      </div>
    </ModalForm>
  </PageCard>
</template>

<style scoped>
/* 插件卡片网格 - 响应式 CSS Grid */
.plugin-grid {
  margin-top: var(--space-2);
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

/* 平板：2列 */
@media (max-width: 1024px) {
  .plugin-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
  }
}

/* 手机：1列 */
@media (max-width: 640px) {
  .plugin-grid {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}

.plugin-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: transform var(--t-fast) var(--ease), border-color var(--t-fast) var(--ease);
}

.plugin-card:hover {
  transform: translateY(-2px);
}

.plugin-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: var(--space-4);
}

.plugin-card__header {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.plugin-card__icon {
  width: var(--size-md);
  height: var(--size-md);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--fs-headline);
  font-weight: 700;
  color: var(--on-color);
  flex-shrink: 0;
}

.plugin-card__meta {
  flex: 1;
  min-width: 0;
}

.plugin-card__name {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
  line-height: 1.3;
}

.plugin-card__version {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: 2px;
}

.plugin-card__desc {
  font-size: var(--fs-body);
  color: var(--sub);
  line-height: var(--lh-body);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: var(--space-3);
  flex: 1;
}

.plugin-card__author {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-bottom: var(--space-3);
}

.plugin-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: var(--space-3);
  border-top: 1px solid var(--line);
}

.plugin-card__toggle {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.plugin-card__toggle-label {
  font-size: var(--fs-caption);
  color: var(--sub);
}

/* 移动端卡片优化 */
@media (max-width: 640px) {
  .plugin-card :deep(.n-card__content) {
    padding: var(--space-3);
  }

  .plugin-card__header {
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }

  .plugin-card__icon {
    width: var(--size-sm);
    height: var(--size-sm);
    font-size: var(--fs-title);
  }

  .plugin-card__name {
    font-size: var(--fs-body);
  }

  .plugin-card__desc {
    font-size: var(--fs-caption);
  }
}

.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

.config-hint {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: var(--space-2);
}

/* NCard 覆盖 */
:deep(.n-card) {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
}

:deep(.n-card::before) {
  border: none;
}

/* 小屏：单列 */
@media (max-width: 480px) {
  .plugin-grid {
    grid-template-columns: 1fr;
  }
}
</style>
