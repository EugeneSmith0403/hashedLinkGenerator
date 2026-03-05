export interface Plan {
  ID: number
  name: string
  cost: number
  currency: string
  features: string[] | null
  isActive: boolean
  stripePriceId: string
}

export function usePlansService() {
  const { request } = useApiClient()

  return {
    getAll: () => request<Plan[]>('/plans'),
  }
}
