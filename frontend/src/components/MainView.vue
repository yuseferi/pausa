<template>
  <div class="main-view">
    <div class="header">
      <h1>☕ Pausa</h1>
      <button class="btn-icon" @click="$emit('openPreferences')" title="Preferences">⚙️</button>
    </div>

    <div class="status-card card">
      <div class="status-icon">{{ state.isPaused ? '⏸️' : '⏱️' }}</div>
      <div class="status-info">
        <div class="status-label">{{ state.isPaused ? 'Breaks paused' : 'Next break in' }}</div>
        <div class="status-time" v-if="!state.isPaused">{{ formatTimeUntil(state.breakTimeLeft) }}</div>
        <div class="status-type">{{ state.nextBreakType === 'mini' ? 'Mini break' : 'Long break' }}</div>
      </div>
    </div>

    <div class="progress-info">
      <span>{{ state.miniBreakCount }} / {{ config.longBreakInterval }} mini breaks until long break</span>
      <div class="progress-dots">
        <span v-for="i in config.longBreakInterval" :key="i" 
              class="dot" :class="{ filled: i <= state.miniBreakCount }"></span>
      </div>
    </div>

    <div class="actions">
      <button v-if="state.isPaused" class="btn btn-primary" @click="$emit('resumeBreaks')">
        ▶️ Resume Breaks
      </button>
      <button v-else class="btn btn-secondary" @click="$emit('pauseBreaks')">
        ⏸️ Pause Breaks
      </button>
      <button class="btn btn-secondary" @click="$emit('resetBreaks')">
        🔄 Reset
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MainView',
  props: {
    state: { type: Object, required: true },
    config: { type: Object, required: true }
  },
  emits: ['openPreferences', 'pauseBreaks', 'resumeBreaks', 'resetBreaks'],
  setup() {
    const formatTimeUntil = (seconds) => {
      if (!seconds || seconds < 0) return '--:--'
      const mins = Math.floor(seconds / 60)
      const secs = seconds % 60
      return `${mins}:${secs.toString().padStart(2, '0')}`
    }
    return { formatTimeUntil }
  }
}
</script>

<style scoped>
.main-view {
  height: 100%;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h1 {
  font-size: 20px;
  font-weight: 600;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
}

.status-card {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.status-icon {
  font-size: 40px;
}

.status-info {
  flex: 1;
}

.status-label {
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: uppercase;
}

.status-time {
  font-size: 32px;
  font-weight: 300;
}

.status-type {
  font-size: 14px;
  color: var(--text-secondary);
}

.progress-info {
  text-align: center;
  margin-bottom: 20px;
  font-size: 13px;
  color: var(--text-secondary);
}

.progress-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 8px;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--border);
}

.dot.filled {
  background: var(--accent);
}

.actions {
  margin-top: auto;
  display: flex;
  gap: 10px;
}

.actions .btn {
  flex: 1;
}
</style>