export interface StatRecord {
  ID: number
  linkId: number
  clicks: number
  date: string
  Link?: { url: string; hash: string }
}

export interface StatsByDate {
  amountClicks: number
  date: string
}

export function useStatsService() {
  const { request } = useApiClient()

  return {
    getStats: (params?: { from?: string; to?: string; linkId?: number }) => {
      const q = new URLSearchParams()
      if (params?.from) q.set('from', params.from)
      if (params?.to) q.set('to', params.to)
      if (params?.linkId) q.set('linkId', String(params.linkId))
      const qs = q.toString()
      return request<StatRecord[]>(`/stats${qs ? `?${qs}` : ''}`)
    },

    getStatsByDate: (linkId: number, params?: { from?: string; to?: string }) => {
      const q = new URLSearchParams()
      if (params?.from) q.set('from', params.from)
      if (params?.to) q.set('to', params.to)
      const qs = q.toString()
      return request<StatsByDate[]>(`/stats/link/${linkId}${qs ? `?${qs}` : ''}`)
    },
  }
}
