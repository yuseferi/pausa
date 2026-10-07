<script setup>
// Sidebar preferences. Local mutations are made on a draft; on save, we send
// to the backend, which validates and broadcasts a `config:updated` event
// the store listens to.

import { computed, reactive, ref, watch } from 'vue'
import { state, ui, saveConfig } from '../lib/store'
import SectionSchedule from '../components/preferences/SectionSchedule.vue'
import SectionNotifications from '../components/preferences/SectionNotifications.vue'
import SectionDisplay from '../components/preferences/SectionDisplay.vue'
import SectionWorkingHours from '../components/preferences/SectionWorkingHours.vue'
import SectionIdle from '../components/preferences/SectionIdle.vue'
import SectionGeneral from '../components/preferences/SectionGeneral.vue'

const tabs = [
  { id: 'breaks',     label: 'Breaks' },
  { id: 'detection',  label: 'Detection' },
  { id: 'experience', label: 'Experience' },
  { id: 'general',    label: 'General' },
]
const tab = ref('breaks')

// Deep-clone the live config so cancellations don't mutate it. The modal is
// gated on `ready` because the native menu can open preferences before the
// initial config has loaded; rendering sections with an empty draft would
// dereference undefined paths and throw.
const ready = computed(() => !!state.config)
const draft = reactive(state.config ? JSON.parse(JSON.stringify(state.config)) : {})

// Don't clobber in-progress edits when `config:updated` arrives while open.
// Only re-sync from the store when the draft still matches the last synced
// snapshot (i.e. the user has no unsaved edits).
const lastSynced = ref(ready.value ? JSON.stringify(draft) : '')

watch(() => state.config, (cfg) => {
  if (!cfg) return
  if (ready.value && JSON.stringify(draft) !== lastSynced.value) return
  Object.assign(draft, JSON.parse(JSON.stringify(cfg)))
  lastSynced.value = JSON.stringify(draft)
}, { deep: true })

const resetDraft = () => {
  if (!state.config) return
  Object.assign(draft, JSON.parse(JSON.stringify(state.config)))
  lastSynced.value = JSON.stringify(draft)
}

const onSave = async () => {
  await saveConfig(JSON.parse(JSON.stringify(draft)))
  lastSynced.value = JSON.stringify(draft)
  ui.closePreferences()
}
</script>

<template>
  <div v-if="ready" class="modal-overlay" @click.self="ui.closePreferences">
    <div class="modal" role="dialog" aria-modal="true">
      <header class="modal__header">
        <h2 class="modal__title">Preferences</h2>
        <button class="btn btn--ghost btn--icon" @click="ui.closePreferences">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </header>

      <div class="prefs">
        <nav class="prefs__side">
          <button v-for="t in tabs" :key="t.id"
                  :class="['modal__tab', { 'modal__tab--active': tab === t.id }]"
                  @click="tab = t.id">
            {{ t.label }}
          </button>
        </nav>

        <div class="prefs__content">
          <template v-if="tab === 'breaks'">
            <div class="pref-group">
              <h3 class="section__title">Break schedule</h3>
              <SectionSchedule v-model="draft.schedule" />
            </div>
            <div class="pref-group">
              <h3 class="section__title">Working hours</h3>
              <SectionWorkingHours v-model="draft.workingHours" />
            </div>
          </template>

          <template v-else-if="tab === 'detection'">
            <div class="pref-group">
              <h3 class="section__title">Idle &amp; natural breaks</h3>
              <SectionIdle v-model="draft.idle" />
            </div>
          </template>

          <template v-else-if="tab === 'experience'">
            <div class="pref-group">
              <h3 class="section__title">Notifications</h3>
              <SectionNotifications v-model="draft.notification" />
            </div>
            <div class="pref-group">
              <h3 class="section__title">Display</h3>
              <SectionDisplay v-model="draft.display" />
            </div>
          </template>

          <template v-else-if="tab === 'general'">
            <div class="pref-group">
              <h3 class="section__title">General</h3>
              <SectionGeneral v-model="draft.general" />
            </div>
          </template>
        </div>
      </div>

      <footer class="modal__footer">
        <button class="btn btn--ghost" @click="resetDraft">Reset</button>
        <span class="modal__spacer" />
        <button class="btn btn--ghost" @click="ui.closePreferences">Cancel</button>
        <button class="btn btn--primary" @click="onSave">Save</button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
  display: grid; place-items: center;
  z-index: 100;
  padding: 12px;
}
.modal {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  width: 100%;
  max-width: 640px;
  max-height: 92vh;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border);
}
.modal__title { font-size: 16px; font-weight: 600; }
.btn--icon { padding: 6px; }

.prefs {
  display: flex;
  flex: 1;
  min-height: 0;
}
.prefs__side {
  width: 180px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.prefs__content {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding: 20px 24px;
}

.modal__tab {
  background: transparent;
  border: 0;
  padding: 8px 12px;
  font: inherit;
  font-size: 13px;
  color: var(--fg-muted);
  border-radius: 8px;
  cursor: pointer;
  white-space: nowrap;
  text-align: left;
}
.modal__tab:hover { background: var(--bg-deep); color: var(--fg); }
.modal__tab--active {
  background: var(--accent-soft);
  color: var(--accent);
}

.modal__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 18px;
  border-top: 1px solid var(--border);
}
.modal__spacer { flex: 1; }

@media (max-width: 560px) {
  .prefs__side { width: 140px; padding: 8px; }
  .prefs__content { padding: 16px; }
}
</style>
