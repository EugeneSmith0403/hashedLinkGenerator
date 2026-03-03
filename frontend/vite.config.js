import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/auth': 'http://localhost:8081',
      '/subscriptions': 'http://localhost:8081',
      '/account': 'http://localhost:8081',
    },
  },
})
