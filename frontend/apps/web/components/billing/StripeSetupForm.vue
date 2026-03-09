<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import type { StripeCardElement } from '@stripe/stripe-js'

const props = defineProps<{ planId: number | null }>()

const { t } = useI18n()
const { $stripe } = useNuxtApp()
const queryClient = useQueryClient()

const cardRef = ref<HTMLDivElement>()
let cardElement: StripeCardElement | null = null

const isCardMounted = ref(false)
const cardError = ref<string | null>(null)
const isSavingCard = ref(false)
const isSubscribing = ref(false)
const paymentMethodReady = ref(false)
const globalError = ref<string | null>(null)

onMounted(async () => {
  if (!$stripe || !cardRef.value) return

  const elements = ($stripe as any).elements()
  cardElement = elements.create('card', {
    style: {
      base: {
        fontSize: '14px',
        color: '#111827',
        fontFamily: 'Inter, system-ui, sans-serif',
        '::placeholder': { color: '#9CA3AF' },
      },
    },
  })
  cardElement.mount(cardRef.value)
  cardElement.on('change', (e: any) => {
    cardError.value = e.error?.message ?? null
  })
  isCardMounted.value = true
})

onBeforeUnmount(() => {
  cardElement?.destroy()
})

const { mutateAsync: createSetupIntent } = useMutation({
  mutationFn: () => useSubscriptionService().createSetupIntent(),
})

const { mutateAsync: createAccount } = useMutation({
  mutationFn: () => useAccountService().create(),
})

const { mutateAsync: createSubscription } = useMutation({
  mutationFn: (planId: number) => useSubscriptionService().create(planId),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['subscription'] }),
})

async function saveCard() {
  if (!cardElement || !$stripe) return
  globalError.value = null
  isSavingCard.value = true

  try {
    await createAccount()

    const { clientSecret } = await createSetupIntent()

    const result = await ($stripe as any).confirmCardSetup(clientSecret, {
      payment_method: { card: cardElement },
    })

    if (result.error) {
      globalError.value = result.error.message ?? t('billing.stripeError')
      return
    }

    paymentMethodReady.value = true
  } catch (e: any) {
    globalError.value = e.message ?? t('billing.stripeError')
  } finally {
    isSavingCard.value = false
  }
}

async function subscribe() {
  if (!props.planId || !paymentMethodReady.value) return
  globalError.value = null
  isSubscribing.value = true

  try {
    await createSubscription(props.planId)
  } catch (e: any) {
    globalError.value = e.message ?? t('billing.stripeError')
  } finally {
    isSubscribing.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Step 1: Card details -->
    <div>
      <p class="text-sm font-medium text-gray-700 mb-2">{{ t('billing.cardDetails') }}</p>
      <div
        ref="cardRef"
        class="border border-gray-300 rounded-lg px-3 py-3 focus-within:ring-2 focus-within:ring-indigo-500 focus-within:border-indigo-500 transition-all"
        :class="{ 'opacity-50 pointer-events-none': paymentMethodReady }"
      />
      <p v-if="cardError" class="mt-1 text-xs text-red-600">{{ cardError }}</p>
    </div>

    <UiButton
      v-if="!paymentMethodReady"
      class="w-full"
      variant="secondary"
      :disabled="!isCardMounted || isSavingCard"
      :loading="isSavingCard"
      @click="saveCard"
    >
      {{ t('billing.saveCard') }}
    </UiButton>

    <div
      v-else
      class="flex items-center gap-2 text-sm text-green-600 font-medium"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
      </svg>
      {{ t('billing.cardSaved') }}
    </div>

    <p v-if="globalError" class="text-sm text-red-600">{{ globalError }}</p>

    <!-- Step 2: Subscribe -->
    <UiButton
      class="w-full"
      :disabled="!planId || !paymentMethodReady || isSubscribing"
      :loading="isSubscribing"
      @click="subscribe"
    >
      {{ t('billing.subscribe') }}
    </UiButton>

    <p class="text-xs text-gray-400 text-center">
      {{ t('billing.stripeSecure') }}
    </p>
  </div>
</template>
