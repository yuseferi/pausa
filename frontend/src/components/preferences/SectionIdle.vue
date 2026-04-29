<script setup>
import { computed } from 'vue'
import { toSeconds, fromSeconds } from '../../lib/duration'

const model = defineModel({ required: true })

const thresholdSec = computed({
  get: () => toSeconds(model.value.idleThreshold),
  set: (v) => { model.value.idleThreshold = fromSeconds(v) },
})

const mediaDebounceSec = computed({
  get: () => toSeconds(model.value.busyMediaDebounce),
  set: (v) => { model.value.busyMediaDebounce = fromSeconds(v) },
})
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">
        Pause when idle
        <small class="field__hint">If you stop interacting with your Mac, breaks pause automatically.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.pauseWhenIdle" />
        <span class="toggle__slider" />
      </span>
    </label>

    <label v-if="model.pauseWhenIdle" class="field">
      <span class="field__label">Idle threshold (seconds)</span>
      <input class="field__input" type="number" min="30" max="3600" v-model.number="thresholdSec" />
    </label>

    <label class="field">
      <span class="field__label">
        Count natural breaks
        <small class="field__hint">If you've been idle for the upcoming break's duration, count it as taken.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.naturalBreaks" />
        <span class="toggle__slider" />
      </span>
    </label>

    <label class="field">
      <span class="field__label">
        Pause during meetings &amp; videos
        <small class="field__hint">Detects microphone use, system media playback, and supported browser video pages (YouTube, Meet, Netflix, Vimeo, Twitch, etc.). Resumes automatically.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.pauseWhenBusy" />
        <span class="toggle__slider" />
      </span>
    </label>

    <label v-if="model.pauseWhenBusy" class="field">
      <span class="field__label">
        Media debounce (seconds)
        <small class="field__hint">How long audio output must continue before counting as media. Browser video pages are immediate; this only affects the audio-output fallback.</small>
      </span>
      <input class="field__input" type="number" min="0" max="300" v-model.number="mediaDebounceSec" />
    </label>
  </section>
</template>
