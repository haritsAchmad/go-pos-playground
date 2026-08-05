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

export function createPaymentPoller(options: {
  poll: () => Promise<void>
  intervalMs?: number
  setIntervalFn?: typeof setInterval
  clearIntervalFn?: typeof clearInterval
}) {
  const intervalMs = options.intervalMs ?? 3000
  const setIntervalFn = options.setIntervalFn ?? setInterval
  const clearIntervalFn = options.clearIntervalFn ?? clearInterval
  let timer: ReturnType<typeof setInterval> | null = null
  let inFlight = false

  async function run() {
    if (inFlight) return
    inFlight = true
    try {
      await options.poll()
    } finally {
      inFlight = false
    }
  }

  function start() {
    if (timer) return
    void run()
    timer = setIntervalFn(() => void run(), intervalMs)
  }

  function stop() {
    if (!timer) return
    clearIntervalFn(timer)
    timer = null
  }

  return { run, start, stop }
}
