<script setup lang="ts">
defineProps<{ text: string }>()

const anchor = ref<HTMLElement | null>(null)
const visible = ref(false)
const pos = ref({ top: 0, left: 0 })

function show() {
  if (!anchor.value) return
  const rect = anchor.value.getBoundingClientRect()
  pos.value = {
    top: rect.top + window.scrollY - 8,
    left: rect.left + rect.width / 2,
  }
  visible.value = true
}

function hide() {
  visible.value = false
}
</script>

<template>
  <span ref="anchor" class="inline-flex" @mouseenter="show" @mouseleave="hide">
    <slot />
  </span>

  <Teleport to="body">
    <div
      v-if="visible"
      class="pointer-events-none fixed z-[9999] -translate-x-1/2 -translate-y-full
             max-w-xs px-2.5 py-1.5 rounded-md bg-gray-900 text-white text-xs leading-snug break-all"
      :style="{ top: `${pos.top}px`, left: `${pos.left}px` }"
    >
      {{ text }}
      <div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900" />
    </div>
  </Teleport>
</template>
