<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'

const { t } = useI18n()

const { data: subscription, isLoading } = useQuery({
  queryKey: ['subscription'],
  queryFn: () => useSubscriptionService().getCurrent(),
})

const hasActiveSub = computed(() =>
  subscription.value?.status === 'active' || subscription.value?.status === 'trialing',
)

const selectedPlanId = ref<number | null>(null)
const paymentMode = ref<'subscription' | 'onetime'>('subscription')
</script>

<template>
  <div class="max-w-2xl space-y-8">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">{{ t('billing.title') }}</h1>
      <p class="text-sm text-gray-500 mt-1">{{ t('billing.subtitle') }}</p>
    </div>

    <div v-if="isLoading" class="flex justify-center py-8">
      <UiSpinner size="lg" />
    </div>

    <template v-else>
      <!-- Active subscription -->
      <template v-if="hasActiveSub && subscription">
        <SubscriptionCard :subscription="subscription" />
      </template>

      <!-- No active subscription → show subscribe form -->
      <template v-else>
        <UiCard>
          <div class="space-y-6">
            <div>
              <h2 class="text-lg font-semibold text-gray-900">{{ t('billing.noActiveSub') }}</h2>
              <p class="text-sm text-gray-500 mt-1">{{ t('billing.chooseToUnlock') }}</p>
            </div>

            <!-- Payment mode toggle -->
            <div class="flex rounded-lg border border-gray-200 p-1 gap-1 bg-gray-50">
              <button
                :class="[
                  'flex-1 py-2 px-4 rounded-md text-sm font-medium transition-all',
                  paymentMode === 'subscription'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700',
                ]"
                @click="paymentMode = 'subscription'"
              >
                {{ t('billing.modeSubscription') }}
              </button>
              <button
                :class="[
                  'flex-1 py-2 px-4 rounded-md text-sm font-medium transition-all',
                  paymentMode === 'onetime'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700',
                ]"
                @click="paymentMode = 'onetime'"
              >
                {{ t('billing.modeOneTime') }}
              </button>
            </div>

            <PlanSelector v-model="selectedPlanId" />

            <div class="border-t border-gray-100 pt-6">
              <StripeSetupForm v-if="paymentMode === 'subscription'" :plan-id="selectedPlanId" />
              <StripeOneTimeForm v-else :plan-id="selectedPlanId" />
            </div>
          </div>
        </UiCard>
      </template>

      <!-- Canceled subscription notice -->
      <template v-if="subscription && subscription.status === 'canceled'">
        <div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
          {{ t('billing.subCanceled') }}
        </div>
      </template>
    </template>
  </div>
</template>
