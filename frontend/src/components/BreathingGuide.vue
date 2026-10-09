<script setup>
// Pure-CSS breathing animation. 4 seconds in, 4 seconds out (8s loop).
// No JS timers; we rely on the keyframes for both the size and label
// transitions.
</script>

<template>
  <div class="breathe">
    <div class="breathe__circle">
      <div class="breathe__inner">
        <span class="breathe__phase"></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.breathe {
  display: grid;
  place-items: center;
  margin-top: -4px;
}
.breathe__circle {
  width: 90px; height: 90px;
  border-radius: 50%;
  background: rgba(255,255,255,0.14);
  backdrop-filter: blur(10px);
  display: grid;
  place-items: center;
  animation: breathe 8s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  box-shadow: 0 0 20px rgba(255,255,255,0.2);
}
.breathe__inner {
  text-align: center;
}
.breathe__phase {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  opacity: 0.9;
}
.breathe__phase::after {
  content: 'Inhale';
  animation: phaseText 8s infinite;
}
@keyframes breathe {
  0%, 100% { transform: scale(1); background: rgba(255,255,255,0.12); }
  50%      { transform: scale(1.45); background: rgba(255,255,255,0.28); }
}
@keyframes phaseText {
  0%, 45%   { content: 'Inhale'; }
  50%, 95%  { content: 'Exhale'; }
}
</style>
