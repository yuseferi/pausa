// Register keyboard shortcuts for the duration of a component's lifetime.
// Pass a map of `'KeyCode'` (e.g. 'Escape', 'Space', 'KeyB') to handler.
// Modifier keys can be combined with '+': 'Meta+B', 'Shift+Escape'.

import { onMounted, onUnmounted } from 'vue'

export function useShortcuts(map) {
  const handler = (e) => {
    // Ignore when typing in form fields
    const t = e.target
    if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) {
      return
    }
    const parts = []
    if (e.metaKey)  parts.push('Meta')
    if (e.ctrlKey)  parts.push('Ctrl')
    if (e.shiftKey) parts.push('Shift')
    if (e.altKey)   parts.push('Alt')
    parts.push(e.code)
    const combo = parts.join('+')
    const fn = map[combo] || map[e.code]
    if (typeof fn === 'function') {
      e.preventDefault()
      fn(e)
    }
  }
  onMounted(() => window.addEventListener('keydown', handler))
  onUnmounted(() => window.removeEventListener('keydown', handler))
}
