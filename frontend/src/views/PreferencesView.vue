<script setup>
// Modal preferences. Local mutations are made on a draft; on save, we send
// to the backend, which validates and broadcasts a `config:updated` event
// the store listens to.
//
// Settings are organized into focused sections, each its own component, so
// adding a new section is a one-file change.

import { reactive, ref, watch } from 'vue'
import { state, ui, saveConfig } from '../lib/store'
import SectionSchedule from '../components/preferences/SectionSchedule.vue'
import SectionNotifications from '../components/preferences/SectionNotifications.vue'
import SectionDisplay from '../components/preferences/SectionDisplay.vue'
import SectionWorkingHours from '../components/preferences/SectionWorkingHours.vue'
import SectionIdle from '../components/preferences/SectionIdle.vue'
import SectionGeneral from '../components/preferences/SectionGeneral.vue'

const tabs = [
  { id: 'schedule',     label: 'Schedule' },
  { id: 'notifications', label: 'Notifications' },
  { id: 'display',      label: 'Display' },
  { id: 'workingHours', label: 'Working Hours' },
  { id: 'idle',         label: 'Idle & Natural' },
  { id: 'general',      label: 'General' },
]
const tab = ref('schedule')

// Deep-clone the live config so cancellations don't mutate it.
const draft = reactive(JSON.parse(JSON.stringify(state.config || {})))

watch(() => state.config, (cfg) => {
  if (!cfg) return
  Object.assign(draft, JSON.parse(JSON.stringify(cfg)))
}, { deep: true })

const onSave = async () => {
  await saveConfig(JSON.parse(JSON.stringify(draft)))
  ui.closePreferences()
}
</script>

<template>
  <div class="modal-overlay" @click.self="ui.closePreferences">
    <div class="modal" role="dialog" aria-modal="true">
      <header class="modal__header">
        <h2 class="modal__title">Preferences</h2>
        <button class="btn btn--ghost btn--icon" @click="ui.closePreferences">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </header>

      <nav class="modal__tabs">
        <button v-for="t in tabs" :key="t.id"
                :class="['modal__tab', { 'modal__tab--active': tab === t.id }]"
                @click="tab = t.id">
          {{ t.label }}
        </button>
      </nav>

      <div class="modal__body">
        <SectionSchedule       v-if="tab === 'schedule'"     v-model="draft.schedule" />
        <SectionNotifications  v-else-if="tab === 'notifications'" v-model="draft.notification" />
        <SectionDisplay        v-else-if="tab === 'display'" v-model="draft.display" />
        <SectionWorkingHours   v-else-if="tab === 'workingHours'" v-model="draft.workingHours" />
        <SectionIdle           v-else-if="tab === 'idle'"    v-model="draft.idle" />
        <SectionGeneral        v-else-if="tab === 'general'" v-model="draft.general" />
      </div>

      <footer class="modal__footer">
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
  padding: 20px;
}
.modal {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
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

.modal__tabs {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
}
.modal__tab {
  background: transparent;
  border: 0;
  padding: 6px 12px;
  font: inherit;
  font-size: 13px;
  color: var(--fg-muted);
  border-radius: 8px;
  cursor: pointer;
  white-space: nowrap;
}
.modal__tab:hover { background: var(--bg-deep); color: var(--fg); }
.modal__tab--active {
  background: var(--accent-soft);
  color: var(--accent);
}

.modal__body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 14px 18px;
  border-top: 1px solid var(--border);
}
</style>
