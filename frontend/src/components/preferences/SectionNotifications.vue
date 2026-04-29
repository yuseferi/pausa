<script setup>
import { computed } from 'vue'
import { toSeconds, fromSeconds } from '../../lib/duration'

const model = defineModel({ required: true })

const warnShortSec = computed({
  get: () => toSeconds(model.value.warnShort),
  set: (v) => { model.value.warnShort = fromSeconds(v) },
})
const warnLongSec = computed({
  get: () => toSeconds(model.value.warnLong),
  set: (v) => { model.value.warnLong = fromSeconds(v) },
})
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">
        Show notification before break
        <small class="field__hint">macOS will show a banner with Skip / Postpone buttons</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.enabled" />
        <span class="toggle__slider" />
      </span>
    </label>

    <template v-if="model.enabled">
      <label class="field">
        <span class="field__label">Warn before short break (seconds)</span>
        <input class="field__input" type="number" min="0" max="300" v-model.number="warnShortSec" />
      </label>
      <label class="field">
        <span class="field__label">Warn before long break (seconds)</span>
        <input class="field__input" type="number" min="0" max="600" v-model.number="warnLongSec" />
      </label>
      <label class="field">
        <span class="field__label">Show action buttons</span>
        <span class="toggle">
          <input type="checkbox" v-model="model.showActions" />
          <span class="toggle__slider" />
        </span>
      </label>
      <label class="field">
        <span class="field__label">Play sound</span>
        <span class="toggle">
          <input type="checkbox" v-model="model.playSound" />
          <span class="toggle__slider" />
        </span>
      </label>
    </template>
  </section>
</template>
