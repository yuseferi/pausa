// Typed wrapper around the Wails-bound Go App. We import the generated
// bindings directly so we never have to wait for `window.go.*` to populate;
// each binding is a function that resolves once the runtime is ready.

import {
  EndBreak,
  GetConfig,
  GetSnapshot,
  GetTips,
  IsFirstLaunch,
  MarkFirstLaunchComplete,
  PauseBreaks,
  PostponeBreak,
  QuitApp,
  ResetBreaks,
  ResumeBreaks,
  SaveConfig,
  ShowMainWindow,
  SkipBreak,
  TakeBreakNow,
} from '../../wailsjs/go/breakapp/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

// Wrap each binding so it logs (rather than throws) on failure. UI code
// can `await` without try/catch when it doesn't care about the error.
const safe = (name, fn) => async (...args) => {
  try {
    return await fn(...args)
  } catch (err) {
    console.error(`api.${name} failed:`, err)
    return undefined
  }
}

export const api = {
  // Read
  getConfig:        safe('getConfig', GetConfig),
  getSnapshot:      safe('getSnapshot', GetSnapshot),
  getTips:          safe('getTips', GetTips),
  isFirstLaunch:    safe('isFirstLaunch', IsFirstLaunch),

  // Write
  saveConfig:               safe('saveConfig', SaveConfig),
  markFirstLaunchComplete:  safe('markFirstLaunchComplete', MarkFirstLaunchComplete),

  // Commands
  takeBreakNow:    safe('takeBreakNow', TakeBreakNow),
  endBreak:        safe('endBreak', EndBreak),
  skipBreak:       safe('skipBreak', SkipBreak),
  postponeBreak:   safe('postponeBreak', PostponeBreak),
  pauseBreaks:     safe('pauseBreaks', PauseBreaks),
  resumeBreaks:    safe('resumeBreaks', ResumeBreaks),
  resetBreaks:     safe('resetBreaks', ResetBreaks),
  showMainWindow:  safe('showMainWindow', ShowMainWindow),
  quitApp:         safe('quitApp', QuitApp),

  // Events
  on(event, handler) {
    EventsOn(event, handler)
    return () => EventsOff(event)
  },
  off(event) { EventsOff(event) },
}
