<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'

const { t, locale } = useI18n()

const { data: payments, isLoading } = useQuery({
  queryKey: ['payments'],
  queryFn: () => usePaymentService().getPayments(),
})

function formatAmount(amount: number, currency: string) {
  return new Intl.NumberFormat(locale.value, {
    style: 'currency',
    currency: currency.toUpperCase(),
    minimumFractionDigits: 2,
  }).format(amount / 100)
}

function formatDate(dateStr: string) {
  return new Intl.DateTimeFormat(locale.value, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(dateStr))
}

const statusKeyMap: Record<string, string> = {
  succeeded: 'statusSucceeded',
  failed: 'statusFailed',
  pending: 'statusPending',
  canceled: 'statusCanceled',
  requires_action: 'statusRequiresAction',
  requires_capture: 'statusRequiresCapture',
  requires_confirmation: 'statusRequiresConfirmation',
  requires_payment_method: 'statusRequiresPaymentMethod',
  processing: 'statusProcessing',
}

function statusLabel(status: string) {
  const key = statusKeyMap[status]
  return key ? t(`payments.${key}`) : status
}

function statusColor(status: string) {
  if (status === 'succeeded') return 'text-green-600 bg-green-50'
  if (status === 'failed' || status === 'canceled') return 'text-red-600 bg-red-50'
  return 'text-amber-600 bg-amber-50'
}
</script>

<template>
  <div class="max-w-4xl space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">{{ t('payments.title') }}</h1>
      <p class="text-sm text-gray-500 mt-1">{{ t('payments.subtitle') }}</p>
    </div>

    <div v-if="isLoading" class="flex justify-center py-12">
      <UiSpinner size="lg" />
    </div>

    <template v-else>
      <p v-if="!payments || payments.length === 0" class="text-sm text-gray-400 py-8 text-center">
        {{ t('payments.empty') }}
      </p>

      <div v-else class="overflow-hidden rounded-xl border border-gray-200">
        <table class="min-w-full divide-y divide-gray-200 text-sm">
          <thead class="bg-gray-50 text-xs text-gray-500 uppercase tracking-wide">
            <tr>
              <th class="px-4 py-3 text-left">{{ t('payments.date') }}</th>
              <th class="px-4 py-3 text-left">{{ t('payments.amount') }}</th>
              <th class="px-4 py-3 text-left">{{ t('payments.method') }}</th>
              <th class="px-4 py-3 text-left">{{ t('payments.status') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 bg-white">
            <tr v-for="p in payments" :key="p.id">
              <td class="px-4 py-3 text-gray-600 whitespace-nowrap">{{ formatDate(p.createdAt) }}</td>
              <td class="px-4 py-3 font-medium text-gray-900 whitespace-nowrap">
                {{ formatAmount(p.amount, p.currency) }}
              </td>
              <td class="px-4 py-3 text-gray-500 capitalize">{{ p.paymentMethodType || '—' }}</td>
              <td class="px-4 py-3">
                <span
                  :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', statusColor(p.status)]"
                >
                  {{ statusLabel(p.status) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>
