<script setup>
// Circular countdown driven by `duration` and `remaining` props (both in
// seconds). The store updates `remaining` every second from backend ticks;
// CSS handles the smooth visual interpolation.

import { computed } from 'vue'
import { formatHMS } from '../composables/useCountdown'

const props = defineProps({
  duration:  { type: Number, required: true }, // total break length
  remaining: { type: Number, required: true }, // seconds left
})

const radius = 70
const circumference = 2 * Math.PI * radius

const dashoffset = computed(() => {
  if (props.duration <= 0) return circumference
  const ratio = Math.max(0, Math.min(1, props.remaining / props.duration))
  return circumference * (1 - ratio)
})

const display = computed(() => formatHMS(props.remaining))
</script>

<template>
  <div class="countdown">
    <svg class="countdown__svg" viewBox="0 0 160 160">
      <circle class="countdown__track" cx="80" cy="80" :r="radius"
              fill="none" stroke-width="6" />
      <circle class="countdown__fill"  cx="80" cy="80" :r="radius"
              fill="none" stroke-width="6"
              :stroke-dasharray="circumference"
              :stroke-dashoffset="dashoffset" />
    </svg>
    <div class="countdown__inner">
      <div class="countdown__time">{{ display }}</div>
      <div class="countdown__label">remaining</div>
    </div>
  </div>
</template>

<style scoped>
.countdown {
  position: relative;
  width: 180px; height: 180px;
}
.countdown__svg {
  width: 100%; height: 100%;
  transform: rotate(-90deg);
}
.countdown__track { stroke: rgba(255,255,255,0.18); }
.countdown__fill {
  stroke: white;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s linear;
  filter: drop-shadow(0 0 8px rgba(255,255,255,0.4));
}
.countdown__inner {
  position: absolute; inset: 0;
  display: grid; place-items: center;
  text-align: center;
}
.countdown__time {
  font-size: 40px;
  font-weight: 200;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}
.countdown__label {
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.7;
  margin-top: 2px;
}
</style>
