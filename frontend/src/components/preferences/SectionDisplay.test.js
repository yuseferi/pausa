import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

// The section reads capabilities from the store to decide which rows exist.
// vi.hoisted keeps the object reachable from the mock factory, which vitest
// lifts above the imports.
const { caps } = vi.hoisted(() => ({
  caps: {
    nativeOverlays: true,
    systemTray: true,
    dockIcon: true,
    loginItem: true,
    busyDetection: true,
    idleDetection: true,
  },
}))

vi.mock('../../lib/store', () => ({ state: { capabilities: caps } }))

import SectionDisplay from './SectionDisplay.vue'

const display = () => ({
  theme: 'system',
  fullscreen: true,
  allMonitors: true,
  showExerciseTips: true,
  showBreathing: true,
  accentColor: '#0ea5e9',
})

const mountSection = () => mount(SectionDisplay, { props: { modelValue: display() } })

describe('SectionDisplay', () => {
  beforeEach(() => { caps.nativeOverlays = true })

  it('offers "show on all monitors" when the platform has native overlays', () => {
    const wrapper = mountSection()
    expect(wrapper.text()).toContain('Show on all monitors')
    wrapper.unmount()
  })

  it('hides "show on all monitors" on platforms without native overlays', () => {
    caps.nativeOverlays = false
    const wrapper = mountSection()
    // There is only one break surface there — the app window — so the option
    // would be a no-op.
    expect(wrapper.text()).not.toContain('Show on all monitors')
    wrapper.unmount()
  })

  it('keeps the settings that work everywhere', () => {
    caps.nativeOverlays = false
    const wrapper = mountSection()
    expect(wrapper.text()).toContain('Fullscreen breaks')
    expect(wrapper.text()).toContain('Accent color')
    wrapper.unmount()
  })
})
