<script setup>
// Root view router. Picks one of three primary surfaces based on store
// state: WelcomeView (first run), BreakView (during break), DashboardView
// (idle / scheduled / paused).
//
// PreferencesView is a modal overlay shown on top of any other surface.

import { computed } from 'vue'
import { state, onBreak } from './lib/store'
import WelcomeView from './views/WelcomeView.vue'
import DashboardView from './views/DashboardView.vue'
import BreakView from './views/BreakView.vue'
import PreferencesView from './views/PreferencesView.vue'

const view = computed(() => {
  if (!state.ready)        return 'loading'
  if (state.firstLaunch)   return 'welcome'
  if (onBreak.value)       return 'break'
  return 'dashboard'
})
</script>

<template>
  <div class="app-root">
    <Transition name="fade" mode="out-in">
      <WelcomeView    v-if="view === 'welcome'"    key="welcome" />
      <BreakView      v-else-if="view === 'break'" key="break" />
      <DashboardView  v-else-if="view === 'dashboard'" key="dashboard" />
      <div v-else class="loading" key="loading">
        <div class="titlebar-drag" />
        <div class="loading__inner">
          <div class="spinner" />
          <div class="loading__text">Pausa</div>
        </div>
      </div>
    </Transition>

    <PreferencesView v-if="state.ui.preferencesOpen" />
  </div>
</template>

<style scoped>
.app-root { height: 100%; width: 100%; }
.loading {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg);
}
.loading__inner {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
}
.loading__text {
  font-size: 13px;
  color: var(--fg-faint);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}
.spinner {
  width: 32px; height: 32px;
  border-radius: 50%;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
