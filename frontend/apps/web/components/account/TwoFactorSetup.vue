<script setup lang="ts">
const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const localePath = useLocalePath()

// --- State ---
const step = ref<'idle' | 'setup' | 'enabled'>('idle')
const qrImageUrl = ref('')
const code = ref('')
const codeError = ref('')
const loading = ref(false)

// On mount: if 2FA already enabled, show enabled state
onMounted(() => {
  if (auth.twoFactorEnabled) {
    step.value = 'enabled'
  }
})

async function startSetup() {
  loading.value = true
  codeError.value = ''
  try {
    const data = await useUserService().setup2FA()
    qrImageUrl.value = data.qrCode
    step.value = 'setup'
    code.value = ''
  } finally {
    loading.value = false
  }
}

async function verify() {
  if (!/^\d{6}$/.test(code.value)) {
    codeError.value = t('account.twoFactor.invalidCode')
    return
  }

  loading.value = true
  codeError.value = ''

  try {
    const res = await useUserService().verify2FA(code.value, auth.email ?? '')
    auth.setAuth(res.token, auth.email ?? '', true, false)
    router.push(localePath('/dashboard'))
  } catch {
    codeError.value = t('account.twoFactor.wrongCode')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UiCard>
    <div class="space-y-5">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold text-gray-900">{{ t('account.twoFactor.title') }}</h2>
          <p class="text-sm text-gray-500 mt-0.5">{{ t('account.twoFactor.subtitle') }}</p>
        </div>

        <!-- Status badge -->
        <UiBadge :variant="step === 'enabled' ? 'success' : 'default'">
          {{ step === 'enabled' ? t('account.twoFactor.enabled') : t('account.twoFactor.disabled') }}
        </UiBadge>
      </div>

      <!-- Idle: not yet set up -->
      <template v-if="step === 'idle'">
        <div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
          {{ t('account.twoFactor.notEnabledWarning') }}
        </div>
        <UiButton :loading="loading" @click="startSetup">
          {{ t('account.twoFactor.enableButton') }}
        </UiButton>
      </template>

      <!-- Setup: show QR + input -->
      <template v-else-if="step === 'setup'">
        <div class="space-y-4">
          <p class="text-sm text-gray-600">{{ t('account.twoFactor.scanInstruction') }}</p>

          <!-- QR Code -->
          <div class="flex justify-center">
            <div class="border border-gray-200 rounded-xl p-3 bg-white inline-block">
              <img
                :src="qrImageUrl"
                alt="2FA QR Code"
                width="200"
                height="200"
                class="block"
              />
            </div>
          </div>

          <!-- Code input -->
          <UiInput
            id="totp-code"
            v-model="code"
            :label="t('account.twoFactor.codeLabel')"
            type="text"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            :placeholder="t('account.twoFactor.codePlaceholder')"
            :error="codeError"
          />

          <div class="flex gap-3">
            <UiButton :loading="loading" @click="verify">
              {{ t('account.twoFactor.verifyButton') }}
            </UiButton>
            <UiButton variant="ghost" @click="step = 'idle'">
              {{ t('common.cancel') }}
            </UiButton>
          </div>
        </div>
      </template>

      <!-- Enabled -->
      <template v-else-if="step === 'enabled'">
        <div class="rounded-xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
          {{ t('account.twoFactor.enabledMessage') }}
        </div>
      </template>
    </div>
  </UiCard>
</template>
