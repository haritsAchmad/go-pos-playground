<script setup lang="ts">
const api = useKoperasiApi()
const form = reactive({ email: '', password: '' })
const error = ref('')
const submitting = ref(false)

onMounted(async () => {
  if (!api.token.value) return
  try {
    await api.me()
    await navigateTo('/')
  } catch {
    api.token.value = null
  }
})

async function login() {
  error.value = ''
  submitting.value = true
  try {
    const result = await api.login(form.email, form.password)
    api.token.value = result.access_token
    await navigateTo('/')
  } catch (reason: any) {
    error.value = reason?.data?.message || 'Login gagal'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="login-page">
    <form class="login-card" @submit.prevent="login">
      <div class="login-mark">K</div>
      <h1>Masuk ke Koperasi</h1>
      <p>Gunakan akun yang diberikan administrator.</p>
      <label>Email<input v-model.trim="form.email" type="email" autocomplete="username" required></label>
      <label>Password<input v-model="form.password" type="password" autocomplete="current-password" required></label>
      <p v-if="error" class="alert error">{{ error }}</p>
      <button class="primary" :disabled="submitting">{{ submitting ? 'Memproses…' : 'Masuk' }}</button>
    </form>
  </section>
</template>
