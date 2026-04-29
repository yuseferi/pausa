<script setup>
import { computed } from 'vue'
import { toSeconds, fromSeconds, toMinutes, fromMinutes } from '../../lib/duration'

const model = defineModel({ required: true })

// Two-way computed bindings that translate string ↔ number for inputs.
const num = (path, parser, formatter) => computed({
  get: () => parser(model.value[path]),
  set: (v) => { model.value[path] = formatter(v) },
})

const shortIntervalMin = num('shortInterval', toMinutes, fromMinutes)
const shortDurationSec = num('shortDuration', toSeconds, fromSeconds)
const longDurationMin  = num('longDuration', toMinutes, fromMinutes)
const postponeShortMin = num('postponeShort', toMinutes, fromMinutes)
const postponeLongMin  = num('postponeLong', toMinutes, fromMinutes)
</script>

<template>
  <section class="section">
    <h3 class="section__title">Short breaks</h3>
    <label class="field">
      <span class="field__label">Interval (minutes)</span>
      <input class="field__input" type="number" min="1" max="240" v-model.number="shortIntervalMin" />
    </label>
    <label class="field">
      <span class="field__label">Duration (seconds)</span>
      <input class="field__input" type="number" min="5" max="600" v-model.number="shortDurationSec" />
    </label>
    <label class="field">
      <span class="field__label">Postpone (minutes)</span>
      <input class="field__input" type="number" min="1" max="60" v-model.number="postponeShortMin" />
    </label>

    <h3 class="section__title">Long breaks</h3>
    <label class="field">
      <span class="field__label">After every N short breaks</span>
      <input class="field__input" type="number" min="1" max="20" v-model.number="model.longEvery" />
    </label>
    <label class="field">
      <span class="field__label">Duration (minutes)</span>
      <input class="field__input" type="number" min="1" max="60" v-model.number="longDurationMin" />
    </label>
    <label class="field">
      <span class="field__label">Postpone (minutes)</span>
      <input class="field__input" type="number" min="1" max="120" v-model.number="postponeLongMin" />
    </label>
  </section>
</template>

<style scoped>
.section__title {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--fg-faint);
  margin: 12px 0 4px;
}
.section__title:first-child { margin-top: 0; }
</style>
