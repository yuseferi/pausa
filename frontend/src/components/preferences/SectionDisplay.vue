<script setup>
const model = defineModel({ required: true })

const themes = [
  { value: 'system', label: 'System' },
  { value: 'light',  label: 'Light' },
  { value: 'dark',   label: 'Dark' },
]

const accentPresets = ['#0ea5e9', '#7c3aed', '#22c55e', '#f59e0b', '#ef4444', '#ec4899']
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">Theme</span>
      <select class="field__input" style="width: auto;" v-model="model.theme">
        <option v-for="t in themes" :key="t.value" :value="t.value">{{ t.label }}</option>
      </select>
    </label>

    <div class="field">
      <span class="field__label">Accent color</span>
      <div class="swatches">
        <button v-for="c in accentPresets" :key="c"
                type="button"
                class="swatch"
                :class="{ 'swatch--active': model.accentColor === c }"
                :style="{ background: c }"
                @click="model.accentColor = c" />
      </div>
    </div>

    <label class="field">
      <span class="field__label">Fullscreen breaks</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.fullscreen" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">Show on all monitors</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.allMonitors" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">Show exercise tips</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.showExerciseTips" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">Breathing guide on long breaks</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.showBreathing" />
        <span class="toggle__slider" />
      </span>
    </label>
  </section>
</template>

<style scoped>
.swatches { display: flex; gap: 8px; }
.swatch {
  width: 24px; height: 24px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  transition: transform 0.1s ease;
}
.swatch:hover { transform: scale(1.1); }
.swatch--active {
  border-color: var(--fg);
  box-shadow: 0 0 0 2px var(--bg);
}
</style>
