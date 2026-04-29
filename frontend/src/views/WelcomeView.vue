<script setup>
import { completeWelcome } from '../lib/store'
</script>

<template>
  <div class="welcome">
    <div class="titlebar-drag" />

    <div class="welcome__inner">
      <div class="welcome__logo">☕</div>
      <h1 class="welcome__title">Welcome to Pausa</h1>
      <p class="welcome__sub">A mindful break reminder for healthier work.</p>

      <ul class="welcome__features">
        <li><span>👁️</span> Rest your eyes every 10 minutes</li>
        <li><span>🧘</span> Take a longer break every 30 minutes</li>
        <li><span>🌿</span> Pauses when you step away</li>
      </ul>

      <button class="btn btn--primary btn--lg" @click="completeWelcome">
        Get started
      </button>
    </div>
  </div>
</template>

<style scoped>
.welcome {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(160deg, var(--accent), color-mix(in srgb, var(--accent) 60%, #1e3a8a));
  color: white;
}
.welcome__inner {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 32px;
  gap: 16px;
}
.welcome__logo { font-size: 64px; margin-bottom: 8px; }
.welcome__title { font-size: 28px; font-weight: 600; }
.welcome__sub { opacity: 0.9; font-size: 15px; }

.welcome__features {
  list-style: none;
  text-align: left;
  margin: 16px 0 24px;
  width: 100%;
  max-width: 320px;
}
.welcome__features li {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255,255,255,0.18);
  font-size: 14px;
}
.welcome__features li:last-child { border-bottom: none; }
.welcome__features span { font-size: 22px; }

.btn--primary {
  background: white;
  color: var(--accent);
}
.btn--primary:hover { background: rgba(255,255,255,0.92); }
</style>
