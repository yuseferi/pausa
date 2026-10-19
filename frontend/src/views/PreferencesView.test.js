import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

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

// The section components are separate units with their own concerns, and
// SectionGeneral imports the generated Wails bindings from the gitignored
// wailsjs/ directory. Stub them so this test covers only the dialog chrome and
// needs no `wails generate module` step.
vi.mock('../components/preferences/SectionSchedule.vue', () => ({ default: { name: 'SectionSchedule', template: '<div />' } }))
vi.mock('../components/preferences/SectionWorkingHours.vue', () => ({ default: { name: 'SectionWorkingHours', template: '<div />' } }))
vi.mock('../components/preferences/SectionIdle.vue', () => ({ default: { name: 'SectionIdle', template: '<div />' } }))
vi.mock('../components/preferences/SectionNotifications.vue', () => ({ default: { name: 'SectionNotifications', template: '<div />' } }))
vi.mock('../components/preferences/SectionDisplay.vue', () => ({ default: { name: 'SectionDisplay', template: '<div />' } }))
vi.mock('../components/preferences/SectionGeneral.vue', () => ({ default: { name: 'SectionGeneral', template: '<div />' } }))

import PreferencesView from './PreferencesView.vue'
import { ui } from '../lib/store'

const mountDialog = (options = {}) => mount(PreferencesView, { attachTo: document.body, ...options })

describe('PreferencesView', () => {
  beforeEach(() => ui.closePreferences.mockClear())

  it('exposes an accessible dialog', () => {
    const wrapper = mountDialog()
    const dialog = wrapper.find('[role="dialog"]')
    expect(dialog.exists()).toBe(true)
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('aria-labelledby')).toBe('prefs-title')
    expect(wrapper.find('#prefs-title').text()).toBe('Preferences')
    wrapper.unmount()
  })

  it('closes on Escape', () => {
    const wrapper = mountDialog()
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(ui.closePreferences).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('marks the active sidebar tab with aria-current', () => {
    const wrapper = mountDialog()
    const tabs = wrapper.findAll('.modal__tab')
    expect(tabs.length).toBeGreaterThan(0)
    expect(tabs[0].attributes('aria-current')).toBe('true')
    expect(tabs[1].attributes('aria-current')).toBeUndefined()
    wrapper.unmount()
  })

  it('moves focus into the dialog on open', async () => {
    const wrapper = mountDialog()
    await nextTick()
    expect(document.activeElement).toBe(wrapper.find('[role="dialog"]').element)
    wrapper.unmount()
  })

  it('restores focus to the previously focused element on close', async () => {
    const trigger = document.createElement('button')
    document.body.appendChild(trigger)
    trigger.focus()

    const wrapper = mountDialog()
    await nextTick()
    expect(document.activeElement).not.toBe(trigger)

    wrapper.unmount()
    expect(document.activeElement).toBe(trigger)
    trigger.remove()
  })
})
