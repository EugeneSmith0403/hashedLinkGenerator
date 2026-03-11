export default defineNuxtPlugin(() => {
  const auth = useAuthStore()
  try {
    const raw = localStorage.getItem('auth')
    if (raw) {
      const { token, email, twoFactorEnabled, twoFactorSetupPending } = JSON.parse(raw)
      if (token && email) {
        auth.token = token
        auth.email = email
        auth.twoFactorEnabled = !!twoFactorEnabled
        auth.twoFactorSetupPending = !!twoFactorSetupPending
      }
    }
  } catch {}
  auth.initialized = true
})
