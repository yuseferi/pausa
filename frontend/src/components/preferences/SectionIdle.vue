<script setup>
import { computed } from 'vue'
import { toSeconds, fromSeconds } from '../../lib/duration'

const model = defineModel({ required: true })

const thresholdSec = computed({
  get: () => toSeconds(model.value.idleThreshold),
  set: (v) => { model.value.idleThreshold = fromSeconds(v) },
})
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">
        Pause when idle
        <small class="field__hint">If you stop interacting with your Mac, breaks pause automatically and show as "Away". Media playing (meetings, videos, music) always shows as busy instead — it never counts as idle rest.</small>
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
        <small class="field__hint">If you've been away (no input, no media playing) for the upcoming break's duration, count it as taken. Movies, streams, and calls never count as natural breaks.</small>
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

    <label v-if="model.pauseWhenIdle" class="field">
      <span class="field__label">
        Media counts as activity
        <small class="field__hint">While meetings, videos, or music are playing, lack of keyboard/mouse input never triggers the "Away" pause. Whether media pauses the countdown as busy is still set by "Pause during meetings &amp; videos" above — turn that off to let breaks count down during movies.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.mediaCountsAsActivity" />
        <span class="toggle__slider" />
      </span>
    </label>
  </section>
</template>
