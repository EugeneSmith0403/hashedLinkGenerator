export default defineNuxtPlugin(() => {
  const auth = useAuthStore()
  try {
    const raw = localStorage.getItem('auth')
    if (raw) {
      const { token, email } = JSON.parse(raw)
      if (token && email) {
        auth.token = token
        auth.email = email
      }
    }
  } catch {}
  auth.initialized = true
})
