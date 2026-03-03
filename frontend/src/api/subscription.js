import { api } from './client.js'

export const subscriptionApi = {
  addPaymentMethod: () => api.post('/subscriptions/method', {}),
  createSubscription: (planId) => api.post('/subscriptions', { planId }),
}
