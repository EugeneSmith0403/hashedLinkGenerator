export interface Subscription {
  id: number
  createdAt: string
  userId: number
  planId: number
  status: string
  currentPeriodStart: string
  currentPeriodEnd: string
  cancelAt: string | null
  canceledAt: string | null
  trialStart: string | null
  trialEnd: string | null
  isPaymentIntent: boolean
}

export interface SetupIntentResponse {
  clientSecret: string
}

export function useSubscriptionService() {
  const { request } = useApiClient()

  return {
    getCurrent: () => request<Subscription | null>('/subscriptions/me'),

    createSetupIntent: () =>
      request<SetupIntentResponse>('/subscriptions/method', { method: 'POST' }),

    create: (planId: number) =>
      request<Subscription>('/subscriptions', { method: 'POST', body: JSON.stringify({ planId }) }),

    cancel: () =>
      request<Subscription>('/subscriptions/cancel', { method: 'PATCH' }),
  }
}
