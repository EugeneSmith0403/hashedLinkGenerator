<script setup lang="ts">
import { useQueryClient } from '@tanstack/vue-query'

const auth = useAuthStore()
const { t } = useI18n()
const router = useRouter()
const localePath = useLocalePath()
const queryClient = useQueryClient()

function logout() {
  auth.logout()
  queryClient.clear()
  router.push(localePath('/auth/login'))
}

const allNav = [
  { to: '/dashboard', icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1', labelKey: 'common.links', requiresTwoFactor: true },
  { to: '/billing', icon: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z', labelKey: 'common.billing', requiresTwoFactor: true },
  { to: '/payments', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2', labelKey: 'common.payments', requiresTwoFactor: true },
  { to: '/account', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z', labelKey: 'common.account', requiresTwoFactor: false },
]

const nav = computed(() =>
  allNav.filter(item => !item.requiresTwoFactor || auth.twoFactorEnabled),
)
</script>

<template>
  <div class="flex h-screen bg-gray-50">
    <!-- Sidebar -->
    <aside class="w-60 bg-white border-r border-gray-200 flex flex-col">
      <div class="px-5 py-5 border-b border-gray-200">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center">
            <svg class="h-4 w-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
          </div>
          <span class="font-semibold text-gray-900">LinkShort</span>
        </div>
      </div>

      <nav class="flex-1 px-3 py-4 space-y-1">
        <NuxtLink
          v-for="item in nav"
          :key="item.to"
          :to="localePath(item.to)"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          exact-active-class="bg-indigo-50 text-indigo-700"
          inactive-class="text-gray-600 hover:bg-gray-100 hover:text-gray-900"
        >
          <svg class="h-5 w-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon" />
          </svg>
          {{ t(item.labelKey) }}
        </NuxtLink>
      </nav>

      <div class="px-3 py-4 border-t border-gray-200 space-y-1">
        <div class="px-3 py-2">
          <p class="text-xs text-gray-400 truncate">{{ auth.email }}</p>
        </div>
        <button
          class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-gray-600 hover:bg-red-50 hover:text-red-700 transition-colors"
          @click="logout"
        >
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          {{ t('common.logout') }}
        </button>
      </div>
    </aside>

    <!-- Main area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <header class="bg-white border-b border-gray-200 px-6 py-3.5 flex items-center justify-end gap-3">
        <LocaleSwitcher />
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
