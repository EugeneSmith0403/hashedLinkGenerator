<script setup lang="ts">
import { useForm, useField } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { loginSchema } from '~/schemas/auth'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const localePath = useLocalePath()
const queryClient = useQueryClient()
const config = useRuntimeConfig()

// Step 1: credentials | Step 2: 2FA code
const step = ref<'credentials' | 'twoFactor'>('credentials')

// Temporary token stored after login, before 2FA is confirmed
const pendingToken = ref<string | null>(null)
const pendingEmail = ref<string | null>(null)

const { handleSubmit, errors } = useForm({
  validationSchema: toTypedSchema(loginSchema),
})

const { value: email, handleBlur: blurEmail, handleChange: changeEmail } = useField<string>('email')
const { value: password, handleBlur: blurPassword, handleChange: changePassword } = useField<string>('password')

const loginError = ref('')

const { mutate, isPending } = useMutation({
  mutationFn: () => useAuthService().login({ email: email.value, password: password.value }),
  onSuccess(data) {
    queryClient.clear()
    loginError.value = ''

    if (data.is2faEnabled) {
      pendingToken.value = data.token
      pendingEmail.value = data.email
      step.value = 'twoFactor'
      return
    }

    auth.setAuth(data.token, data.email, false)
    router.push(localePath('/dashboard'))
  },
  onError(err) {
    loginError.value = (err as Error).message
  },
})

const onSubmit = handleSubmit(() => mutate())

// --- 2FA step ---
const totpCode = ref('')
const totpError = ref('')
const totpPending = ref(false)

async function submitTotpCode() {
  if (!/^\d{6}$/.test(totpCode.value)) {
    totpError.value = t('account.twoFactor.invalidCode')
    return
  }
  totpError.value = ''
  totpPending.value = true

  try {
    const res = await fetch(`${config.public.apiBase}/users/2fa/verify`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${pendingToken.value}`,
      },
      body: JSON.stringify({ code: totpCode.value, email: pendingEmail.value }),
    })

    if (!res.ok) {
      totpError.value = t('account.twoFactor.wrongCode')
      return
    }

    const data = await res.json()
    auth.setAuth(data.token, pendingEmail.value!, true)
    router.push(localePath('/dashboard'))
  }
  catch {
    totpError.value = t('account.twoFactor.wrongCode')
  }
  finally {
    totpPending.value = false
  }
}

function backToCredentials() {
  step.value = 'credentials'
  totpCode.value = ''
  totpError.value = ''
  pendingToken.value = null
  pendingEmail.value = null
}
</script>

<template>
  <!-- Step 1: email + password -->
  <form v-if="step === 'credentials'" class="space-y-5" @submit.prevent="onSubmit">
    <div>
      <h2 class="text-xl font-bold text-gray-900">{{ t('auth.loginTitle') }}</h2>
      <p class="text-sm text-gray-500 mt-1">{{ t('auth.loginSubtitle') }}</p>
    </div>

    <UiInput
      id="email"
      v-model="email"
      :label="t('auth.email')"
      type="email"
      placeholder="you@example.com"
      required
      :error="errors.email"
      @blur="blurEmail"
      @change="changeEmail"
    />

    <UiInput
      id="password"
      v-model="password"
      :label="t('auth.password')"
      type="password"
      placeholder="••••••••"
      required
      :error="errors.password"
      @blur="blurPassword"
      @change="changePassword"
    />

    <p v-if="loginError" class="text-sm text-red-600">{{ loginError }}</p>

    <UiButton type="submit" class="w-full" :loading="isPending">
      {{ t('auth.login') }}
    </UiButton>

    <p class="text-center text-sm text-gray-500">
      {{ t('auth.noAccount') }}
      <NuxtLink to="/auth/register" class="text-indigo-600 hover:underline font-medium">
        {{ t('auth.register') }}
      </NuxtLink>
    </p>
  </form>

  <!-- Step 2: 2FA code -->
  <div v-else-if="step === 'twoFactor'" class="space-y-5">
    <div>
      <h2 class="text-xl font-bold text-gray-900">{{ t('auth.twoFactorTitle') }}</h2>
      <p class="text-sm text-gray-500 mt-1">{{ t('auth.twoFactorSubtitle') }}</p>
    </div>

    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1.5">
        {{ t('account.twoFactor.codeLabel') }}
      </label>
      <input
        v-model="totpCode"
        type="text"
        inputmode="numeric"
        autocomplete="one-time-code"
        maxlength="6"
        :placeholder="t('account.twoFactor.codePlaceholder')"
        class="w-full rounded-lg border border-gray-300 px-3 py-2 text-center text-2xl tracking-[0.5em] font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
      />
      <p v-if="totpError" class="mt-1.5 text-sm text-red-600">{{ totpError }}</p>
    </div>

    <UiButton class="w-full" :loading="totpPending" @click="submitTotpCode">
      {{ t('auth.twoFactorVerify') }}
    </UiButton>

    <button
      class="w-full text-sm text-gray-500 hover:text-gray-700 text-center"
      @click="backToCredentials"
    >
      {{ t('auth.backToLogin') }}
    </button>
  </div>
</template>
