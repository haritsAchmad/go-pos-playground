export default defineNuxtRouteMiddleware(async (to) => {
  const token = useCookie<string | null>('pos_access_token')

  if (!token.value && to.path !== '/login') {
    return navigateTo('/login')
  }

  if (!token.value) return

  let role = ''
  try {
    const payload = await useRequestFetch()<{
      data: { role: string }
    }>('/api/auth/me', {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    role = payload.data.role
  } catch {
    token.value = null
    if (to.path !== '/login') return navigateTo('/login')
    return
  }

  if (to.path === '/login') return navigateTo('/')
  if (!isUIRouteAllowed(to.path, role)) return navigateTo('/')
})
