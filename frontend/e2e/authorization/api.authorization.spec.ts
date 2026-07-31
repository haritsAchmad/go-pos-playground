import { expect, test } from '@playwright/test'
import { assertSafeTarget, login } from './env'
import { personas, protectedRoutes, replaceIds, type Persona } from './policy'

const sentinelA = 2_147_483_646
const sentinelB = 2_147_483_647
const roles = Object.keys(personas) as Persona[]

test.beforeAll(assertSafeTarget)

for (const route of protectedRoutes) {
  test.describe(`${route.method} ${route.path}`, () => {
    test('authorization matrix via direct API request', async ({ request }) => {
      test.info().annotations.push(
        { type: 'endpoint', description: `${route.method} ${route.path}` },
        { type: 'allowed_roles', description: route.allowed.join(', ') },
        { type: 'expected', description: 'anonymous=401; disallowed=403; allowed reaches handler (not 401/403)' },
        { type: 'risk', description: route.risk },
        { type: 'root_cause_hint', description: 'Check router middleware order, role policy, and current database role lookup.' },
      )
      const path = `/api${replaceIds(route.path, sentinelA, sentinelB)}`
      const observations: Array<{ account: string; status: number }> = []
      const anonymous = await request.fetch(path, { method: route.method, data: route.body })
      observations.push({ account: 'unauthenticated', status: anonymous.status() })
      expect.soft(anonymous.status(), 'unauthenticated request must be rejected').toBe(401)

      for (const persona of roles) {
        const token = await login(request, persona)
        const response = await request.fetch(path, {
          method: route.method,
          headers: { Authorization: `Bearer ${token}` },
          data: route.body,
        })
        test.info().annotations.push({
          type: 'account',
          description: `${persona} (${personas[persona].appRole}) => HTTP ${response.status()}`,
        })
        observations.push({ account: persona, status: response.status() })
        if (route.allowed.includes(persona)) {
          expect.soft([401, 403], `${persona} is allowed; handler may return 2xx/4xx validation or not-found, but not an auth rejection`).not.toContain(response.status())
        } else {
          expect.soft(response.status(), `${persona} is not allowed`).toBe(403)
        }
      }
      await test.info().attach('authorization-observations.json', {
        body: Buffer.from(JSON.stringify({ method: route.method, path, observations }, null, 2)),
        contentType: 'application/json',
      })
    })

    if (route.objectLevel) {
      test('object-level decision is unchanged when resource ID is replaced', async ({ request }) => {
        test.info().annotations.push(
          { type: 'endpoint', description: `${route.method} ${route.path}` },
          { type: 'allowed_roles', description: route.allowed.join(', ') },
          { type: 'expected', description: 'same RBAC decision for both substituted resource IDs' },
          { type: 'risk', description: route.risk },
          { type: 'root_cause_hint', description: 'A differing 401/403 decision by ID suggests object-level authorization inconsistency or middleware bypass.' },
        )
        const observations: Array<{ account: string; ids: number[]; statuses: number[] }> = []
        for (const persona of roles) {
          const token = await login(request, persona)
          const statuses = []
          for (const id of [sentinelA, sentinelB]) {
            const response = await request.fetch(`/api${replaceIds(route.path, id, id)}`, {
              method: route.method,
              headers: { Authorization: `Bearer ${token}` },
              data: route.body,
            })
            statuses.push(response.status())
          }
          observations.push({ account: persona, ids: [sentinelA, sentinelB], statuses })
          if (route.allowed.includes(persona)) {
            expect.soft(statuses.every(status => status !== 401 && status !== 403), `${persona} should pass RBAC for both resource IDs`).toBeTruthy()
          } else {
            expect.soft(statuses, `${persona} must be denied for both resource IDs`).toEqual([403, 403])
          }
        }
        await test.info().attach('object-level-observations.json', {
          body: Buffer.from(JSON.stringify({ method: route.method, path: route.path, observations }, null, 2)),
          contentType: 'application/json',
        })
      })
    }
  })
}
