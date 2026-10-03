<script setup>
import { onMounted, ref } from 'vue'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import { api } from '../../lib/store'

const model = defineModel({ required: true })

const version = ref('')
const checking = ref(false)
const result = ref(null) // {current, latest, available, url, error}

onMounted(async () => {
  try {
    const v = await api.getVersion()
    if (v) version.value = v
  } catch {
    // bindings not regenerated yet (run `wails generate module`)
  }
})

const check = async () => {
  checking.value = true
  result.value = null
  try {
    const info = await api.checkForUpdates()
    if (info) result.value = info
    else result.value = { error: 'Update check failed (no response).' }
  } catch (e) {
    result.value = { error: String(e?.message || e) }
  } finally {
    checking.value = false
  }
}

const openReleases = () => {
  const url = result.value?.url || 'https://github.com/yuseferi/pausa/releases/latest'
  try {
    BrowserOpenURL(url)
  } catch {
    window.open(url, '_blank', 'noopener')
  }
}
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">Start at login</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.startAtLogin" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">Show in Dock</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.showInDock" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">
        Show countdown in menu bar
        <small class="field__hint">Hides the timer text but keeps the icon.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.statusBarTitle" />
        <span class="toggle__slider" />
      </span>
    </label>

    <div class="field field--stack">
      <span class="field__label">
        App updates
        <small class="field__hint">
          Pausa never checks automatically and makes no network requests on its own.
          This button queries GitHub releases once, only when you click it.
        </small>
      </span>
      <div class="update-row">
        <span v-if="version" class="update-version">v{{ version }}</span>
        <button class="btn btn--soft" :disabled="checking" @click="check">
          {{ checking ? 'Checking…' : 'Check for updates' }}
        </button>
      </div>
      <div v-if="result" class="update-result">
        <template v-if="result.error">
          <span class="update-result__error">Check failed: {{ result.error }}</span>
        </template>
        <template v-else-if="result.available">
          <span>
            Update available: <strong>v{{ result.latest }}</strong>
            (you have v{{ result.current }}).
          </span>
          <button class="btn btn--primary btn--sm" @click="openReleases">
            Open releases
          </button>
        </template>
        <template v-else>
          <span>You're up to date (v{{ result.current }}).</span>
        </template>
      </div>
    </div>
  </section>
</template>

<style scoped>
.field--stack {
  display: grid;
  gap: 10px;
}
.update-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.update-version {
  font-size: 13px;
  color: var(--fg-muted);
  font-variant-numeric: tabular-nums;
}
.update-result {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--fg-muted);
  flex-wrap: wrap;
}
.update-result__error {
  color: color-mix(in srgb, red 65%, var(--fg));
}
.btn--sm {
  padding: 6px 12px;
  font-size: 13px;
}
</style>
