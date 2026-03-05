export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxtjs/i18n',
    '@pinia/nuxt',
  ],

  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE ?? 'http://localhost:8081',
      stripeKey: process.env.NUXT_PUBLIC_STRIPE_KEY ?? '',
    },
  },

  components: [
    { path: '~/components', pathPrefix: false },
  ],

  i18n: {
    locales: [
      { code: 'en', name: 'English', file: 'en.json' },
      { code: 'ru', name: 'Русский', file: 'ru.json' },
      { code: 'de', name: 'Deutsch', file: 'de.json' },
    ],
    defaultLocale: 'en',
    strategy: 'prefix_except_default',
    langDir: 'locales/',
    lazy: true,
  },

  pinia: {
    storesDirs: ['./stores/**'],
  },

  imports: {
    dirs: ['stores', 'composables', 'services'],
  },

  typescript: {
    strict: true,
  },
})
