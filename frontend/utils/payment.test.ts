import { afterEach, describe, expect, it, vi } from 'vitest'
import { createPaymentPoller, formatPaymentCountdown, isTerminalPaymentStatus, paymentSecondsRemaining } from './payment'

afterEach(() => vi.useRealTimers())

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

  it('polls immediately, repeats, and stops cleanly', async () => {
    vi.useFakeTimers()
    const poll = vi.fn(async () => {})
    const poller = createPaymentPoller({ poll })

    poller.start()
    await vi.advanceTimersByTimeAsync(0)
    expect(poll).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(6000)
    expect(poll).toHaveBeenCalledTimes(3)

    poller.stop()
    await vi.advanceTimersByTimeAsync(6000)
    expect(poll).toHaveBeenCalledTimes(3)
  })

  it('does not overlap slow polling requests', async () => {
    vi.useFakeTimers()
    let finish: (() => void) | undefined
    const poll = vi.fn(() => new Promise<void>((resolve) => { finish = resolve }))
    const poller = createPaymentPoller({ poll })

    poller.start()
    await vi.advanceTimersByTimeAsync(9000)
    expect(poll).toHaveBeenCalledTimes(1)

    finish?.()
    await Promise.resolve()
    await vi.advanceTimersByTimeAsync(3000)
    expect(poll).toHaveBeenCalledTimes(2)
    poller.stop()
  })
})
