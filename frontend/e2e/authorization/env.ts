import { expect, type APIRequestContext } from '@playwright/test'
import { createHmac } from 'node:crypto'
import { personas, type Persona } from './policy'

export const baseURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:3000'
export const runID = process.env.E2E_TEST_RUN_ID || 'local'
export const accountPassword = process.env.E2E_ACCOUNT_PASSWORD || 'Authz-Test-Only-42!'

export function assertSafeTarget() {
  const target = new URL(baseURL)
  const local = ['localhost', '127.0.0.1', '::1'].includes(target.hostname)
  if (!local && process.env.E2E_ALLOW_REMOTE_TEST_TARGET !== 'true') {
    throw new Error(`Refusing authorization tests against non-local target ${target.origin}. Set E2E_ALLOW_REMOTE_TEST_TARGET=true only for an isolated test environment.`)
  }
  if (/prod|production/i.test(target.hostname)) {
    throw new Error(`Refusing target whose hostname looks like production: ${target.hostname}`)
  }
}

export const emailFor = (persona: Persona) => `e2e-authz-${runID}-${persona}@example.test`

export async function login(request: APIRequestContext, persona: Persona) {
  const response = await request.post('/api/auth/login', {
    data: { email: emailFor(persona), password: accountPassword },
  })
  expect(response.status(), `login failed for ${persona}`).toBe(200)
  const payload = await response.json()
  return String(payload.data.access_token)
}

export function expiredTokenFor(validToken: string) {
  const secret = process.env.E2E_JWT_SECRET
  if (!secret || secret.length < 32) {
    throw new Error('E2E_JWT_SECRET (the isolated test backend JWT secret) is required to create a genuinely expired signed token.')
  }
  const [, encodedPayload] = validToken.split('.')
  const current = JSON.parse(Buffer.from(encodedPayload, 'base64url').toString('utf8'))
  const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url')
  const payload = Buffer.from(JSON.stringify({ ...current, iat: 1, exp: 2 })).toString('base64url')
  const unsigned = `${header}.${payload}`
  const signature = createHmac('sha256', secret).update(unsigned).digest('base64url')
  return `${unsigned}.${signature}`
}

export const accountRows = () =>
  (Object.keys(personas) as Persona[]).map(persona => ({
    persona,
    name: `E2E AUTHZ ${runID} ${persona}`,
    email: emailFor(persona),
    password: accountPassword,
    role: personas[persona].appRole,
    active: true,
  }))
