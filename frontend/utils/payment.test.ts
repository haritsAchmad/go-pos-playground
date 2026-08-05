import { describe, expect, it } from 'vitest'
import { formatPaymentCountdown, isTerminalPaymentStatus, paymentSecondsRemaining } from './payment'

describe('payment lifecycle utilities', () => {
  it('rounds a partial remaining second up and stops at zero', () => {
    const now = Date.parse('2026-08-05T10:00:00.000Z')
    expect(paymentSecondsRemaining('2026-08-05T10:00:01.001Z', now)).toBe(2)
    expect(paymentSecondsRemaining('2026-08-05T09:59:59.000Z', now)).toBe(0)
    expect(paymentSecondsRemaining('invalid', now)).toBe(0)
  })

  it('formats countdowns consistently', () => {
    expect(formatPaymentCountdown(65)).toBe('01:05')
    expect(formatPaymentCountdown(-1)).toBe('00:00')
  })

  it('recognizes terminal statuses', () => {
    expect(isTerminalPaymentStatus('PAID')).toBe(true)
    expect(isTerminalPaymentStatus('expired')).toBe(true)
    expect(isTerminalPaymentStatus('PENDING')).toBe(false)
  })
})
