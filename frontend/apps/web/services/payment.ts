export interface Payment {
  id: string
  createdAt: string
  updatedAt: string
  accountId: number
  invoiceId: number | null
  planId: number | null
  subscriptionId: number | null
  paymentIntentId: string
  chargeId: string | null
  amount: number
  platformFee: number
  netAmount: number
  currency: string
  status: string
  paymentMethodType: string
  failureCode: string
  failureMessage: string
}

export interface PaymentIntentResponse {
  id: string
  client_secret: string
  status: string
  amount: number
  currency: string
  metadata: {
    payment_id: string
    plan_id: string
    user_id: string
  }
}

export interface ConfirmPaymentResponse {
  confirmed: boolean
  paymentId: string
  confirmedUrl: string
}

export function usePaymentService() {
  const { request } = useApiClient()

  return {
    createPaymentIntent: (cardType: string, planId: number) =>
      request<PaymentIntentResponse>('/stripe/paymentIntent', {
        method: 'POST',
        body: JSON.stringify({ cardType, planId }),
      }),

    confirmPaymentIntent: (paymentId: string) =>
      request<ConfirmPaymentResponse>('/stripe/paymentIntent/confirm', {
        method: 'POST',
        body: JSON.stringify({ paymentId }),
      }),

    getPayments: () => request<Payment[]>('/payments'),

    cancelPaymentIntent: () =>
      request<void>('/stripe/paymentIntent/cancel', { method: 'POST' }),
  }
}
