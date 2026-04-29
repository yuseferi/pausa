<script setup>
import { computed } from 'vue'

const props = defineProps({
  tip: { type: Object, required: true },
})

const icon = computed(() => {
  const map = { eyes: '👀', stretch: '🙆', move: '🚶', breathe: '🌬️', posture: '🧍' }
  return map[props.tip?.category] || '💡'
})

const fallback = 'Take a moment to relax.'
</script>

<template>
  <div v-if="tip?.text" class="tip">
    <span class="tip__icon">{{ icon }}</span>
    <p class="tip__text">{{ tip.text || fallback }}</p>
  </div>
</template>

<style scoped>
.tip {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  background: rgba(255,255,255,0.12);
  backdrop-filter: blur(10px);
  border-radius: 14px;
  max-width: 440px;
}
.tip__icon { font-size: 22px; flex-shrink: 0; }
.tip__text {
  font-size: 15px;
  line-height: 1.4;
  color: rgba(255,255,255,0.95);
  text-align: left;
}
</style>
