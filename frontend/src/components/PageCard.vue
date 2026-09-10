<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
  padding?: string
}>()
</script>

<template>
  <div class="page-card">
    <div v-if="title || subtitle || $slots.extra" class="page-card__header">
      <div class="page-card__header-left">
        <h2 v-if="title" class="page-card__title">{{ title }}</h2>
        <p v-if="subtitle" class="page-card__subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.extra" class="page-card__header-right">
        <slot name="extra" />
      </div>
    </div>
    <div class="page-card__body" :style="{ padding: padding || 'var(--space-5)' }">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.page-card {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  transition: border-color var(--t-fast) var(--ease);
}

.page-card__header {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.page-card__header-left {
  flex: 1;
  min-width: 0;
}

.page-card__header-right {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.page-card__title {
  font-size: var(--fs-title);
  font-weight: var(--fw-title);
  color: var(--ink);
  line-height: var(--lh-title);
}

.page-card__subtitle {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: var(--space-1);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .page-card__header {
    padding: var(--space-3) var(--space-4);
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .page-card__header-right {
    width: 100%;
    justify-content: flex-end;
  }
  .page-card__body {
    padding: var(--space-4) !important;
  }
}
</style>
