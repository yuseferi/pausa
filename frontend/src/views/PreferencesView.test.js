import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

// PreferencesView reads its draft from the store. Mock it with a full default
// config so the sections render, and spy on the close action.
vi.mock('../lib/store', () => {
  const config = {
    version: 1,
    schedule: {
      shortInterval: '10m', shortDuration: '20s', longEvery: 3,
      longDuration: '5m', postponeShort: '2m', postponeLong: '5m',
    },
    notification: { enabled: true, warnShort: '10s', warnLong: '30s', playSound: false, showActions: true },
    display: {
      theme: 'system', fullscreen: true, allMonitors: true,
      showExerciseTips: true, showBreathing: true, accentColor: '#0ea5e9',
    },
    workingHours: {
      enabled: false, days: [false, true, true, true, true, true, false],
      startMinute: 540, endMinute: 1020,
    },
    idle: {
      pauseWhenIdle: true, idleThreshold: '2m', naturalBreaks: true,
      pauseWhenBusy: true, mediaCountsAsActivity: true,
    },
    general: { startAtLogin: false, showInDock: false, statusBarTitle: true },
  }
  return {
    state: { config },
    ui: { closePreferences: vi.fn() },
    saveConfig: vi.fn(),
    api: {
      getVersion: vi.fn().mockResolvedValue('1.0.0'),
      checkForUpdates: vi.fn().mockResolvedValue({}),
    },
  }
})

import PreferencesView from './PreferencesView.vue'
import { ui } from '../lib/store'

describe('PreferencesView', () => {
  beforeEach(() => ui.closePreferences.mockClear())

  it('exposes an accessible dialog', () => {
    const wrapper = mount(PreferencesView)
    const dialog = wrapper.find('[role="dialog"]')
    expect(dialog.exists()).toBe(true)
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('aria-labelledby')).toBe('prefs-title')
    expect(wrapper.find('#prefs-title').text()).toBe('Preferences')
  })

  it('closes on Escape', async () => {
    mount(PreferencesView)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(ui.closePreferences).toHaveBeenCalled()
  })

  it('marks the active sidebar tab with aria-current', () => {
    const wrapper = mount(PreferencesView)
    const tabs = wrapper.findAll('.modal__tab')
    expect(tabs.length).toBeGreaterThan(0)
    expect(tabs[0].attributes('aria-current')).toBe('true')
    expect(tabs[1].attributes('aria-current')).toBeUndefined()
  })
})
