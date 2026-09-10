<script setup lang="ts">
withDefaults(defineProps<{
  size?: number
}>(), {
  size: 32
})
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 64 64"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    class="rhino-logo"
  >
    <!-- 背景 -->
    <rect x="3" y="3" width="58" height="58" rx="16" fill="var(--surface2)" />

    <!-- 前角（大角，最标志性） -->
    <path
      d="M32 8 C30 14 28 20 28 26 L36 26 C36 20 34 14 32 8 Z"
      stroke="var(--acc)"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      fill="none"
    />

    <!-- 后角（小角） -->
    <path
      d="M32 22 C31 26 30 29 30 32 L34 32 C34 29 33 26 32 22 Z"
      stroke="var(--acc)"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      fill="none"
      opacity="0.7"
    />

    <!-- 头部轮廓（正面，宽口鼻） -->
    <path
      d="
        M 20 30
        C 16 32 14 36 14 42
        C 14 50 20 54 32 54
        C 44 54 50 50 50 42
        C 50 36 48 32 44 30
        L 38 30
        L 38 28
        L 26 28
        L 26 30
        Z
      "
      stroke="var(--acc)"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      fill="none"
    />

    <!-- 耳朵（左右） -->
    <path d="M20 30 L16 24 L24 28" stroke="var(--acc)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none" />
    <path d="M44 30 L48 24 L40 28" stroke="var(--acc)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none" />

    <!-- 眼睛（左右） -->
    <circle cx="24" cy="38" r="1.8" fill="var(--acc2)" />
    <circle cx="40" cy="38" r="1.8" fill="var(--acc2)" />

    <!-- 鼻孔 -->
    <ellipse cx="28" cy="46" rx="2" ry="1.2" fill="var(--acc)" opacity="0.35" />
    <ellipse cx="36" cy="46" rx="2" ry="1.2" fill="var(--acc)" opacity="0.35" />
  </svg>
</template>

<style scoped>
.rhino-logo {
  display: block;
  flex-shrink: 0;
}
</style>
