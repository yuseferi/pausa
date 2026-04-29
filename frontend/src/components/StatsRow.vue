<script setup>
defineProps({ stats: { type: Object, required: true } })
</script>

<template>
  <div class="stats">
    <div class="stat">
      <div class="stat__value">{{ stats.breaksTaken }}</div>
      <div class="stat__label">Taken</div>
    </div>
    <div class="stat">
      <div class="stat__value">{{ stats.naturalBreaks }}</div>
      <div class="stat__label">Natural</div>
    </div>
    <div class="stat">
      <div class="stat__value">{{ stats.breaksPostponed }}</div>
      <div class="stat__label">Postponed</div>
    </div>
    <div class="stat stat--muted">
      <div class="stat__value">{{ stats.breaksSkipped }}</div>
      <div class="stat__label">Skipped</div>
    </div>
  </div>
</template>

<style scoped>
.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--border);
}
.stat { text-align: center; padding: 8px 4px; }
.stat__value {
  font-size: 22px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--fg);
}
.stat__label {
  font-size: 10px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--fg-faint);
  margin-top: 2px;
}
.stat--muted .stat__value { color: var(--fg-faint); }
</style>
