import { defineStore } from 'pinia'

const STORAGE_KEY = 'auth'

interface AuthState {
  token: string | null
  email: string | null
  initialized: boolean
}

function loadFromStorage(): AuthState {
  if (import.meta.server) return { token: null, email: null }
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw) as AuthState
  } catch {}
  return { token: null, email: null }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({ ...loadFromStorage(), initialized: false }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setAuth(token: string, email: string) {
      this.token = token
      this.email = email
      if (import.meta.client) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, email }))
      }
    },
    logout() {
      this.token = null
      this.email = null
      if (import.meta.client) {
        localStorage.removeItem(STORAGE_KEY)
      }
    },
  },
})
