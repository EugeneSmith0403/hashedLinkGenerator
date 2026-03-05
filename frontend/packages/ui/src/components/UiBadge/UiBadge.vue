<script setup lang="ts">
defineProps<{
  variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral'
}>()

const classes: Record<string, string> = {
  success: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  danger: 'bg-red-100 text-red-800',
  info: 'bg-blue-100 text-blue-800',
  neutral: 'bg-gray-100 text-gray-700',
}
</script>

<template>
  <span
    :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', classes[variant ?? 'neutral']]"
  >
    <slot />
  </span>
</template>
