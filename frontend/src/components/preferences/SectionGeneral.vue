<script setup>
const model = defineModel({ required: true })
</script>

<template>
  <section>
    <label class="field">
      <span class="field__label">Start at login</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.startAtLogin" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">Show in Dock</span>
      <span class="toggle">
        <input type="checkbox" v-model="model.showInDock" />
        <span class="toggle__slider" />
      </span>
    </label>
    <label class="field">
      <span class="field__label">
        Show countdown in menu bar
        <small class="field__hint">Hides the timer text but keeps the icon.</small>
      </span>
      <span class="toggle">
        <input type="checkbox" v-model="model.statusBarTitle" />
        <span class="toggle__slider" />
      </span>
    </label>
  </section>
</template>
