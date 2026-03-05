<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import type { Link } from '~/services/links'

defineProps<{
  links: Link[]
  loading: boolean
}>()

const emit = defineEmits<{ rowClick: [link: Link] }>()

const { t } = useI18n()
const queryClient = useQueryClient()

const config = useRuntimeConfig()
function shortUrl(hash: string) {
  return `${config.public.apiBase}/${hash}`
}

async function copyToClipboard(text: string) {
  await navigator.clipboard.writeText(text)
}

const { mutate: deleteLink } = useMutation({
  mutationFn: (id: number) => useLinksService().delete(id),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['links'] }),
})

const columns = computed(() => [
  { key: 'url', label: t('dashboard.originalUrl') },
  { key: 'hash', label: t('dashboard.shortUrl') },
  { key: 'CreatedAt', label: t('common.createdAt') },
  { key: 'actions', label: '', width: '72px' },
])
</script>

<template>
  <UiTable :columns="columns" :rows="links" :loading="loading">
    <template #empty>{{ t('dashboard.noLinks') }}</template>

    <template #cell-url="{ value }">
      <UiTooltip :text="String(value)">
        <span class="truncate max-w-xs block text-gray-700">{{ value }}</span>
      </UiTooltip>
    </template>

    <template #cell-hash="{ value, row }">
      <div class="flex items-center gap-2" @click.stop>
        <span class="font-mono text-sm text-gray-700">{{ value }}</span>
        <button
          class="text-gray-400 hover:text-indigo-600 transition-colors"
          :title="t('common.copy')"
          @click="copyToClipboard(shortUrl(String(value)))"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>
      </div>
    </template>

    <template #cell-CreatedAt="{ value }">
      <span class="text-gray-500 text-sm">{{ new Date(String(value)).toLocaleDateString() }}</span>
    </template>

    <template #cell-actions="{ row }">
      <div class="flex items-center gap-1">
        <button
          class="text-gray-400 hover:text-indigo-600 transition-colors p-1 rounded"
          :title="t('dashboard.stats')"
          @click="emit('rowClick', row as Link)"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
        </button>
        <button
          class="text-gray-400 hover:text-red-600 transition-colors p-1 rounded"
          :title="t('common.delete')"
          @click="deleteLink((row as Link).ID)"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </template>
  </UiTable>
</template>
