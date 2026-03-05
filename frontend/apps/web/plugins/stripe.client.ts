import { loadStripe } from '@stripe/stripe-js'

export default defineNuxtPlugin(async () => {
  const config = useRuntimeConfig()
  const stripe = await loadStripe(config.public.stripeKey)

  return {
    provide: {
      stripe,
    },
  }
})
