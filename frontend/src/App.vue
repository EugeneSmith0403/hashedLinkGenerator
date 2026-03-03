<template>
  <div class="app">
    <header class="app-header">
      <span class="logo">⚡ SubApp</span>
      <div v-if="authStore.isLoggedIn" class="user-info">
        <span class="email">{{ authStore.email }}</span>
        <button class="btn-logout" @click="handleLogout">Logout</button>
      </div>
    </header>
    <main>
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { authStore } from './stores/auth.js'

const router = useRouter()

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style>
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background: #f0f2f5;
  color: #1a1a2e;
  min-height: 100vh;
}

.app-header {
  background: #fff;
  border-bottom: 1px solid #e8e8f0;
  padding: 0 32px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo { font-size: 18px; font-weight: 700; color: #635bff; }

.user-info { display: flex; align-items: center; gap: 12px; }
.email { font-size: 14px; color: #666; }

.btn-logout {
  padding: 6px 14px;
  background: transparent;
  border: 1.5px solid #ddd;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-logout:hover { border-color: #635bff; color: #635bff; }

main { padding: 40px 16px; }
</style>
