export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const auth = useAuthStore()
  const localePath = useLocalePath()

  // Route names include locale suffix: 'auth-login___ru', 'auth-login' (default)
  const routeName = String(to.name ?? '')
  const isPublic = routeName.includes('auth-login') || routeName.includes('auth-register')
  const isAccount = routeName.includes('account')

  if (!auth.isLoggedIn && !isPublic) {
    return navigateTo(localePath('/auth/login'))
  }

  if (auth.isLoggedIn && isPublic) {
    return navigateTo(localePath('/dashboard'))
  }

  // Block all pages except /account until 2FA is enabled and confirmed
  if (auth.isLoggedIn && !auth.twoFactorEnabled && !isAccount) {
    return navigateTo(localePath('/account'))
  }
})
