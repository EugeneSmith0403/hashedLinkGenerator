<script setup lang="ts">
const auth = useAuthStore()
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <Transition name="fade">
    <div v-if="!auth.initialized" class="fixed inset-0 z-[9999] flex items-center justify-center bg-white">
      <div class="h-10 w-10 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
    </div>
  </Transition>
</template>

<style scoped>
.fade-leave-active { transition: opacity 0.15s ease; }
.fade-leave-to { opacity: 0; }
</style>
