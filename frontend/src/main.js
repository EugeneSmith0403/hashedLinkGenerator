import { createApp } from 'vue'
import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import LoginView from './views/LoginView.vue'
import SubscriptionView from './views/SubscriptionView.vue'
import { authStore } from './stores/auth.js'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
    mutations: { retry: false },
  },
})

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: () => (authStore.isLoggedIn ? '/subscribe' : '/login') },
    { path: '/login', component: LoginView },
    {
      path: '/subscribe',
      component: SubscriptionView,
      beforeEnter: () => (authStore.isLoggedIn ? true : '/login'),
    },
  ],
})

createApp(App).use(VueQueryPlugin, { queryClient }).use(router).mount('#app')
