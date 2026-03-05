<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()

const open = ref(false)

const available = computed(() =>
  locales.value.filter((l) => l.code !== locale.value),
)

function select(code: string) {
  setLocale(code as 'en' | 'ru' | 'de')
  open.value = false
}
</script>

<template>
  <div class="relative">
    <button
      class="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-800 transition-colors px-2 py-1 rounded-lg hover:bg-gray-100"
      @click="open = !open"
    >
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
      </svg>
      {{ locale.toUpperCase() }}
    </button>

    <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />

    <div
      v-if="open"
      class="absolute right-0 top-full mt-1 bg-white rounded-lg shadow-lg border border-gray-200 py-1 min-w-[120px] z-50"
    >
      <button
        v-for="l in available"
        :key="l.code"
        class="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
        @click="select(l.code)"
      >
        {{ l.name }}
      </button>
    </div>
  </div>
</template>
