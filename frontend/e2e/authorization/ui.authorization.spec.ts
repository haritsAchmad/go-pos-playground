import { expect, test } from '@playwright/test'
import { accountPassword, assertSafeTarget, emailFor } from './env'
import { personas, type Persona, uiRoutes } from './policy'

const roles = Object.keys(personas) as Persona[]

test.beforeAll(assertSafeTarget)

for (const route of uiRoutes) {
  test.describe(`UI ${route.path}`, () => {
    test('redirects an unauthenticated browser to login', async ({ page }) => {
      await page.goto(route.path)
      await expect(page).toHaveURL(/\/login$/)
    })

    for (const persona of roles) {
      test(`${persona} direct navigation follows UI policy`, async ({ page }) => {
        test.info().annotations.push(
          { type: 'endpoint', description: `UI ${route.path}` },
          { type: 'account', description: `${persona} (${personas[persona].appRole})` },
          { type: 'allowed_roles', description: route.allowed.join(', ') },
          { type: 'expected', description: route.allowed.includes(persona) ? 'route remains accessible' : 'browser is redirected away from forbidden route' },
          { type: 'risk', description: route.path === '/pengguna' || route.path === '/audit' ? 'high' : 'medium' },
          { type: 'root_cause_hint', description: 'Global middleware checks only token presence; route-level role guards must reject direct navigation.' },
        )
        await page.goto('/login')
        await page.getByLabel('Email').fill(emailFor(persona))
        await page.getByLabel('Password').fill(accountPassword)
        await page.getByRole('button', { name: /masuk/i }).click()
        await expect(page).toHaveURL('/')
        const navigation = await page.goto(route.path)
        test.info().annotations.push({
          type: 'actual',
          description: `document HTTP ${navigation?.status() ?? 'unknown'}; final URL ${page.url()}`,
        })
        if (route.allowed.includes(persona)) {
          await expect(page).toHaveURL(new RegExp(`${route.path === '/' ? '/$' : `${route.path}$`}`))
          await expect(page.locator('body')).not.toContainText('you do not have permission')
        } else {
          await expect(page, `${persona} must not remain on forbidden UI route ${route.path}`).not.toHaveURL(new RegExp(`${route.path}$`))
        }
      })
    }
  })
}
