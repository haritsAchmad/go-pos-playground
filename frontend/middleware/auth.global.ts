export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie<string | null>('pos_access_token')

  if (!token.value && to.path !== '/login') {
    return navigateTo('/login')
  }

  if (token.value && to.path === '/login') {
    return navigateTo('/')
  }
})
