<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import type { Link } from '~/services/links'

const { t } = useI18n()

const { data: links, isLoading } = useQuery({
  queryKey: ['links'],
  queryFn: () => useLinksService().getAll(),
})

const { data: subscription } = useQuery({
  queryKey: ['subscription'],
  queryFn: () => useSubscriptionService().getCurrent(),
})

const hasActiveSub = computed(() =>
  subscription.value?.status === 'active' || subscription.value?.status === 'trialing',
)

const showCreateModal = ref(false)
const selectedLink = ref<Link | null>(null)
const showStats = ref(false)

function onRowClick(link: Link) {
  selectedLink.value = link
  showStats.value = true
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('dashboard.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('dashboard.subtitle') }}</p>
      </div>

      <div class="flex items-center gap-3">
        <span v-if="!hasActiveSub" class="text-xs text-amber-600 bg-amber-50 border border-amber-200 px-3 py-1.5 rounded-lg">
          {{ t('dashboard.noSubWarning') }}
        </span>
        <UiButton
          :disabled="!hasActiveSub"
          :title="!hasActiveSub ? t('dashboard.noSubWarning') : undefined"
          @click="showCreateModal = true"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('dashboard.createLink') }}
        </UiButton>
      </div>
    </div>

    <LinksTable
      :links="links ?? []"
      :loading="isLoading"
      @row-click="onRowClick"
    />

    <CreateLinkModal :open="showCreateModal" @close="showCreateModal = false" />

    <StatsDrawer
      :link="selectedLink"
      :open="showStats"
      @close="showStats = false"
    />
  </div>
</template>
