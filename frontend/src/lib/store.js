// Lightweight reactive store. Wraps Vue's reactive() so any component can
// `import { state } from './store'` and read or react to it. Mutations
// happen through named actions defined here so behavior stays discoverable.

import { reactive, computed, readonly, watch } from 'vue'
import { api } from './api'

const _state = reactive({
  ready: false,
  firstLaunch: false,
  config: null,
  snapshot: {
    phase: 'idle',
    nextKind: 'short',
    nextBreakAt: null,
    currentKind: '',
    breakEndsAt: null,
    shortsCompleted: 0,
    shortsUntilLong: 0,
    stats: { breaksTaken: 0, breaksSkipped: 0, breaksPostponed: 0, naturalBreaks: 0 },
  },
  // Live break info populated from EventBreakStart
  break: {
    durationSeconds: 0,
    secondsLeft: 0,
    tip: { id: 0, text: '', category: '' },
  },
  ui: {
    preferencesOpen: false,
  },
  notification: null, // {kind: 'short'|'long', secondsUntil} when scheduler:event fires EventNotify
})

export const state = readonly(_state)

export const onBreak     = computed(() => _state.snapshot.phase === 'on_break')
export const isPaused    = computed(() => _state.snapshot.phase === 'paused')
export const isAutoPaused = computed(() => _state.snapshot.phase === 'auto_paused')
export const isAnyPaused  = computed(() => isPaused.value || isAutoPaused.value)
export const autoPauseReason = computed(() => _state.snapshot.autoPauseReason || '')
export const themeMode = computed(() => {
  const t = _state.config?.display?.theme || 'system'
  if (t !== 'system') return t
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
})

export const accent = computed(() => _state.config?.display?.accentColor || '#0ea5e9')

// ---------- Actions ----------

export async function init() {
  if (window.__pausaInited) return
  window.__pausaInited = true

  const [config, snapshot, firstLaunch] = await Promise.all([
    api.getConfig(),
    api.getSnapshot(),
    api.isFirstLaunch(),
  ])
  if (config) _state.config = config
  if (snapshot) _state.snapshot = snapshot
  _state.firstLaunch = !!firstLaunch
  _state.ready = true

  api.on('scheduler:event', onSchedulerEvent)
  api.on('scheduler:hydrate', onSchedulerHydrate)
  api.on('config:updated', (cfg) => { if (cfg) _state.config = cfg })
  api.on('ui:open-preferences', () => { _state.ui.preferencesOpen = true })

  // Sync the document theme attribute whenever theme changes.
  watch(themeMode, (mode) => {
    document.documentElement.dataset.theme = mode
  }, { immediate: true })

  watch(accent, (color) => {
    document.documentElement.style.setProperty('--accent', color)
  }, { immediate: true })
}

function onSchedulerHydrate(snap) {
  if (snap) _state.snapshot = snap
}

function onSchedulerEvent(ev) {
  if (!ev) return
  _state.snapshot = ev.snapshot

  switch (ev.kind) {
    case 'breakStart':
      _state.break.durationSeconds = ev.durationSeconds || 0
      _state.break.secondsLeft = ev.durationSeconds || 0
      _state.break.tip = ev.tip || { id: 0, text: '', category: '' }
      _state.notification = null
      break
    case 'breakTick':
      _state.break.secondsLeft = ev.secondsLeft
      break
    case 'breakEnd':
    case 'skipped':
    case 'natural':
      _state.break.secondsLeft = 0
      break
    case 'notify':
      _state.notification = { kind: ev.breakKind, secondsUntil: ev.secondsUntil }
      break
    case 'paused':
    case 'resumed':
    case 'postponed':
      _state.notification = null
      break
  }
}

export async function saveConfig(cfg) {
  const saved = await api.saveConfig(cfg)
  if (saved) _state.config = saved
  return saved
}

export async function completeWelcome() {
  await api.markFirstLaunchComplete()
  _state.firstLaunch = false
}

export const ui = {
  openPreferences() { _state.ui.preferencesOpen = true },
  closePreferences() { _state.ui.preferencesOpen = false },
}

export { api }
