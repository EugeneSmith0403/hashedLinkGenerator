<template>
  <div class="center">
    <div class="card">
      <h1>{{ isRegister ? 'Create account' : 'Sign in' }}</h1>

      <form @submit.prevent="handleSubmit">
        <div v-if="isRegister" class="field">
          <label>Name</label>
          <input v-model="form.name" type="text" placeholder="John Doe" required />
        </div>

        <div class="field">
          <label>Email</label>
          <input v-model="form.email" type="email" placeholder="you@example.com" required />
        </div>

        <div class="field">
          <label>Password</label>
          <input v-model="form.password" type="password" placeholder="••••••••" required />
        </div>

        <div v-if="mutation.isError.value" class="alert error">
          {{ mutation.error.value?.message }}
        </div>

        <button type="submit" class="btn-primary" :disabled="mutation.isPending.value">
          {{ mutation.isPending.value ? 'Loading...' : (isRegister ? 'Register' : 'Login') }}
        </button>
      </form>

      <p class="toggle">
        {{ isRegister ? 'Already have an account?' : "Don't have an account?" }}
        <button class="btn-link" @click="isRegister = !isRegister">
          {{ isRegister ? 'Sign in' : 'Register' }}
        </button>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'
import { authApi } from '../api/auth.js'
import { authStore } from '../stores/auth.js'

const router = useRouter()
const isRegister = ref(false)

const form = reactive({ name: '', email: '', password: '' })

const mutation = useMutation({
  mutationFn: async () => {
    if (!isRegister.value) {
      return authApi.login(form.email, form.password)
    }
    await authApi.register(form.name, form.email, form.password)
    const loginData = await authApi.login(form.email, form.password)
    authStore.setAuth(loginData.token, loginData.email)
    await authApi.createAccount()
    return { _registered: true }
  },
  onSuccess(data) {
    if (data._registered) {
      router.push('/subscribe')
    } else {
      authStore.setAuth(data.token, data.email)
      router.push('/subscribe')
    }
  },
})

// reset error when switching mode
watch(isRegister, () => mutation.reset())

function handleSubmit() {
  mutation.mutate()
}
</script>

<style scoped>
.center {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 40px;
}

.card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
}

h1 { font-size: 22px; margin-bottom: 28px; }

.field { margin-bottom: 18px; }

label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #555;
  margin-bottom: 6px;
}

input {
  width: 100%;
  padding: 11px 14px;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}
input:focus { border-color: #635bff; }

.btn-primary {
  width: 100%;
  padding: 13px;
  background: #635bff;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
  transition: background 0.2s;
}
.btn-primary:hover:not(:disabled) { background: #4e47d1; }
.btn-primary:disabled { background: #a0a0c0; cursor: not-allowed; }

.alert {
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 14px;
}
.alert.error { background: #fff0f0; color: #c0392b; }

.toggle {
  margin-top: 20px;
  font-size: 14px;
  color: #666;
  text-align: center;
}

.btn-link {
  background: none;
  border: none;
  color: #635bff;
  font-size: 14px;
  cursor: pointer;
  font-weight: 600;
  padding: 0;
  margin-left: 4px;
}
.btn-link:hover { text-decoration: underline; }
</style>
