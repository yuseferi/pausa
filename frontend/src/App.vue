<template>
  <div class="app-container" :data-theme="theme">
    <WelcomeScreen v-if="showWelcome" @complete="onWelcomeComplete" />
    <BreakWindow 
      v-else-if="isOnBreak" 
      :breakType="currentBreakType"
      :duration="breakDuration"
      :tip="currentTip"
      :fullscreen="config.fullScreenBreak"
      @skip="skipBreak"
      @postpone="postponeBreak"
    />
    <MainView 
      v-else 
      :state="state"
      :config="config"
      @openPreferences="showPreferences = true"
      @pauseBreaks="pauseBreaks"
      @resumeBreaks="resumeBreaks"
      @resetBreaks="resetBreaks"
    />
    <PreferencesModal 
      v-if="showPreferences" 
      :config="config"
      @close="showPreferences = false"
      @save="saveConfig"
    />
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import WelcomeScreen from './components/WelcomeScreen.vue'
import BreakWindow from './components/BreakWindow.vue'
import MainView from './components/MainView.vue'
import PreferencesModal from './components/PreferencesModal.vue'

export default {
  name: 'App',
  components: {
    WelcomeScreen,
    BreakWindow,
    MainView,
    PreferencesModal
  },
  setup() {
    const showWelcome = ref(false)
    const showPreferences = ref(false)
    const isOnBreak = ref(false)
    const currentBreakType = ref('mini')
    const breakDuration = ref(20)
    const currentTip = ref({ text: '', category: '' })
    const state = ref({
      nextBreakTime: new Date(),
      nextBreakType: 'mini',
      miniBreakCount: 0,
      isOnBreak: false,
      isPaused: false
    })
    const config = ref({
      miniBreakInterval: 10,
      miniBreakDuration: 20,
      longBreakInterval: 3,
      longBreakDuration: 300,
      theme: 'system',
      language: 'en'
    })

    const theme = computed(() => {
      if (config.value.theme === 'system') {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
      }
      return config.value.theme
    })

    const loadState = async () => {
      try {
        if (window.go?.main?.App?.GetState) {
          state.value = await window.go.main.App.GetState()
          isOnBreak.value = state.value.isOnBreak
        }
      } catch (e) { console.error(e) }
    }

    const loadConfig = async () => {
      try {
        if (window.go?.main?.App?.GetConfig) {
          config.value = await window.go.main.App.GetConfig()
        }
      } catch (e) { console.error(e) }
    }

    const saveConfig = async (newConfig) => {
      try {
        if (window.go?.main?.App?.SaveConfig) {
          await window.go.main.App.SaveConfig(newConfig)
          config.value = newConfig
        }
        showPreferences.value = false
      } catch (e) { console.error(e) }
    }

    const pauseBreaks = async () => {
      if (window.go?.main?.App?.PauseBreaks) {
        await window.go.main.App.PauseBreaks()
        state.value.isPaused = true
      }
    }

    const resumeBreaks = async () => {
      if (window.go?.main?.App?.ResumeBreaks) {
        await window.go.main.App.ResumeBreaks()
        state.value.isPaused = false
        loadState()
      }
    }

    const resetBreaks = async () => {
      if (window.go?.main?.App?.ResetBreaks) {
        await window.go.main.App.ResetBreaks()
        loadState()
      }
    }

    const skipBreak = async () => {
      if (window.go?.main?.App?.SkipBreak) {
        await window.go.main.App.SkipBreak()
        isOnBreak.value = false
        loadState()
      }
    }

    const postponeBreak = async () => {
      if (window.go?.main?.App?.PostponeBreak) {
        await window.go.main.App.PostponeBreak()
        isOnBreak.value = false
        loadState()
      }
    }

    const loadTip = async () => {
      if (window.go?.main?.App?.GetExerciseTip) {
        currentTip.value = await window.go.main.App.GetExerciseTip('')
      }
    }

    const onWelcomeComplete = async () => {
      if (window.go?.main?.App?.MarkFirstLaunchComplete) {
        await window.go.main.App.MarkFirstLaunchComplete()
      }
      showWelcome.value = false
    }

    let stateInterval

    onMounted(async () => {
      await loadConfig()
      await loadState()
      
      if (window.go?.main?.App?.IsFirstLaunch) {
        showWelcome.value = await window.go.main.App.IsFirstLaunch()
      }

      if (window.runtime?.EventsOn) {
        window.runtime.EventsOn('breakStarted', (data) => {
          isOnBreak.value = true
          currentBreakType.value = data.type
          breakDuration.value = data.duration
          loadTip()
        })
        window.runtime.EventsOn('breakEnded', () => {
          isOnBreak.value = false
          loadState()
        })
      }
      
      stateInterval = setInterval(loadState, 1000)
    })

    onUnmounted(() => {
      if (stateInterval) clearInterval(stateInterval)
    })

    return {
      showWelcome, showPreferences, isOnBreak, currentBreakType,
      breakDuration, currentTip, state, config, theme,
      onWelcomeComplete, saveConfig, pauseBreaks, resumeBreaks,
      resetBreaks, skipBreak, postponeBreak
    }
  }
}
</script>

<style scoped>
.app-container {
  height: 100%;
  width: 100%;
}
</style>