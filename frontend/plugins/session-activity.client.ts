// Records browser interaction locally. It does not call the backend by itself, keeping idle load at zero.
export default defineNuxtPlugin((nuxtApp) => {
  const lastActivity = useState<number>('session-last-activity', () => Date.now())
  let lastRecordedAt = 0

  const recordActivity = () => {
    const now = Date.now()
    // Throttle noisy pointer and keyboard events to one reactive update every 30 seconds.
    if (now-lastRecordedAt < 30_000) return
    lastRecordedAt = now
    lastActivity.value = now
  }

  for (const eventName of ['pointerdown', 'keydown', 'input']) {
    window.addEventListener(eventName, recordActivity, { passive: true })
  }

  // Route changes count as activity even when the destination has no immediate API request.
  nuxtApp.hook('page:start', recordActivity)
})
