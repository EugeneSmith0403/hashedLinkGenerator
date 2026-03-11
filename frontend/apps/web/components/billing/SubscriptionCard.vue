<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import type { Subscription } from '~/services/subscription'

defineProps<{ subscription: Subscription }>()

const { t } = useI18n()
const queryClient = useQueryClient()

const statusVariant: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
  active: 'success',
  trialing: 'info',
  past_due: 'warning',
  canceled: 'danger',
  unpaid: 'danger',
  incomplete: 'warning',
  paused: 'neutral',
}

const invalidate = () => {
  queryClient.invalidateQueries({ queryKey: ['me'] })
  queryClient.invalidateQueries({ queryKey: ['payments'] })
}

const { mutate: cancelSub, isPending: isCancelSubPending } = useMutation({
  mutationFn: () => useSubscriptionService().cancel(),
  onSuccess: invalidate,
})

const { mutate: cancelPI, isPending: isCancelPIPending } = useMutation({
  mutationFn: () => usePaymentService().cancelPaymentIntent(),
  onSuccess: invalidate,
})
</script>

<template>
  <UiCard>
    <div class="flex items-start justify-between">
      <div class="space-y-1">
        <h3 class="text-lg font-semibold text-gray-900">{{ t('billing.currentPlan') }}</h3>
        <UiBadge :variant="statusVariant[subscription.status] ?? 'neutral'">
          {{ subscription.status }}
        </UiBadge>
      </div>

      <template v-if="['active', 'trialing'].includes(subscription.status)">
        <UiButton
          v-if="subscription.isPaymentIntent"
          variant="danger"
          size="sm"
          :loading="isCancelPIPending"
          @click="cancelPI()"
        >
          {{ t('billing.cancelPayment') }}
        </UiButton>
        <UiButton
          v-else
          variant="danger"
          size="sm"
          :loading="isCancelSubPending"
          @click="cancelSub()"
        >
          {{ t('billing.cancel') }}
        </UiButton>
      </template>
    </div>

    <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
      <div>
        <p class="text-gray-400">{{ t('billing.periodStart') }}</p>
        <p class="font-medium text-gray-900">{{ new Date(subscription.currentPeriodStart).toLocaleDateString() }}</p>
      </div>
      <div>
        <p class="text-gray-400">{{ t('billing.periodEnd') }}</p>
        <p class="font-medium text-gray-900">{{ new Date(subscription.currentPeriodEnd).toLocaleDateString() }}</p>
      </div>
      <div v-if="subscription.canceledAt">
        <p class="text-gray-400">{{ t('billing.canceledAt') }}</p>
        <p class="font-medium text-red-600">{{ new Date(subscription.canceledAt).toLocaleDateString() }}</p>
      </div>
    </div>
  </UiCard>
</template>
