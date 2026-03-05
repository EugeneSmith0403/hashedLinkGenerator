<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import type { Link } from '~/services/links'

const props = defineProps<{ link: Link | null; open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()

const { data: stats, isLoading, refetch } = useQuery({
  queryKey: computed(() => ['stats', props.link?.ID]),
  queryFn: () => useStatsService().getStats({ linkId: props.link!.ID }),
  enabled: computed(() => !!props.link && props.open),
})

watch(() => props.open, (val) => {
  if (val) refetch()
})

const totalClicks = computed(() =>
  stats.value?.reduce((sum, s) => sum + s.clicks, 0) ?? 0,
)

const config = useRuntimeConfig()
function shortUrl(hash: string) {
  return `${config.public.apiBase}/${hash}`
}
</script>

<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="translate-x-full opacity-0"
    enter-to-class="translate-x-0 opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="translate-x-0 opacity-100"
    leave-to-class="translate-x-full opacity-0"
  >
    <div v-if="open && link" class="fixed inset-y-0 right-0 w-[420px] bg-white shadow-2xl border-l border-gray-200 flex flex-col z-40">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-900">{{ t('dashboard.stats') }}</h2>
        <button class="text-gray-400 hover:text-gray-600" @click="emit('close')">
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- Link info -->
        <div class="space-y-1">
          <p class="text-xs text-gray-400 uppercase tracking-wider">{{ t('dashboard.originalUrl') }}</p>
          <a :href="link.url" target="_blank" class="text-sm text-indigo-600 hover:underline break-all">{{ link.url }}</a>
        </div>
        <div class="space-y-1">
          <p class="text-xs text-gray-400 uppercase tracking-wider">{{ t('dashboard.shortUrl') }}</p>
          <p class="text-sm font-mono text-gray-700">{{ shortUrl(link.hash) }}</p>
        </div>

        <!-- Total clicks -->
        <UiCard>
          <div class="text-center">
            <p class="text-3xl font-bold text-gray-900">{{ totalClicks }}</p>
            <p class="text-sm text-gray-500 mt-1">{{ t('dashboard.totalClicks') }}</p>
          </div>
        </UiCard>

        <!-- Stats table -->
        <div>
          <p class="text-sm font-medium text-gray-700 mb-3">{{ t('dashboard.clicksByDate') }}</p>
          <div v-if="isLoading" class="flex justify-center py-6">
            <UiSpinner />
          </div>
          <div v-else-if="!stats?.length" class="text-sm text-gray-400 text-center py-6">
            {{ t('dashboard.noStats') }}
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="stat in stats"
              :key="stat.ID"
              class="flex items-center justify-between px-3 py-2 rounded-lg bg-gray-50"
            >
              <span class="text-sm text-gray-600">{{ new Date(stat.date).toLocaleDateString() }}</span>
              <UiBadge variant="info">{{ stat.clicks }} {{ t('dashboard.clicks') }}</UiBadge>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>

  <!-- backdrop -->
  <Transition enter-active-class="transition-opacity duration-200" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition-opacity duration-200" leave-from-class="opacity-100" leave-to-class="opacity-0">
    <div v-if="open && link" class="fixed inset-0 bg-black/20 z-30" @click="emit('close')" />
  </Transition>
</template>
