<script setup lang="ts">
import type { UserMe } from '~/services/user'

const { t } = useI18n()
const auth = useAuthStore()

const user = ref<UserMe | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    user.value = await useUserService().getMe()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="max-w-2xl space-y-8">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">{{ t('account.title') }}</h1>
      <p class="text-sm text-gray-500 mt-1">{{ t('account.subtitle') }}</p>
    </div>

    <!-- Setup required banner -->
    <div
      v-if="auth.twoFactorSetupPending"
      class="rounded-xl border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-700"
    >
      {{ t('account.setupRequiredBanner') }}
    </div>

    <!-- Profile info -->
    <UiCard>
      <div class="space-y-3">
        <h2 class="text-base font-semibold text-gray-900">{{ t('account.profileTitle') }}</h2>

        <div v-if="loading" class="text-sm text-gray-400">{{ t('common.loading') }}</div>

        <div v-else class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-semibold text-sm">
            {{ user?.name?.[0]?.toUpperCase() ?? user?.email?.[0]?.toUpperCase() ?? '?' }}
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900">{{ user?.name }}</p>
            <p class="text-xs text-gray-500">{{ user?.email }}</p>
          </div>
        </div>
      </div>
    </UiCard>

    <!-- 2FA -->
    <TwoFactorSetup />
  </div>
</template>
