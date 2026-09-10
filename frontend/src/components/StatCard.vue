<script setup lang="ts">
import { computed } from 'vue'
import { NIcon, NSkeleton } from 'naive-ui'

const props = withDefaults(
  defineProps<{
    title: string
    value: string | number
    unit?: string
    subtitle?: string
    trend?: number
    trendLabel?: string
    loading?: boolean
    accent?: 'acc' | 'acc2' | 'ok' | 'warn' | 'err' | 'info'
  }>(),
  {
    loading: false,
    accent: 'acc2'
  }
)

const accentColor = computed(() => {
  const map: Record<string, string> = {
    acc: 'var(--acc)',
    acc2: 'var(--acc2)',
    ok: 'var(--ok)',
    warn: 'var(--warn)',
    err: 'var(--err)',
    info: 'var(--info)'
  }
  return map[props.accent] || map.acc2
})

const trendPositive = computed(() => (props.trend ?? 0) >= 0)
</script>

<template>
  <div class="stat-card">
    <div class="stat-card__top">
      <span class="stat-card__title">{{ title }}</span>
      <span class="led" :class="loading ? 'led--offline' : 'led--ok'" />
    </div>
    <div class="stat-card__value-wrap">
      <template v-if="!loading">
        <span class="stat-card__value mono" :style="{ color: accentColor }">{{ value }}</span>
        <span v-if="unit" class="stat-card__unit">{{ unit }}</span>
      </template>
      <NSkeleton v-else text :repeat="1" style="width: 60%" />
    </div>
    <div class="stat-card__bottom">
      <span v-if="subtitle" class="stat-card__subtitle">{{ subtitle }}</span>
      <span v-if="trend !== undefined" class="stat-card__trend" :class="{ 'is-up': trendPositive, 'is-down': !trendPositive }">
        {{ trendPositive ? '↑' : '↓' }} {{ Math.abs(trend) }}%
        <span v-if="trendLabel" class="stat-card__trend-label">{{ trendLabel }}</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.stat-card {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  padding: var(--space-4) var(--space-5);
  transition: border-color var(--t-fast) var(--ease), transform var(--t-fast) var(--ease);
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: var(--stat-bar-height);
  background: var(--acc2);
  opacity: 0.6;
}

.stat-card:hover {
  border-color: var(--line-strong);
  transform: translateY(var(--hover-lift));
}

.stat-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.stat-card__title {
  font-size: var(--fs-caption);
  color: var(--sub);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 500;
}

.stat-card__value-wrap {
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
  min-height: var(--stat-value-min-height);
}

.stat-card__value {
  font-size: var(--fs-hero);
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.02em;
}

.stat-card__unit {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.stat-card__bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--space-3);
}

.stat-card__subtitle {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.stat-card__trend {
  font-size: var(--fs-caption);
  font-family: var(--font-mono);
  font-weight: 600;
}

.stat-card__trend.is-up {
  color: var(--ok);
}

.stat-card__trend.is-down {
  color: var(--err);
}

.stat-card__trend-label {
  color: var(--faint);
  margin-left: var(--space-1);
  font-weight: 400;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .stat-card {
    padding: var(--space-3) var(--space-4);
  }
  .stat-card__bottom {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-1);
  }
}
</style>
