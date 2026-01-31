<template>
  <div class="break-overlay" :class="[breakType, { 'is-ending': isEnding }]">
    <!-- Animated background -->
    <div class="bg-gradient"></div>
    <div class="bg-particles">
      <div v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></div>
    </div>
    
    <!-- Main content -->
    <div class="break-content fade-in-up">
      <!-- Icon with animation -->
      <div class="icon-container">
        <div class="icon-glow"></div>
        <div class="icon animate-float">{{ breakIcon }}</div>
      </div>
      
      <!-- Break type badge -->
      <div class="break-badge slide-down stagger-1">
        <span class="badge-dot"></span>
        {{ breakType === 'mini' ? 'Mini Break' : 'Long Break' }}
      </div>
      
      <!-- Main heading -->
      <h1 class="break-title slide-up stagger-2">
        Time to Rest
      </h1>
      
      <!-- Tip message -->
      <div class="tip-container slide-up stagger-3">
        <div class="tip-icon">{{ tipIcon }}</div>
        <p class="tip-text">{{ tip.text || defaultTip }}</p>
      </div>
      
      <!-- Circular countdown timer -->
      <div class="countdown-container slide-up stagger-4">
        <svg class="countdown-svg" viewBox="0 0 120 120">
          <circle
            class="countdown-track"
            cx="60"
            cy="60"
            r="54"
            fill="none"
            stroke-width="6"
          />
          <circle
            class="countdown-progress"
            cx="60"
            cy="60"
            r="54"
            fill="none"
            stroke-width="6"
            :stroke-dasharray="circumference"
            :stroke-dashoffset="strokeDashoffset"
          />
        </svg>
        <div class="countdown-inner">
          <div class="countdown-time">{{ formatTime(timeLeft) }}</div>
          <div class="countdown-label">remaining</div>
        </div>
      </div>
      
      <!-- Progress bar (linear) -->
      <div class="progress-container slide-up stagger-5">
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
        </div>
      </div>
      
      <!-- Action buttons -->
      <div class="actions-container fade-in stagger-5">
        <button class="action-btn action-postpone" @click="handlePostpone">
          <span class="btn-icon">⏰</span>
          <span>Postpone</span>
        </button>
        <button class="action-btn action-skip" @click="handleSkip">
          <span class="btn-icon">⏭️</span>
          <span>Skip</span>
        </button>
      </div>
    </div>
    
    <!-- Breathing guide for long breaks -->
    <div v-if="breakType === 'long' && showBreathingGuide" class="breathing-guide">
      <div class="breathing-circle" :class="{ 'inhale': isInhaling, 'exhale': !isInhaling }">
        <span>{{ isInhaling ? 'Breathe In' : 'Breathe Out' }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { WindowSetAlwaysOnTop, WindowFullscreen, WindowMaximise, WindowShow, WindowSetSize, WindowSetPosition } from '../../wailsjs/runtime/runtime'

export default {
  name: 'BreakWindow',
  props: {
    breakType: { type: String, default: 'mini' },
    duration: { type: Number, default: 20 },
    tip: { type: Object, default: () => ({ text: '', category: '' }) },
    fullscreen: { type: Boolean, default: true }
  },
  emits: ['skip', 'postpone'],
  setup(props, { emit }) {
    const timeLeft = ref(props.duration)
    const isEnding = ref(false)
    const showBreathingGuide = ref(true)
    const isInhaling = ref(true)
    let timer
    let breathingTimer
    
    const circumference = 2 * Math.PI * 54

    const breakIcon = computed(() => {
      if (props.breakType === 'mini') {
        return '👁️'
      }
      return '🧘'
    })

    const tipIcon = computed(() => {
      const icons = {
        'eyes': '👀',
        'stretch': '🙆',
        'move': '🚶',
        'breathe': '🌬️'
      }
      return icons[props.tip.category] || '💡'
    })

    const defaultTip = computed(() => {
      if (props.breakType === 'mini') {
        return 'Look away from the screen and focus on something distant'
      }
      return 'Stand up, stretch, and take a short walk'
    })

    const progressPercent = computed(() => {
      return (timeLeft.value / props.duration) * 100
    })

    const strokeDashoffset = computed(() => {
      const progress = timeLeft.value / props.duration
      return circumference * (1 - progress)
    })

    const formatTime = (seconds) => {
      const mins = Math.floor(seconds / 60)
      const secs = seconds % 60
      if (mins > 0) {
        return `${mins}:${secs.toString().padStart(2, '0')}`
      }
      return secs.toString()
    }

    const getParticleStyle = (index) => {
      const size = 4 + Math.random() * 8
      return {
        '--size': `${size}px`,
        '--x': `${Math.random() * 100}%`,
        '--y': `${Math.random() * 100}%`,
        '--duration': `${15 + Math.random() * 20}s`,
        '--delay': `${Math.random() * 5}s`
      }
    }

    const handleSkip = () => {
      isEnding.value = true
      setTimeout(() => emit('skip'), 300)
    }

    const handlePostpone = () => {
      isEnding.value = true
      setTimeout(() => emit('postpone'), 300)
    }

    onMounted(async () => {
      try {
        await WindowSetAlwaysOnTop(true)
        await WindowShow()
      } catch (e) { console.log('Window setup error:', e) }
      
      if (props.fullscreen) {
        const docEl = document.documentElement
        if (docEl.requestFullscreen) {
          try { await docEl.requestFullscreen() } catch (e) {}
        } else if (docEl.webkitRequestFullscreen) {
          try { docEl.webkitRequestFullscreen() } catch (e) {}
        }
        
        setTimeout(async () => {
          try { await WindowFullscreen() } catch (e) {}
        }, 100)
        
        setTimeout(async () => {
          try {
            const screenWidth = window.screen.availWidth
            const screenHeight = window.screen.availHeight
            await WindowSetPosition(0, 0)
            await WindowSetSize(screenWidth, screenHeight)
          } catch (e) {}
        }, 200)
      } else {
        try { await WindowMaximise() } catch (e) {}
      }
      
      // Main countdown timer
      timer = setInterval(() => {
        if (timeLeft.value > 0) {
          timeLeft.value--
          if (timeLeft.value <= 3) {
            isEnding.value = true
          }
        } else {
          clearInterval(timer)
          if (window.go?.main?.App?.EndBreak) {
            window.go.main.App.EndBreak()
          }
        }
      }, 1000)

      // Breathing guide timer (4 seconds in, 4 seconds out)
      if (props.breakType === 'long') {
        breathingTimer = setInterval(() => {
          isInhaling.value = !isInhaling.value
        }, 4000)
      }
    })

    onUnmounted(() => {
      if (timer) clearInterval(timer)
      if (breathingTimer) clearInterval(breathingTimer)
    })

    return {
      timeLeft,
      isEnding,
      showBreathingGuide,
      isInhaling,
      circumference,
      breakIcon,
      tipIcon,
      defaultTip,
      progressPercent,
      strokeDashoffset,
      formatTime,
      getParticleStyle,
      handleSkip,
      handlePostpone
    }
  }
}
</script>

<style scoped>
.break-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  transition: opacity 0.3s ease;
}

