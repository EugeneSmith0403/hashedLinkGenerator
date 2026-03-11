import type { Subscription } from './subscription'

export interface UserMe {
  id: number
  name: string
  email: string
  is2faEnabled: boolean
  subscription: Subscription | null
}

export interface Setup2FAResponse {
  qrCode: string
}

export interface Verify2FAResponse {
  token: string
}

export function useUserService() {
  const { request } = useApiClient()

  return {
    getMe: () => request<UserMe>('/users/me'),
    setup2FA: () => request<Setup2FAResponse>('/users/me/2fa/setup', { method: 'POST' }),
    verify2FA: (code: string, email: string) =>
      request<Verify2FAResponse>('/users/2fa/verify', { method: 'POST', body: JSON.stringify({ code, email }) }),
  }
}
