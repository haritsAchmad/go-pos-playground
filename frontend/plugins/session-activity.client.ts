import { createActivityRecorder } from '~/utils/session'

// Records browser interaction locally. It does not call the backend by itself, keeping idle load at zero.
export default defineNuxtPlugin((nuxtApp) => {
  const lastActivity = useState<number>('session-last-activity', () => Date.now())
  // Throttle noisy pointer and keyboard events to one reactive update every 30 seconds.
  const recordActivity = createActivityRecorder((recordedAt) => {
    lastActivity.value = recordedAt
  })

  for (const eventName of ['pointerdown', 'keydown', 'input']) {
    window.addEventListener(eventName, recordActivity, { passive: true })
  }

  // Route changes count as activity even when the destination has no immediate API request.
  const removePageHook = nuxtApp.hook('page:start', () => {
    recordActivity()
  })

  const cleanup = () => {
    for (const eventName of ['pointerdown', 'keydown', 'input']) {
      window.removeEventListener(eventName, recordActivity)
    }
    removePageHook()
  }

  // Nuxt apps live for the browser tab lifetime; dispose listeners when this plugin is hot-reloaded.
  if (import.meta.hot) import.meta.hot.dispose(cleanup)
})
