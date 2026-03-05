<script setup lang="ts">
import { useForm, useField } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { registerSchema } from '~/schemas/auth'
import { useMutation } from '@tanstack/vue-query'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()

const { handleSubmit, errors } = useForm({
  validationSchema: toTypedSchema(registerSchema),
})

const { value: name, handleBlur: blurName, handleChange: changeName } = useField<string>('name')
const { value: email, handleBlur: blurEmail, handleChange: changeEmail } = useField<string>('email')
const { value: password, handleBlur: blurPassword, handleChange: changePassword } = useField<string>('password')

const { mutate, isPending, error } = useMutation({
  mutationFn: () =>
    useAuthService().register({ name: name.value, email: email.value, password: password.value }),
  async onSuccess(data) {
    auth.setAuth(data.token, data.email)
    await useAccountService().create()
    router.push('/dashboard')
  },
})

const onSubmit = handleSubmit(() => mutate())
</script>

<template>
  <form class="space-y-5" @submit.prevent="onSubmit">
    <div>
      <h2 class="text-xl font-bold text-gray-900">{{ t('auth.registerTitle') }}</h2>
      <p class="text-sm text-gray-500 mt-1">{{ t('auth.registerSubtitle') }}</p>
    </div>

    <UiInput
      id="name"
      v-model="name"
      :label="t('auth.name')"
      type="text"
      :placeholder="t('auth.namePlaceholder')"
      required
      :error="errors.name"
      @blur="blurName"
      @change="changeName"
    />

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
      {{ t('auth.createAccount') }}
    </UiButton>

    <p class="text-center text-sm text-gray-500">
      {{ t('auth.haveAccount') }}
      <NuxtLink to="/auth/login" class="text-indigo-600 hover:underline font-medium">
        {{ t('auth.login') }}
      </NuxtLink>
    </p>
  </form>
</template>
