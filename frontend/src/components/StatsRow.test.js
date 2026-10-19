import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import StatsRow from './StatsRow.vue'

describe('StatsRow', () => {
  it('renders each stat value and label', () => {
    const wrapper = mount(StatsRow, {
      props: {
        stats: {
          breaksTaken: 3,
          naturalBreaks: 1,
          breaksPostponed: 2,
          breaksSkipped: 4,
        },
      },
    })

    expect(wrapper.findAll('.stat__value').map((n) => n.text())).toEqual(['3', '1', '2', '4'])
    expect(wrapper.findAll('.stat__label').map((n) => n.text())).toEqual([
      'Taken',
      'Natural',
      'Postponed',
      'Skipped',
    ])
  })
})