.break-overlay.is-ending {
  opacity: 0.8;
}

/* Gradient background */
.bg-gradient {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #0891b2 0%, #0e7490 50%, #155e75 100%);
  animation: gradientShift 20s ease infinite;
  background-size: 200% 200%;
}

.break-overlay.long .bg-gradient {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 50%, #5b21b6 100%);
}

/* Floating particles */
.bg-particles {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.particle {
  position: absolute;
  width: var(--size);
  height: var(--size);
  background: rgba(255, 255, 255, 0.15);
  border-radius: 50%;
  left: var(--x);
  top: var(--y);
  animation: particleFloat var(--duration) ease-in-out infinite;
  animation-delay: var(--delay);
}

@keyframes particleFloat {
  0%, 100% {
    transform: translateY(0) translateX(0) scale(1);
    opacity: 0.3;
  }
  25% {
    transform: translateY(-30px) translateX(10px) scale(1.1);
    opacity: 0.5;
  }
  50% {
    transform: translateY(-20px) translateX(-10px) scale(0.9);
    opacity: 0.4;
  }
  75% {
    transform: translateY(-40px) translateX(5px) scale(1.05);
    opacity: 0.35;
  }
}

@keyframes gradientShift {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

/* Main content */
.break-content {
  position: relative;
  z-index: 10;
  text-align: center;
  color: white;
  padding: 40px;
  max-width: 500px;
}

/* Icon */
.icon-container {
  position: relative;
  display: inline-block;
  margin-bottom: 24px;
}

.icon {
  font-size: 80px;
  position: relative;
  z-index: 2;
}

.icon-glow {
  position: absolute;
  inset: -20px;
  background: radial-gradient(circle, rgba(255,255,255,0.3) 0%, transparent 70%);
  border-radius: 50%;
  animation: pulse 3s ease-in-out infinite;
}

/* Badge */
.break-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border-radius: 30px;
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 16px;
}

.badge-dot {
  width: 8px;
  height: 8px;
  background: #4ade80;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

/* Title */
.break-title {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: 20px;
  letter-spacing: -0.02em;
  text-shadow: 0 2px 20px rgba(0, 0, 0, 0.2);
}

/* Tip */
.tip-container {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  margin-bottom: 32px;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

.tip-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.tip-text {
  font-size: 16px;
  line-height: 1.5;
  opacity: 0.95;
  text-align: left;
}

/* Countdown */
.countdown-container {
  position: relative;
  width: 160px;
  height: 160px;
  margin: 0 auto 24px;
}

.countdown-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.countdown-track {
  stroke: rgba(255, 255, 255, 0.2);
}

.countdown-progress {
  stroke: white;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s linear;
  filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.5));
}

.countdown-inner {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.countdown-time {
  font-size: 42px;
  font-weight: 300;
  letter-spacing: -0.02em;
}

.countdown-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 2px;
  opacity: 0.7;
}

/* Linear progress */
.progress-container {
  width: 100%;
  max-width: 300px;
  margin: 0 auto 32px;
}

.progress-track {
  height: 6px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, rgba(255,255,255,0.8) 0%, white 100%);
  border-radius: 3px;
  transition: width 1s linear;
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
}

