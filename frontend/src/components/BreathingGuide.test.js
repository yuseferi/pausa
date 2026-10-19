import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BreathingGuide from './BreathingGuide.vue'

describe('BreathingGuide', () => {
  it('renders the breathing prompt', () => {
    const wrapper = mount(BreathingGuide)
    expect(wrapper.find('.breathe__label').text()).toBe('Breathe')
    expect(wrapper.find('.breathe__circle').classes()).toContain('breathe__circle')
  })
})
