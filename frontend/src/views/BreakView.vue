<script setup>
// Full-screen break surface. Driven entirely by backend events:
// - duration & tip arrive in EventBreakStart
// - secondsLeft is updated by EventBreakTick once per second
// - EventBreakEnd / Skipped flips us back to DashboardView
//
// We never run our own countdown. The UI just renders what the store says.

import { computed } from 'vue'
import { state, accent, api } from '../lib/store'
import { useShortcuts } from '../composables/useShortcuts'
import BreakCountdown from '../components/BreakCountdown.vue'
import BreathingGuide from '../components/BreathingGuide.vue'
import ExerciseTip from '../components/ExerciseTip.vue'

const kind = computed(() => state.snapshot.currentKind || 'short')
const isLong = computed(() => kind.value === 'long')
const duration = computed(() => state.break.durationSeconds)
const remaining = computed(() => state.break.secondsLeft)

const showBreathing = computed(() => isLong.value && state.config?.display?.showBreathing)
const showTip = computed(() => state.config?.display?.showExerciseTips !== false)

const title = computed(() => isLong.value ? 'Long break' : 'Short break')
const heading = computed(() => isLong.value ? 'Time to recharge' : 'Rest your eyes')

useShortcuts({
  Escape:        () => api.skipBreak(),
  KeyP:          () => api.postponeBreak(),
  Space:         () => api.skipBreak(),
})
</script>

<template>
  <div class="break" :class="{ 'break--long': isLong }">
    <div class="break__bg">
      <div class="break__gradient" />
      <div class="break__particles">
        <span v-for="i in 18" :key="i" class="particle" :style="{
          '--x': `${(i * 53) % 100}%`,
          '--y': `${(i * 37) % 100}%`,
          '--d': `${20 + (i % 10) * 2}s`,
          '--delay': `${i * 0.4}s`,
        }" />
      </div>
    </div>

    <div class="break__content">
      <div class="break__badge">
        <span class="break__badge-dot" />
        {{ title }}
      </div>

      <h1 class="break__heading">{{ heading }}</h1>

      <ExerciseTip v-if="showTip" :tip="state.break.tip" />

      <BreakCountdown
        :duration="duration"
        :remaining="remaining"
      />

      <BreathingGuide v-if="showBreathing" />

      <div class="break__actions">
        <button class="btn btn--soft" @click="api.postponeBreak">
          <kbd>P</kbd> Postpone
        </button>
        <button class="btn btn--ghost" @click="api.skipBreak">
          <kbd>Esc</kbd> Skip
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.break {
  position: fixed;
  inset: 0;
  overflow: hidden;
  display: grid;
  place-items: center;
  color: white;
}

.break__bg { position: absolute; inset: 0; }
.break__gradient {
  position: absolute; inset: 0;
  background: linear-gradient(135deg,
    color-mix(in srgb, var(--accent) 70%, #0e7490),
    color-mix(in srgb, var(--accent) 40%, #155e75));
  background-size: 200% 200%;
  animation: shift 24s ease-in-out infinite;
}
.break--long .break__gradient {
  background: linear-gradient(135deg, #7c3aed, #6d28d9, #5b21b6);
  background-size: 200% 200%;
}
@keyframes shift {
  0%, 100% { background-position: 0% 50%; }
  50%      { background-position: 100% 50%; }
}

.break__particles { position: absolute; inset: 0; pointer-events: none; }
.particle {
  position: absolute;
  left: var(--x); top: var(--y);
  width: 6px; height: 6px;
  border-radius: 50%;
  background: rgba(255,255,255,0.18);
  animation: float-particle var(--d) ease-in-out var(--delay) infinite;
}
@keyframes float-particle {
  0%, 100% { transform: translate(0,0) scale(1); opacity: 0.3; }
  50%      { transform: translate(20px,-30px) scale(1.2); opacity: 0.6; }
}

.break__content {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: 40px;
  max-width: 560px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.break__badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: rgba(255,255,255,0.15);
  backdrop-filter: blur(10px);
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.break__badge-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: #4ade80;
  animation: pulse-soft 2s infinite;
}

.break__heading {
  font-size: 44px;
  font-weight: 300;
  letter-spacing: -0.02em;
  text-shadow: 0 2px 24px rgba(0,0,0,0.2);
}

.break__actions {
  display: flex;
  gap: 12px;
  margin-top: 12px;
}
.break__actions .btn {
  background: rgba(255,255,255,0.15);
  color: white;
  backdrop-filter: blur(10px);
}
.break__actions .btn:hover {
  background: rgba(255,255,255,0.25);
}
.break__actions .btn--ghost {
  background: transparent;
  color: rgba(255,255,255,0.8);
}
.break__actions .btn--ghost:hover {
  background: rgba(255,255,255,0.1);
}

kbd {
  font-family: -apple-system, monospace;
  font-size: 11px;
  padding: 1px 6px;
  background: rgba(255,255,255,0.2);
  border-radius: 4px;
  margin-right: 4px;
}
</style>
