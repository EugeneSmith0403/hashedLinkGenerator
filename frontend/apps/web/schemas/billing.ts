import { z } from 'zod'

export const subscribeSchema = z.object({
  planId: z.number().positive(),
})

export type SubscribeInput = z.infer<typeof subscribeSchema>
