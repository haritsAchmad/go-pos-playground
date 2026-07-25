export const SESSION_REFRESH_WINDOW_SECONDS = 300
export const ACTIVITY_THROTTLE_MS = 30_000

export type SessionRefreshAction = 'valid' | 'refresh' | 'expired'

export function getTokenExpiry(token: string | null | undefined): number {
  try {
    const payload = token?.split('.')[1]
    if (!payload) return 0

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const claims = JSON.parse(atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')))
    return Number(claims.exp || 0)
  } catch {
    return 0
  }
}

export function getSessionRefreshAction(
  expiresAt: number,
  nowSeconds = Math.floor(Date.now() / 1000),
): SessionRefreshAction {
  const secondsLeft = expiresAt - nowSeconds
  if (secondsLeft <= 0) return 'expired'
  if (secondsLeft <= SESSION_REFRESH_WINDOW_SECONDS) return 'refresh'
  return 'valid'
}

export function createActivityRecorder(
  onActivity: (recordedAt: number) => void,
  throttleMs = ACTIVITY_THROTTLE_MS,
  now: () => number = Date.now,
) {
  let lastRecordedAt: number | null = null

  return () => {
    const recordedAt = now()
    if (lastRecordedAt !== null && recordedAt - lastRecordedAt < throttleMs) return false

    lastRecordedAt = recordedAt
    onActivity(recordedAt)
    return true
  }
}
