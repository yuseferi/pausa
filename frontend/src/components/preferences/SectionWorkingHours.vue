<script setup>
import { computed } from 'vue'

const model = defineModel({ required: true })

const dayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const startTime = computed({
  get: () => minsToTime(model.value.startMinute),
  set: (v) => { model.value.startMinute = timeToMins(v) },
})
const endTime = computed({
  get: () => minsToTime(model.value.endMinute),
  set: (v) => { model.value.endMinute = timeToMins(v) },
})

function minsToTime(m) {
  const h = Math.floor(m / 60).toString().padStart(2, '0')
  const min = (m % 60).toString().padStart(2, '0')
  return `${h}:${min}`
}
function timeToMins(t) {
  const [h, m] = t.split(':').map(Number)
  return h * 60 + (m || 0)
}
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">
        Restrict to working hours
        <small class="field__hint">Breaks only fire during your work schedule.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.enabled" />
        <span class="toggle__slider" />
      </span>
    </label>

    <template v-if="model.enabled">
      <label class="field">
        <span class="field__label">Start</span>
        <input class="field__input" type="time" v-model="startTime" />
      </label>
      <label class="field">
        <span class="field__label">End</span>
        <input class="field__input" type="time" v-model="endTime" />
      </label>

      <div class="field">
        <span class="field__label">Active days</span>
        <div class="days">
          <button v-for="(label, i) in dayLabels" :key="i"
                  type="button"
                  class="day"
                  :class="{ 'day--on': model.days[i] }"
                  @click="model.days[i] = !model.days[i]">
            {{ label }}
          </button>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.days { display: flex; gap: 4px; flex-wrap: wrap; }
.day {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: transparent;
  font: inherit;
  font-size: 11px;
  font-weight: 600;
  color: var(--fg-muted);
  cursor: pointer;
}
.day--on {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}
</style>
