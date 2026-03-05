export function useApiClient() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (auth.token) {
      headers['Authorization'] = `Bearer ${auth.token}`
    }

    const res = await fetch(`${config.public.apiBase}${path}`, {
      ...options,
      headers,
    })

    if (!res.ok) {
      if (res.status === 401) {
        auth.logout()
        await navigateTo('/auth/login')
      }
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err?.error ?? res.statusText)
    }

    if (res.status === 204) return null as T
    return res.json() as Promise<T>
  }

  return { request }
}
