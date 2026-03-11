import { defineStore } from 'pinia'

const STORAGE_KEY = 'auth'

interface AuthState {
  token: string | null
  email: string | null
  initialized: boolean
  twoFactorEnabled: boolean
  twoFactorSetupPending: boolean
}

function loadFromStorage(): AuthState {
  if (import.meta.server) return { token: null, email: null, twoFactorEnabled: false, twoFactorSetupPending: false }
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw) as AuthState
  } catch {}
  return { token: null, email: null, twoFactorEnabled: false, twoFactorSetupPending: false }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({ ...loadFromStorage(), initialized: false }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setAuth(token: string, email: string, twoFactorEnabled = false, twoFactorSetupPending = false) {
      this.token = token
      this.email = email
      this.twoFactorEnabled = twoFactorEnabled
      this.twoFactorSetupPending = twoFactorSetupPending
      if (import.meta.client) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({
          token,
          email,
          twoFactorEnabled,
          twoFactorSetupPending,
        }))
      }
    },
    completeTwoFactorSetup() {
      this.twoFactorEnabled = true
      this.twoFactorSetupPending = false
      if (import.meta.client) {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (raw) {
          const data = JSON.parse(raw)
          localStorage.setItem(STORAGE_KEY, JSON.stringify({
            ...data,
            twoFactorEnabled: true,
            twoFactorSetupPending: false,
          }))
        }
      }
    },
    logout() {
      this.token = null
      this.email = null
      this.twoFactorEnabled = false
      this.twoFactorSetupPending = false
      if (import.meta.client) {
        localStorage.removeItem(STORAGE_KEY)
      }
    },
  },
})
