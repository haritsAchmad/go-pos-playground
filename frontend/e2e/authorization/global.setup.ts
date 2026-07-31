import { expect, test as setup } from '@playwright/test'
import { accountRows, assertSafeTarget } from './env'

setup('create or reset isolated authorization accounts', async ({ request }) => {
  assertSafeTarget()
  const email = process.env.E2E_BOOTSTRAP_EMAIL
  const password = process.env.E2E_BOOTSTRAP_PASSWORD
  if (!email || !password) {
    throw new Error('E2E_BOOTSTRAP_EMAIL and E2E_BOOTSTRAP_PASSWORD are required for an admin account in the isolated test database.')
  }

  const login = await request.post('/api/auth/login', { data: { email, password } })
  expect(login.status(), 'bootstrap admin login').toBe(200)
  const token = String((await login.json()).data.access_token)
  const headers = { Authorization: `Bearer ${token}` }
  const list = await request.get('/api/users', { headers })
  expect(list.status(), 'list users during setup').toBe(200)
  const users = (await list.json()).data as Array<{ id: number; email: string }>

  for (const account of accountRows()) {
    const existing = users.find(user => user.email.toLowerCase() === account.email.toLowerCase())
    const response = existing
      ? await request.put(`/api/users/${existing.id}`, { headers, data: account })
      : await request.post('/api/users', { headers, data: account })
    expect(response.status(), `provision ${account.persona}`).toBe(existing ? 200 : 201)
  }
})
