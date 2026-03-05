<script setup lang="ts">
defineProps<{
  padding?: 'none' | 'sm' | 'md' | 'lg'
}>()

const paddings: Record<string, string> = {
  none: '',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
}
</script>

<template>
  <div :class="['bg-white rounded-xl border border-gray-200 shadow-sm', paddings[padding ?? 'md']]">
    <slot />
  </div>
</template>
