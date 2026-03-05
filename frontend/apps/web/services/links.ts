export interface Link {
  ID: number
  CreatedAt: string
  userId: number
  url: string
  hash: string
}

export function useLinksService() {
  const { request } = useApiClient()

  return {
    getAll: () => request<Link[]>('/links'),

    create: (url: string) =>
      request<Link>('/link', { method: 'POST', body: JSON.stringify({ url }) }),

    delete: (id: number) =>
      request<void>(`/link/${id}`, { method: 'DELETE' }),
  }
}
