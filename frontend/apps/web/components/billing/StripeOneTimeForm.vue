<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { StripeCardElement } from '@stripe/stripe-js'

const props = defineProps<{ planId: number | null }>()

const { t } = useI18n()
const { $stripe } = useNuxtApp()

const { data: plans } = useQuery({
  queryKey: ['plans'],
  queryFn: () => usePlansService().getAll(),
})

const selectedPlan = computed(() =>
  plans.value?.find((p) => p.ID === props.planId) ?? null,
)

const payLabel = computed(() => {
  if (!selectedPlan.value) return t('billing.pay')
  return `${t('billing.pay')} ${selectedPlan.value.cost} ${selectedPlan.value.currency.toUpperCase()}`
})
const queryClient = useQueryClient()

const cardRef = ref<HTMLDivElement>()
let cardElement: StripeCardElement | null = null

const isCardMounted = ref(false)
const cardError = ref<string | null>(null)
const isProcessing = ref(false)
const globalError = ref<string | null>(null)
const isSuccess = ref(false)

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

const { mutateAsync: createAccount } = useMutation({
  mutationFn: () => useAccountService().create(),
})

const { mutateAsync: createPaymentIntent } = useMutation({
  mutationFn: ({ cardType, planId }: { cardType: string; planId: number }) =>
    usePaymentService().createPaymentIntent(cardType, planId),
})

const { mutateAsync: confirmPaymentIntent } = useMutation({
  mutationFn: (paymentId: string) => usePaymentService().confirmPaymentIntent(paymentId),
  onSuccess: () => {
    isSuccess.value = true
    queryClient.invalidateQueries({ queryKey: ['me'] })
  },
})

async function pay() {
  if (!props.planId || !cardElement || !$stripe) return
  globalError.value = null
  isSuccess.value = false
  isProcessing.value = true

  try {
    // 1. Ensure Stripe customer exists
    await createAccount()

    // 2. Create payment method from card element
    const pmResult = await ($stripe as any).createPaymentMethod({
      type: 'card',
      card: cardElement,
    })

    if (pmResult.error) {
      globalError.value = pmResult.error.message ?? t('billing.stripeError')
      return
    }

    // 3. Create PaymentIntent on backend
    const pi = await createPaymentIntent({
      cardType: pmResult.paymentMethod.id,
      planId: props.planId,
    })

    // 4. Confirm PaymentIntent on backend
    await confirmPaymentIntent(pi.metadata.payment_id)
  } catch (e: any) {
    globalError.value = e.message ?? t('billing.stripeError')
  } finally {
    isProcessing.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="isSuccess" class="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
      {{ t('billing.paymentSuccess') }}
    </div>

    <template v-else>
      <div>
        <p class="text-sm font-medium text-gray-700 mb-2">{{ t('billing.cardDetails') }}</p>
        <div
          ref="cardRef"
          class="border border-gray-300 rounded-lg px-3 py-3 focus-within:ring-2 focus-within:ring-indigo-500 focus-within:border-indigo-500 transition-all"
        />
        <p v-if="cardError" class="mt-1 text-xs text-red-600">{{ cardError }}</p>
      </div>

      <p v-if="globalError" class="text-sm text-red-600">{{ globalError }}</p>

      <UiButton
        class="w-full"
        :disabled="!planId || !isCardMounted || isProcessing"
        :loading="isProcessing"
        @click="pay"
      >
        {{ payLabel }}
      </UiButton>

      <p class="text-xs text-gray-400 text-center">
        {{ t('billing.stripeSecure') }}
      </p>
    </template>
  </div>
</template>
