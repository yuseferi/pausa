<script setup>
// Pure-CSS breathing animation. 4 seconds in, 4 seconds out (8s loop).
// No JS timers; we rely on the keyframes for both the size and label
// transitions.
</script>

<template>
  <div class="breathe">
    <div class="breathe__circle">
      <span class="breathe__label">Breathe</span>
    </div>
  </div>
</template>

<style scoped>
.breathe {
  display: grid;
  place-items: center;
  margin-top: -8px;
}
.breathe__circle {
  width: 80px; height: 80px;
  border-radius: 50%;
  background: rgba(255,255,255,0.12);
  backdrop-filter: blur(10px);
  display: grid;
  place-items: center;
  animation: breathe 8s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}
.breathe__label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  opacity: 0.8;
}
@keyframes breathe {
  0%, 100% { transform: scale(1);   background: rgba(255,255,255,0.12); }
  50%      { transform: scale(1.4); background: rgba(255,255,255,0.22); }
}
</style>
