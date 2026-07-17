import { defineNuxtConfig } from 'nuxt/config'

export default defineNuxtConfig({
  compatibilityDate: '2025-05-15',
  devtools: { enabled: true },
  app: {
    head: {
      title: 'Koperasi Operational Console',
      link: [{ rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
    },
  },
  css: ['~/assets/css/main.css', '~/assets/css/improvements.css', '~/assets/css/payment.css'],
  runtimeConfig: {
    public: { apiBase: '/api' },
  },
  routeRules: { '/api/**': { proxy: 'http://localhost:8082/**' } },
})
