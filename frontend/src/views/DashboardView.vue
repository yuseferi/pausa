<script setup>
// Primary surface when no break is active. Shows the next-break countdown,
// progress toward the next long break, today's stats, and primary controls.

import { computed, toRef } from 'vue'
import { state, isPaused, isAutoPaused, isAnyPaused, autoPauseReason, ui, api } from '../lib/store'
import { useCountdown, formatHMS } from '../composables/useCountdown'
import StatsRow from '../components/StatsRow.vue'
import LongBreakProgress from '../components/LongBreakProgress.vue'

const target = toRef(() => state.snapshot.nextBreakAt)
const seconds = useCountdown(target)

const nextKindLabel = computed(() =>
  state.snapshot.nextKind === 'long' ? 'Long break' : 'Short break'
)

const statusLabel = computed(() => {
  if (isAutoPaused.value) return 'Paused — ' + (autoPauseReason.value || 'busy')
  if (isPaused.value) return 'Paused'
  return 'Up next'
})
</script>

<template>
  <div class="dashboard">
    <div class="titlebar-drag" />

    <header class="dashboard__header">
      <div class="brand">
        <span class="brand__dot" :class="{ 'brand__dot--paused': isAnyPaused }" />
        <span class="brand__text">Pausa</span>
      </div>
      <button class="btn btn--ghost btn--icon" @click="ui.openPreferences" aria-label="Preferences">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h0a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h0a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v0a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
        </svg>
      </button>
    </header>

    <section class="next-break" :class="{ 'next-break--paused': isAnyPaused }">
      <div class="next-break__label">{{ statusLabel }}</div>
      <div class="next-break__time">
        <template v-if="isAnyPaused">—</template>
        <template v-else>{{ formatHMS(seconds) }}</template>
      </div>
      <div class="next-break__sub">{{ nextKindLabel }}</div>
    </section>

    <LongBreakProgress
      :completed="state.snapshot.shortsCompleted"
      :total="state.config?.schedule?.longEvery || 3"
    />

    <div v-if="isAutoPaused" class="busy-note card">
      <div class="busy-note__title">Auto-paused</div>
      <div class="busy-note__text">
        Break timer is paused because Pausa detected
        <strong>{{ autoPauseReason || 'busy activity' }}</strong>.
        The countdown will resume automatically when you're free.
      </div>
    </div>

    <div class="actions">
      <button class="btn btn--primary" @click="api.takeBreakNow">
        Take a break now
      </button>
      <button v-if="isAnyPaused" class="btn btn--soft" @click="api.resumeBreaks">
        Resume
      </button>
      <button v-else class="btn btn--soft" @click="api.pauseBreaks">
        Pause
      </button>
    </div>

    <StatsRow :stats="state.snapshot.stats" />
  </div>
</template>

<style scoped>
.dashboard {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0 24px 24px;
  gap: 18px;
  background: var(--bg);
}

.dashboard__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  margin-bottom: 4px;
}
.brand { display: flex; align-items: center; gap: 8px; }
.brand__dot {
  width: 10px; height: 10px; border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--success) 25%, transparent);
}
.brand__dot--paused {
  background: var(--warning);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--warning) 25%, transparent);
}
.brand__text { font-size: 14px; font-weight: 600; color: var(--fg-muted); }

.btn--icon { padding: 8px; }

.next-break {
  text-align: center;
  padding: 28px 16px;
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}
.next-break__label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--fg-faint);
}
.next-break__time {
  font-size: 56px;
  font-weight: 200;
  font-variant-numeric: tabular-nums;
  margin: 8px 0 4px;
  color: var(--fg);
  letter-spacing: -0.03em;
}
.next-break__sub {
  font-size: 14px;
  color: var(--fg-muted);
}
.next-break--paused .next-break__time { color: var(--fg-faint); }

.actions {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 10px;
}

.busy-note {
  padding: 14px 16px;
  background: color-mix(in srgb, var(--warning) 10%, var(--bg-elev));
  border-color: color-mix(in srgb, var(--warning) 25%, var(--border));
}
.busy-note__title {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--warning) 70%, var(--fg));
  margin-bottom: 4px;
}
.busy-note__text {
  font-size: 13px;
  color: var(--fg-muted);
}
</style>