/* Actions */
.actions-container {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 28px;
  border: none;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.action-btn:active {
  transform: scale(0.98);
}

.action-postpone {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.action-postpone:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}

.action-skip {
  background: transparent;
  color: rgba(255, 255, 255, 0.7);
  border: 1px solid transparent;
}

.action-skip:hover {
  color: white;
  background: rgba(255, 255, 255, 0.1);
}

.btn-icon {
  font-size: 18px;
}

/* Breathing guide */
.breathing-guide {
  position: fixed;
  bottom: 60px;
  left: 50%;
  transform: translateX(-50%);
}

.breathing-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: white;
  transition: all 4s cubic-bezier(0.4, 0, 0.2, 1);
}

.breathing-circle.inhale {
  transform: scale(1.3);
  background: rgba(255, 255, 255, 0.2);
}

.breathing-circle.exhale {
  transform: scale(1);
  background: rgba(255, 255, 255, 0.1);
}

/* Animation classes */
.fade-in-up {
  animation: fadeInUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}

.slide-up {
  animation: slideUp 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}

.slide-down {
  animation: slideDown 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}

.fade-in {
  animation: fadeIn 0.4s ease forwards;
  opacity: 0;
}

.stagger-1 { animation-delay: 0.1s; }
.stagger-2 { animation-delay: 0.2s; }
.stagger-3 { animation-delay: 0.3s; }
.stagger-4 { animation-delay: 0.4s; }
.stagger-5 { animation-delay: 0.5s; }

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.05); }
}
</style>
