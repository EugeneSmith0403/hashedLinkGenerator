<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import type { Plan } from '~/services/plans'

const modelValue = defineModel<number | null>({ default: null })

const { t } = useI18n()

const { data: plans, isLoading } = useQuery({
  queryKey: ['plans'],
  queryFn: () => usePlansService().getAll(),
})
</script>

<template>
  <div class="space-y-3">
    <p class="text-sm font-medium text-gray-700">{{ t('billing.choosePlan') }}</p>
    <div v-if="isLoading" class="flex justify-center py-4">
      <UiSpinner />
    </div>
    <div v-else class="grid gap-3">
      <label
        v-for="plan in plans"
        :key="plan.ID"
        :class="[
          'flex items-start gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all',
          modelValue === plan.ID
            ? 'border-indigo-500 bg-indigo-50'
            : 'border-gray-200 hover:border-gray-300',
        ]"
      >
        <input
          type="radio"
          :value="plan.ID"
          :checked="modelValue === plan.ID"
          class="mt-1 accent-indigo-600"
          @change="modelValue = plan.ID"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-gray-900">{{ plan.name }}</span>
            <span class="font-bold text-gray-900">
              {{ plan.cost }} {{ plan.currency.toUpperCase() }}
            </span>
          </div>
          <ul v-if="plan.features?.length" class="mt-1 space-y-0.5">
            <li v-for="feat in plan.features" :key="feat" class="text-sm text-gray-500 flex items-center gap-1.5">
              <svg class="h-3.5 w-3.5 text-green-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
              {{ feat }}
            </li>
          </ul>
        </div>
      </label>
    </div>
  </div>
</template>
