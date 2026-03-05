export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const auth = useAuthStore()

  // Route names include locale suffix: 'auth-login___ru', 'auth-login' (default)
  const routeName = String(to.name ?? '')
  const isPublic = routeName.includes('auth-login') || routeName.includes('auth-register')

  if (!auth.isLoggedIn && !isPublic) {
    return navigateTo('/auth/login')
  }

  if (auth.isLoggedIn && isPublic) {
    return navigateTo('/dashboard')
  }
})
