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

const { handleSubmit, errors } = useForm({
  validationSchema: toTypedSchema(loginSchema),
})

const { value: email, handleBlur: blurEmail, handleChange: changeEmail } = useField<string>('email')
const { value: password, handleBlur: blurPassword, handleChange: changePassword } = useField<string>('password')

const { mutate, isPending, error } = useMutation({
  mutationFn: () => useAuthService().login({ email: email.value, password: password.value }),
  onSuccess(data) {
    queryClient.clear()
    auth.setAuth(data.token, data.email)
    router.push(localePath('/dashboard'))
  },
})

const onSubmit = handleSubmit(() => mutate())
</script>

<template>
  <form class="space-y-5" @submit.prevent="onSubmit">
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

    <p v-if="error" class="text-sm text-red-600">{{ (error as Error).message }}</p>

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
</template>
