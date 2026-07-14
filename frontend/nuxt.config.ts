export default defineNuxtConfig({
  compatibilityDate: '2025-05-15',
  devtools: { enabled: true },
  css: ['~/assets/css/main.css', '~/assets/css/improvements.css', '~/assets/css/payment.css'],
  runtimeConfig: {
    public: { apiBase: process.env.NUXT_PUBLIC_API_BASE || '/api' },
  },
  routeRules: { '/api/**': { proxy: 'http://localhost:8080/**' } },
})
