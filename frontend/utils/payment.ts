export const terminalPaymentStatuses = ['PAID', 'FAILED', 'EXPIRED'] as const

export function isTerminalPaymentStatus(status: unknown): boolean {
  return terminalPaymentStatuses.includes(String(status).toUpperCase() as typeof terminalPaymentStatuses[number])
}

export function paymentSecondsRemaining(expiresAt: unknown, now = Date.now()): number {
  const expiry = new Date(String(expiresAt)).getTime()
  if (!Number.isFinite(expiry)) return 0
  return Math.max(0, Math.ceil((expiry - now) / 1000))
}

export function formatPaymentCountdown(seconds: number): string {
  const safeSeconds = Math.max(0, Math.floor(seconds))
  const minutes = Math.floor(safeSeconds / 60)
  const remainder = safeSeconds % 60
  return `${String(minutes).padStart(2, '0')}:${String(remainder).padStart(2, '0')}`
}
