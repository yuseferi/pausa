<template>
  <div class="welcome-screen fade-in">
    <div class="welcome-content slide-up">
      <div class="logo">☕</div>
      <h1>Welcome to Pausa</h1>
      <p class="subtitle">Your break time reminder for a healthier workday</p>
      
      <div class="features">
        <div class="feature">
          <span class="icon">👁️</span>
          <span>Rest your eyes regularly</span>
        </div>
        <div class="feature">
          <span class="icon">🧘</span>
          <span>Stretch and move your body</span>
        </div>
        <div class="feature">
          <span class="icon">💪</span>
          <span>Stay productive and healthy</span>
        </div>
      </div>

      <button class="btn btn-primary btn-large" @click="$emit('complete')">
        Get Started
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'WelcomeScreen',
  emits: ['complete']
}
</script>

<style scoped>
.welcome-screen {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--accent) 0%, #2c5364 100%);
  color: white;
  padding: 20px;
}

.welcome-content {
  text-align: center;
  max-width: 400px;
}

.logo {
  font-size: 64px;
  margin-bottom: 16px;
}

h1 {
  font-size: 28px;
  margin-bottom: 8px;
}

.subtitle {
  opacity: 0.9;
  margin-bottom: 32px;
}

.features {
  text-align: left;
  margin-bottom: 32px;
}

.feature {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255,255,255,0.2);
}

.feature:last-child {
  border-bottom: none;
}

.icon {
  font-size: 24px;
}

.btn-large {
  padding: 14px 40px;
  font-size: 16px;
}
</style>