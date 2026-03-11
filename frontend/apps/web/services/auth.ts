import type { LoginInput, RegisterInput } from '~/schemas/auth'

export interface AuthResponse {
  email: string
  token: string
  is2faEnabled: boolean
}

export function useAuthService() {
  const { request } = useApiClient()

  return {
    login: (body: LoginInput) =>
      request<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify(body) }),

    register: (body: RegisterInput) =>
      request<AuthResponse>('/auth/register', { method: 'POST', body: JSON.stringify(body) }),
  }
}
