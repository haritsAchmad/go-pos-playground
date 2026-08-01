import { expect, test } from '@playwright/test'
import { assertSafeTarget, emailFor, accountPassword, expiredTokenFor, login } from './env'

test.beforeAll(assertSafeTarget)

test('direct API access is rejected after logout and refresh session is revoked', async ({ request }) => {
  test.info().annotations.push(
    { type: 'endpoint', description: 'POST /auth/logout then GET /auth/me and POST /auth/refresh' },
    { type: 'account', description: 'cashier' },
    { type: 'expected', description: 'old access token and refresh session both rejected with 401 after logout' },
    { type: 'risk', description: 'high' },
    { type: 'root_cause_hint', description: 'Logout revokes refresh sessions, but issued stateless access JWTs remain valid until expiry.' },
  )
  const token = await login(request, 'cashier')
  expect((await request.post('/api/auth/logout')).status()).toBe(200)
  // Current documented architecture does not revoke already-issued access JWTs.
  const replay = await request.get('/api/auth/me', { headers: { Authorization: `Bearer ${token}` } })
  test.info().annotations.push({ type: 'actual', description: `access-token replay HTTP ${replay.status()}` })
  expect(replay.status(), 'access-token replay after logout should be rejected').toBe(401)
  expect((await request.post('/api/auth/refresh')).status()).toBe(401)
})

test('expired access token is rejected via direct API', async ({ request }) => {
  test.info().annotations.push(
    { type: 'endpoint', description: 'GET /auth/me' },
    { type: 'account', description: 'cashier with expired token' },
    { type: 'expected', description: '401 invalid or expired access token' },
    { type: 'risk', description: 'high' },
    { type: 'root_cause_hint', description: 'JWT parser must reject exp <= current time before reaching a protected handler.' },
  )
  const expired = expiredTokenFor(await login(request, 'cashier'))
  expect((await request.get('/api/auth/me', { headers: { Authorization: `Bearer ${expired}` } })).status()).toBe(401)
})

test('expired access token is rejected through the UI', async ({ page, request }) => {
  const expired = expiredTokenFor(await login(request, 'cashier'))
  await page.context().addCookies([{
    name: 'pos_access_token',
    value: expired,
    url: test.info().project.use.baseURL as string,
    sameSite: 'Strict',
  }])
  await page.goto('/kasir')
  await expect(page).toHaveURL(/\/login$/)
})

test('UI cannot be reused after logout', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill(emailFor('cashier'))
  await page.getByLabel('Password').fill(accountPassword)
  await page.getByRole('button', { name: /masuk/i }).click()
  await expect(page).toHaveURL('/')
  await page.getByRole('button', { name: 'Keluar' }).click()
  await expect(page).toHaveURL(/\/login$/)
  await page.goto('/kasir')
  await expect(page).toHaveURL(/\/login$/)
})
