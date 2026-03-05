export interface AccountResponse {
  id: number
  accountStatus: string
  provider: string
}

export function useAccountService() {
  const { request } = useApiClient()

  return {
    create: () => request<AccountResponse>('/account', { method: 'POST' }),
  }
}
