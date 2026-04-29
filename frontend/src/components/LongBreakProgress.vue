<script setup>
defineProps({
  completed: { type: Number, required: true },
  total:     { type: Number, required: true },
})
</script>

<template>
  <div class="progress">
    <div class="progress__label">
      {{ completed }} of {{ total }} short breaks until long break
    </div>
    <div class="progress__dots">
      <span
        v-for="i in total" :key="i"
        class="dot"
        :class="{ 'dot--filled': i <= completed }"
      />
    </div>
  </div>
</template>

<style scoped>
.progress { text-align: center; }
.progress__label {
  font-size: 12px;
  color: var(--fg-muted);
  margin-bottom: 8px;
}
.progress__dots {
  display: flex;
  justify-content: center;
  gap: 8px;
}
.dot {
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--border);
  transition: background 0.2s;
}
.dot--filled { background: var(--accent); }
</style>
