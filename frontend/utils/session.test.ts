import { describe, expect, it, vi } from 'vitest'
import {
  ACTIVITY_THROTTLE_MS,
  createActivityRecorder,
  getSessionRefreshAction,
  getTokenExpiry,
  SESSION_REFRESH_WINDOW_SECONDS,
} from './session'

function tokenWithExpiry(expiresAt: number) {
  const payload = btoa(JSON.stringify({ exp: expiresAt }))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
  return `header.${payload}.signature`
}

describe('getTokenExpiry', () => {
  it('reads the expiry claim from a base64url JWT payload', () => {
    expect(getTokenExpiry(tokenWithExpiry(1_800_000_000))).toBe(1_800_000_000)
  })

  it.each([null, '', 'invalid-token'])('returns zero for an invalid token', (token) => {
    expect(getTokenExpiry(token)).toBe(0)
  })
})

describe('getSessionRefreshAction', () => {
  const now = 1_000

  it('keeps a token that is outside the refresh window', () => {
    expect(getSessionRefreshAction(now + SESSION_REFRESH_WINDOW_SECONDS + 1, now)).toBe('valid')
  })

  it('refreshes a token at the refresh-window boundary', () => {
    expect(getSessionRefreshAction(now + SESSION_REFRESH_WINDOW_SECONDS, now)).toBe('refresh')
  })

  it('rejects an expired token', () => {
    expect(getSessionRefreshAction(now, now)).toBe('expired')
  })
})

describe('createActivityRecorder', () => {
  it('records immediately and throttles noisy activity until the boundary', () => {
    const onActivity = vi.fn()
    let now = 10_000
    const recordActivity = createActivityRecorder(onActivity, ACTIVITY_THROTTLE_MS, () => now)

    expect(recordActivity()).toBe(true)
    now += ACTIVITY_THROTTLE_MS - 1
    expect(recordActivity()).toBe(false)
    now += 1
    expect(recordActivity()).toBe(true)
    expect(onActivity).toHaveBeenCalledTimes(2)
    expect(onActivity).toHaveBeenLastCalledWith(now)
  })
})
