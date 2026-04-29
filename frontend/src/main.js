import { createApp } from 'vue'
import App from './App.vue'
import { init } from './lib/store'
import './style.css'

createApp(App).mount('#app')

// Wails injects window.runtime + window.go before dispatching the
// `wails:ready` window event. If we missed it (race), kick off init
// immediately — the wailsjs bindings will queue calls until ready.
if (window.runtime) {
  init()
} else {
  window.addEventListener('wails:ready', init, { once: true })
  // Defensive fallback in case the event was missed entirely.
  setTimeout(() => { if (!window.__pausaInited) init() }, 500)
}
