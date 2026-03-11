export interface AccountResponse {
  id: number
  accountStatus: string
  provider: string
}

export interface TwoFactorSetupResponse {
  secret: string
  qrUrl: string
}

export function useAccountService() {
  const { request } = useApiClient()

  return {
    create: () => request<AccountResponse>('/account', { method: 'POST' }),

    setup2FA: () =>
      request<TwoFactorSetupResponse>('/account/2fa/setup', { method: 'POST' }),

    verify2FA: (code: string) =>
      request<void>('/account/2fa/verify', { method: 'POST', body: JSON.stringify({ code }) }),

    disable2FA: (code: string) =>
      request<void>('/account/2fa/disable', { method: 'POST', body: JSON.stringify({ code }) }),
  }
}
