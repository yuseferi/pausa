import { describe, it, expect } from 'vitest'
import { toSeconds, fromSeconds, toMinutes, fromMinutes } from './duration'

describe('toSeconds', () => {
  it('parses simple durations', () => {
    expect(toSeconds('10m')).toBe(600)
    expect(toSeconds('20s')).toBe(20)
    expect(toSeconds('1h')).toBe(3600)
  })

  it('parses composite and fractional values (as Go emits them)', () => {
    expect(toSeconds('10m20s')).toBe(620)
    expect(toSeconds('1h30m')).toBe(5400)
    expect(toSeconds('1.5s')).toBe(1.5)
    expect(toSeconds('500ms')).toBe(0.5)
  })

  it('returns 0 for non-strings, empty, and unparseable input', () => {
    expect(toSeconds(null)).toBe(0)
    expect(toSeconds(undefined)).toBe(0)
    expect(toSeconds('')).toBe(0)
    expect(toSeconds('nonsense')).toBe(0)
  })
})

describe('fromSeconds', () => {
  it('drops zero parts', () => {
    expect(fromSeconds(0)).toBe('0s')
    expect(fromSeconds(600)).toBe('10m')
    expect(fromSeconds(620)).toBe('10m20s')
    expect(fromSeconds(5400)).toBe('1h30m')
  })

  it('floors fractional seconds and clamps negatives', () => {
    expect(fromSeconds(5.9)).toBe('5s')
    expect(fromSeconds(-10)).toBe('0s')
  })
})

describe('minutes helpers', () => {
  it('round-trips through minutes', () => {
    expect(toMinutes('10m')).toBe(10)
    expect(fromMinutes(10)).toBe('10m')
  })

  it('rounds to the nearest minute', () => {
    expect(toMinutes('90s')).toBe(2)
    expect(toMinutes('89s')).toBe(1)
  })
})
