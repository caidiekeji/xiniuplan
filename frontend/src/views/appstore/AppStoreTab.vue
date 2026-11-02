<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import { NButton, NIcon, NInput, NTag, NEmpty, NSpin, NCard, NSelect } from 'naive-ui'
import { SearchOutline, DownloadOutline } from '@vicons/ionicons5'
import * as appstoreApi from '@/api/appstore'
import type { App, AppCategory, AppSource } from '@/api/appstore'
import {
  getIconColor, categoryLabel, categoryTagType, appSourceTag
} from './shared'

const { t } = useI18n()
const message = useMessage()

/* ==================== Props / Emits ==================== */

const props = defineProps<{
  sources: AppSource[]
  sourcesLoading: boolean
}>()

/* ==================== 应用列表 ==================== */

const apps = ref<App[]>([])
const appsLoading = ref(false)
const selectedCategory = ref<AppCategory | 'all'>('all')
const selectedSource = ref<string>('all')
const searchInput = ref('')
const searchDebounced = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

const categories = computed<{ label: string; value: AppCategory | 'all' }[]>(() => [
  { label: t('appstore.storeTab.allCategories'), value: 'all' },
  { label: t('appstore.categories.website'), value: 'website' },
  { label: t('appstore.categories.devtools'), value: 'devtools' },
  { label: t('appstore.categories.database'), value: 'database' },
  { label: t('appstore.categories.middleware'), value: 'middleware' },
  { label: t('appstore.categories.monitor'), value: 'monitor' },
  { label: t('appstore.categories.office'), value: 'office' },
  { label: t('appstore.categories.storage'), value: 'storage' },
  { label: t('appstore.categories.network'), value: 'network' }
])

const sourceMap = computed(() => {
  const m: Record<string, AppSource> = {}
  props.sources.forEach((s) => {
    m[s.id] = s
  })
  return m
})

/** 来源筛选选项：分组（官方源 / 社区源 / 私有源） */
const sourceFilterOptions = computed(() => {
  const opts: any[] = [{ label: t('appstore.storeTab.allSources'), value: 'all' }]
  const official = props.sources.filter((s) => s.type === 'official')
  const community = props.sources.filter((s) => s.type === 'community')
  const private_ = props.sources.filter((s) => s.type === 'private')
  if (official.length) {
    opts.push({
      type: 'group',
      label: t('appstore.storeTab.officialGroup'),
      children: official.map((s) => ({ label: s.name, value: s.id }))
    })
  }
  if (community.length) {
    opts.push({
      type: 'group',
      label: t('appstore.storeTab.communityGroup'),
      children: community.map((s) => ({ label: s.name, value: s.id }))
    })
  }
  if (private_.length) {
    opts.push({
      type: 'group',
      label: t('appstore.storeTab.privateGroup'),
      children: private_.map((s) => ({ label: s.name, value: s.id }))
    })
  }
  return opts
})

const filteredApps = computed(() => {
  let list = apps.value
  if (selectedCategory.value !== 'all') {
    list = list.filter((a) => a.category === selectedCategory.value)
  }
  if (searchDebounced.value.trim()) {
    const kw = searchDebounced.value.trim().toLowerCase()
    list = list.filter(
      (a) =>
        a.name.toLowerCase().includes(kw) ||
        a.key.toLowerCase().includes(kw) ||
        a.description.toLowerCase().includes(kw)
    )
  }
  return list
})

watch(searchInput, (val) => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    searchDebounced.value = val
  }, 300)
})

watch(selectedSource, () => {
  loadApps()
})

async function loadApps() {
  appsLoading.value = true
  try {
    const sourceId = selectedSource.value !== 'all' ? selectedSource.value : undefined
    const res = await appstoreApi.getApps(undefined, undefined, sourceId)
    apps.value = res.data.items
  } catch (e: any) {
    message.error(e.message || t('appstore.storeTab.loadFailed'))
    apps.value = []
  } finally {
    appsLoading.value = false
  }
}

function selectCategory(cat: AppCategory | 'all') {
  selectedCategory.value = cat
}

// NTabPane 默认 display-directive=if：每次激活本页都会重新挂载，
// 必须在挂载时自行加载，否则切换回来是空列表
onMounted(() => {
  loadApps()
})

defineExpose({ loadApps })
</script>

