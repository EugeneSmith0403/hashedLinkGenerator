<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-indigo-50 to-white flex items-center justify-center p-4">
    <div class="absolute top-4 right-4">
      <LocaleSwitcher />
    </div>
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-12 h-12 bg-indigo-600 rounded-xl mb-4">
          <svg class="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">LinkShort</h1>
      </div>
      <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
        <slot />
      </div>
    </div>
  </div>
</template>
