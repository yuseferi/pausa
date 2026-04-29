// Helpers for converting between Go-style duration strings and a
// {value, unit} pair editable in the UI. The backend accepts the string
// form ("10m", "20s", "1h30m") on save.

const UNITS = ['s', 'm', 'h']

// parse('10m') => 600 (seconds)
// parse('1h30m') => 5400
export function toSeconds(s) {
  if (typeof s !== 'string') return 0
  const re = /(\d+)(h|m|s)/g
  let total = 0, m
  while ((m = re.exec(s)) !== null) {
    const n = parseInt(m[1], 10)
    if (m[2] === 'h') total += n * 3600
    else if (m[2] === 'm') total += n * 60
    else total += n
  }
  return total
}

// fromSeconds(620, 'm') => '10m20s'
// Picks a sensible representation: drops zero parts.
export function fromSeconds(secs) {
  secs = Math.max(0, Math.floor(secs))
  if (secs === 0) return '0s'
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = secs % 60
  let out = ''
  if (h) out += `${h}h`
  if (m) out += `${m}m`
  if (s) out += `${s}s`
  return out
}

// toMinutes / fromMinutes — handy for fields that natively show minutes
export const toMinutes   = (s) => Math.round(toSeconds(s) / 60)
export const fromMinutes = (n) => fromSeconds(Math.max(0, n) * 60)
