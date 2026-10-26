import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

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

import SectionIdle from './SectionIdle.vue'

const idle = () => ({
  pauseWhenIdle: true,
  idleThreshold: '2m',
  naturalBreaks: true,
  pauseWhenBusy: true,
  mediaCountsAsActivity: true,
})

const mountSection = () => mount(SectionIdle, { props: { modelValue: idle() } })

describe('SectionIdle', () => {
  beforeEach(() => {
    caps.idleDetection = true
    caps.busyDetection = true
  })

  it('shows both auto-pause groups on macOS', () => {
    const wrapper = mountSection()
    expect(wrapper.text()).toContain('Pause when idle')
    expect(wrapper.text()).toContain('Count natural breaks')
    expect(wrapper.text()).toContain('Pause during meetings')
    wrapper.unmount()
  })

  it('hides idle controls when idle time is not observable', () => {
    caps.idleDetection = false
    const wrapper = mountSection()
    expect(wrapper.text()).not.toContain('Pause when idle')
    expect(wrapper.text()).not.toContain('Count natural breaks')
    // Busy detection is independent of idle detection.
    expect(wrapper.text()).toContain('Pause during meetings')
    wrapper.unmount()
  })

  it('hides busy controls when media/mic detection is unavailable', () => {
    caps.busyDetection = false
    const wrapper = mountSection()
    expect(wrapper.text()).not.toContain('Pause during meetings')
    expect(wrapper.text()).not.toContain('Media counts as activity')
    expect(wrapper.text()).toContain('Pause when idle')
    wrapper.unmount()
  })

  it('hides everything when the platform can observe neither', () => {
    caps.idleDetection = false
    caps.busyDetection = false
    const wrapper = mountSection()
    expect(wrapper.text().trim()).toBe('')
    wrapper.unmount()
  })
})