<template>
  <div class="app-store-tab">
    <!-- 筛选栏 -->
    <div class="store-filter">
      <div class="store-filter__tags">
        <NTag
          v-for="cat in categories"
          :key="cat.value"
          :type="selectedCategory === cat.value ? 'warning' : 'default'"
          :bordered="selectedCategory !== cat.value"
          size="small"
          round
          class="cat-tag"
          @click="selectCategory(cat.value)"
        >
          {{ cat.label }}
        </NTag>
      </div>
      <div class="store-filter__right">
        <NSelect
          v-model:value="selectedSource"
          :options="sourceFilterOptions"
          :placeholder="t('appstore.storeTab.sourcePlaceholder')"
          size="small"
          style="width: var(--input-xl)"
        />
        <NInput
          v-model:value="searchInput"
          :placeholder="t('appstore.storeTab.searchPlaceholder')"
          clearable
          size="small"
          style="width: var(--input-2xl)"
        >
          <template #prefix>
            <NIcon :size="14"><SearchOutline /></NIcon>
          </template>
        </NInput>
      </div>
    </div>

    <!-- 应用卡片网格 -->
    <NSpin :show="appsLoading">
      <div v-if="filteredApps.length > 0" class="app-grid">
        <NCard
          v-for="app in filteredApps"
          :key="app.key"
          class="app-card"
          hoverable
        >
          <div class="app-card__header">
            <div class="app-card__icon" :style="{ background: getIconColor(app.key) }">
              {{ app.name.charAt(0).toUpperCase() }}
            </div>
            <div class="app-card__meta">
              <div class="app-card__name">{{ app.name }}</div>
              <div class="app-card__version mono">v{{ app.version }}</div>
            </div>
            <div class="app-card__tags">
              <NTag
                size="small"
                bordered
                :type="categoryTagType(app.category) as any"
                class="app-card__cat"
              >
                {{ categoryLabel(t, app.category) }}
              </NTag>
              <component :is="appSourceTag(app, sourceMap)" v-if="app.source_id" />
            </div>
          </div>

          <p class="app-card__desc">{{ app.description }}</p>

          <div v-if="app.resources" class="app-card__resources">
            <NTag v-if="app.resources.memory" size="tiny" bordered type="info">
              {{ t('appstore.storeTab.memory', { value: app.resources.memory }) }}
            </NTag>
            <NTag v-if="app.resources.disk" size="tiny" bordered type="info">
              {{ t('appstore.storeTab.disk', { value: app.resources.disk }) }}
            </NTag>
          </div>

          <div class="app-card__footer">
            <NButton type="primary" size="small" block @click="$emit('install', app)">
              <template #icon><NIcon><DownloadOutline /></NIcon></template>
              {{ t('appstore.storeTab.install') }}
            </NButton>
          </div>
        </NCard>
      </div>

      <div v-else-if="!appsLoading" class="empty-hint">
        <NEmpty
          :description="
            searchDebounced || selectedCategory !== 'all' || selectedSource !== 'all'
              ? t('appstore.storeTab.noMatch')
              : t('appstore.storeTab.noApps')
          "
        />
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
.store-filter {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-5);
  flex-wrap: wrap;
  gap: var(--space-3);
}

.store-filter__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.store-filter__right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.cat-tag {
  cursor: pointer;
  transition: color var(--t-fast) var(--ease), background-color var(--t-fast) var(--ease), border-color var(--t-fast) var(--ease), box-shadow var(--t-fast) var(--ease);
}

.cat-tag:hover {
  opacity: 0.8;
}

/* 应用卡片网格 - 响应式 CSS Grid */
.app-grid {
  margin-top: var(--space-2);
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

/* 平板：2列 */
@media (max-width: 1024px) {
  .app-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
  }
}

/* 手机：1列 */
@media (max-width: 640px) {
  .app-grid {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}

.app-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: transform var(--t-fast) var(--ease), border-color var(--t-fast) var(--ease);
}

.app-card:hover {
  transform: translateY(-2px);
}

.app-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: var(--space-4);
}

.app-card__header {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.app-card__icon {
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

.app-card__meta {
  flex: 1;
  min-width: 0;
}

.app-card__name {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
  line-height: 1.3;
}

.app-card__version {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  margin-top: var(--space-0);
}

.app-card__tags {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--space-1);
  flex-shrink: 0;
}

.app-card__cat {
  flex-shrink: 0;
}

.app-card__source {
  flex-shrink: 0;
}

.app-card__desc {
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

.app-card__resources {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
  flex-wrap: wrap;
}

.app-card__footer {
  margin-top: auto;
}

/* 移动端卡片优化 */
@media (max-width: 640px) {
  .app-card :deep(.n-card__content) {
    padding: var(--space-3);
  }

  .app-card__header {
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }

  .app-card__icon {
    width: var(--size-sm);
    height: var(--size-sm);
    font-size: var(--fs-title);
  }

  .app-card__name {
    font-size: var(--fs-body);
  }

  .app-card__tags {
    display: none; /* 小屏隐藏分类标签，节省空间 */
  }

  .app-card__desc {
    font-size: var(--fs-caption);
    -webkit-line-clamp: 2;
  }
}

/* 空态 */
.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
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
</style>
