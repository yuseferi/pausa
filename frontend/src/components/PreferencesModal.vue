<template>
  <div class="modal-overlay fade-in" @click.self="$emit('close')">
    <div class="modal slide-up">
      <div class="modal-header">
        <h2>Preferences</h2>
        <button class="btn-close" @click="$emit('close')">×</button>
      </div>

      <div class="modal-tabs">
        <button v-for="tab in tabs" :key="tab.id" 
                :class="['tab', { active: activeTab === tab.id }]"
                @click="activeTab = tab.id">
          {{ tab.label }}
        </button>
      </div>

      <div class="modal-content">
        <!-- Breaks Tab -->
        <div v-if="activeTab === 'breaks'" class="tab-content">
          <div class="setting-group">
            <h3>Mini Breaks</h3>
            <div class="setting-row">
              <label>Interval (minutes)</label>
              <input type="number" v-model.number="localConfig.miniBreakInterval" min="1" max="60" />
            </div>
            <div class="setting-row">
              <label>Duration (seconds)</label>
              <input type="number" v-model.number="localConfig.miniBreakDuration" min="5" max="120" />
            </div>
            <div class="setting-row">
              <label>Postpone time (minutes)</label>
              <input type="number" v-model.number="localConfig.postponeTimeMini" min="1" max="30" />
            </div>
          </div>

          <div class="setting-group">
            <h3>Long Breaks</h3>
            <div class="setting-row">
              <label>After mini breaks</label>
              <input type="number" v-model.number="localConfig.longBreakInterval" min="1" max="10" />
            </div>
            <div class="setting-row">
              <label>Duration (seconds)</label>
              <input type="number" v-model.number="localConfig.longBreakDuration" min="60" max="1800" />
            </div>
            <div class="setting-row">
              <label>Postpone time (minutes)</label>
              <input type="number" v-model.number="localConfig.postponeTimeLong" min="1" max="60" />
            </div>
          </div>
        </div>

        <!-- Behavior Tab -->
        <div v-if="activeTab === 'behavior'" class="tab-content">
          <div class="setting-row">
            <label>Show exercise tips</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.showExerciseTips" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="setting-row">
            <label>Notify before break</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.notifyBeforeBreak" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="setting-row" v-if="localConfig.notifyBeforeBreak">
            <label>Notify seconds before mini</label>
            <input type="number" v-model.number="localConfig.notifyTimeMini" min="5" max="60" />
          </div>
          <div class="setting-row" v-if="localConfig.notifyBeforeBreak">
            <label>Notify seconds before long</label>
            <input type="number" v-model.number="localConfig.notifyTimeLong" min="10" max="120" />
          </div>
          <div class="setting-row">
            <label>Monitor idle time</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.monitorIdleTime" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="setting-row" v-if="localConfig.monitorIdleTime">
            <label>Idle threshold (minutes)</label>
            <input type="number" v-model.number="localConfig.idleTimeThreshold" min="1" max="30" />
          </div>
        </div>

        <!-- Appearance Tab -->
        <div v-if="activeTab === 'appearance'" class="tab-content">
          <div class="setting-row">
            <label>Theme</label>
            <select v-model="localConfig.theme">
              <option value="system">System</option>
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </div>
          <div class="setting-row">
            <label>Full screen breaks</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.fullScreenBreak" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="setting-row">
            <label>Show on all monitors</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.showOnAllMonitors" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- General Tab -->
        <div v-if="activeTab === 'general'" class="tab-content">
          <div class="setting-row">
            <label>Start at login</label>
            <label class="toggle-switch">
              <input type="checkbox" v-model="localConfig.startAtLogin" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="setting-row">
            <label>Language</label>
            <select v-model="localConfig.language">
              <option value="en">English</option>
              <option value="de">Deutsch</option>
              <option value="es">Español</option>
              <option value="fr">Français</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" @click="save">Save</button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, watch } from 'vue'

export default {
  name: 'PreferencesModal',
  props: {
    config: { type: Object, required: true }
  },
  emits: ['close', 'save'],
  setup(props, { emit }) {
    const activeTab = ref('breaks')
    const tabs = [
      { id: 'breaks', label: 'Breaks' },
      { id: 'behavior', label: 'Behavior' },
      { id: 'appearance', label: 'Appearance' },
      { id: 'general', label: 'General' }
    ]

    const localConfig = reactive({ ...props.config })

    watch(() => props.config, (newConfig) => {
      Object.assign(localConfig, newConfig)
    }, { deep: true })

    const save = () => {
      emit('save', { ...localConfig })
    }

    return { activeTab, tabs, localConfig, save }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: var(--bg-secondary);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}

.modal-header h2 {
  font-size: 18px;
  font-weight: 600;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--text-secondary);
}

.modal-tabs {
  display: flex;
  border-bottom: 1px solid var(--border);
  padding: 0 20px;
}

.tab {
  padding: 12px 16px;
  border: none;
  background: none;
  font-size: 14px;
  cursor: pointer;
  color: var(--text-secondary);
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}

.tab.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.setting-group {
  margin-bottom: 16px;
}

.setting-group h3 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-secondary);
}

.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.setting-row label:first-child {
  font-size: 14px;
}

.setting-row input[type="number"],
.setting-row select {
  width: 100px;
  text-align: center;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}
</style>