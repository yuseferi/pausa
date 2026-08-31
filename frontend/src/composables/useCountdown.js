// Reactive seconds-until-target countdown. Uses requestAnimationFrame so it
// stays smooth and is automatically throttled by the browser when the tab
// isn't visible. Recomputes from wall-clock each frame, so it is correct
// across system sleep / freeze.

import { ref, onUnmounted, watch } from 'vue'

export function useCountdown(targetIso) {
  const seconds = ref(0)

  let raf = null
  const tick = () => {
    if (!targetIso.value) {
      seconds.value = 0
    } else {
      const ms = new Date(targetIso.value).getTime() - Date.now()
      seconds.value = Math.max(0, Math.ceil(ms / 1000))
    }
    raf = requestAnimationFrame(tick)
  }

  const start = () => { if (raf == null) raf = requestAnimationFrame(tick) }
  const stop  = () => { if (raf != null) { cancelAnimationFrame(raf); raf = null } }

  watch(targetIso, () => {
    stop(); start()
  }, { immediate: true })

  onUnmounted(stop)

  return seconds
}

export function formatHMS(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds))
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}h ${m.toString().padStart(2, '0')}m`
  if (m === 0) return `0:${sec.toString().padStart(2, '0')}`
  return `${m}:${sec.toString().padStart(2, '0')}`
}
