import { describe, it, expect, vi, beforeAll } from 'vitest'

// The store reaches the Wails bindings through ./api. Stub the whole module so
// tests never import wailsjs/ or touch the runtime.
vi.mock('./api', () => ({
  api: {
    getConfig: vi.fn(),
    getSnapshot: vi.fn(),
    getCapabilities: vi.fn(),
    isFirstLaunch: vi.fn(),
    saveConfig: vi.fn(),
    markFirstLaunchComplete: vi.fn(),
    on: vi.fn(() => () => {}),
    off: vi.fn(),
  },
}))

import { api } from './api'
import { state, onBreak, ui, init, saveConfig, completeWelcome } from './store'

const handlers = {}
api.on.mockImplementation((event, handler) => {
  handlers[event] = handler
  return () => {}
})

const baseSnapshot = {
  phase: 'scheduled',
  nextKind: 'short',
  nextBreakAt: '2026-01-01T00:10:00Z',
  currentKind: '',
  breakEndsAt: null,
  shortsCompleted: 0,
  shortsUntilLong: 0,
  stats: { breaksTaken: 0, breaksSkipped: 0, breaksPostponed: 0, naturalBreaks: 0 },
}

beforeAll(async () => {
  api.getConfig.mockResolvedValue({ display: { theme: 'system', accentColor: '#0ea5e9' } })
  api.getSnapshot.mockResolvedValue(baseSnapshot)
  api.isFirstLaunch.mockResolvedValue(true)
  // A lite (Windows/Linux) platform: the store's all-true defaults must be
  // narrowed by what the backend reports.
  api.getCapabilities.mockResolvedValue({
    nativeOverlays: false,
    systemTray: false,
    dockIcon: false,
    loginItem: false,
    busyDetection: false,
    idleDetection: false,
  })
  await init()
})

describe('store', () => {
  it('hydrates config, snapshot, and first-launch state on init', () => {
    expect(state.ready).toBe(true)
    expect(state.config.display.accentColor).toBe('#0ea5e9')
    expect(state.snapshot.phase).toBe('scheduled')
    expect(state.firstLaunch).toBe(true)
    expect(state.capabilities.nativeOverlays).toBe(false)
    expect(state.capabilities.systemTray).toBe(false)
    expect(state.capabilities.idleDetection).toBe(false)
    expect(handlers['scheduler:event']).toBeTypeOf('function')
    expect(handlers['scheduler:hydrate']).toBeTypeOf('function')
    expect(handlers['config:updated']).toBeTypeOf('function')
  })

  it('tracks the break lifecycle from scheduler events', () => {
    handlers['scheduler:event']({
      kind: 'breakStart',
      snapshot: { ...baseSnapshot, phase: 'on_break', currentKind: 'short' },
      breakKind: 'short',
      durationSeconds: 20,
      secondsLeft: 20,
      tip: { id: 1, text: 'Look away', category: 'eyes' },
    })
    expect(onBreak.value).toBe(true)
    expect(state.break.durationSeconds).toBe(20)
    expect(state.break.secondsLeft).toBe(20)
    expect(state.break.tip.text).toBe('Look away')

    handlers['scheduler:event']({
      kind: 'breakTick',
      snapshot: { ...baseSnapshot, phase: 'on_break' },
      secondsLeft: 7,
    })
    expect(state.break.secondsLeft).toBe(7)

    handlers['scheduler:event']({ kind: 'breakEnd', snapshot: baseSnapshot })
    expect(onBreak.value).toBe(false)
    expect(state.break.secondsLeft).toBe(0)
  })

  it('opens and closes preferences', () => {
    ui.openPreferences()
    expect(state.ui.preferencesOpen).toBe(true)
    ui.closePreferences()
    expect(state.ui.preferencesOpen).toBe(false)
  })

  it('applies the config returned by the backend on save', async () => {
    api.saveConfig.mockResolvedValue({ display: { theme: 'dark', accentColor: '#ef4444' } })
    await saveConfig({ display: { theme: 'dark' } })
    expect(state.config.display.theme).toBe('dark')
    expect(state.config.display.accentColor).toBe('#ef4444')
  })

  it('clears firstLaunch after the welcome completes', async () => {
    api.markFirstLaunchComplete.mockResolvedValue(undefined)
    await completeWelcome()
    expect(state.firstLaunch).toBe(false)
  })
})
