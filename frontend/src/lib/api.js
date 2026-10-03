// Typed wrapper around the Wails-bound Go App. We import the generated
// bindings directly so we never have to wait for `window.go.*` to populate;
// each binding is a function that resolves once the runtime is ready.

import {
  CheckForUpdates,
  EndBreak,
  GetConfig,
  GetSnapshot,
  GetTips,
  GetVersion,
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

// Wails' EventsOff removes *every* listener for an event, so we keep our own
// registry per event. Each api.on call gets a unique registration token so a
// disposer removes only its own registration — even when the same handler is
// registered twice, or after api.off() and a later re-registration.
const listeners = new Map() // event -> Set<registration>
const subscribed = new Set()

function dispatch(event, ...args) {
  const set = listeners.get(event)
  if (!set) return
  for (const reg of [...set]) {
    try {
      reg.handler(...args)
    } catch (err) {
      console.error(`api.on(${event}) handler failed:`, err)
    }
  }
}

export const api = {
  // Read
  getConfig:        safe('getConfig', GetConfig),
  getSnapshot:      safe('getSnapshot', GetSnapshot),
  getTips:          safe('getTips', GetTips),
  isFirstLaunch:    safe('isFirstLaunch', IsFirstLaunch),
  getVersion:       safe('getVersion', GetVersion),
  checkForUpdates:  safe('checkForUpdates', CheckForUpdates),

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
    let set = listeners.get(event)
    if (!set) {
      set = new Set()
      listeners.set(event, set)
    }
    // A unique token per api.on call, so the disposer can only ever remove
    // its own registration.
    const reg = { handler }
    set.add(reg)
    if (!subscribed.has(event)) {
      subscribed.add(event)
      EventsOn(event, (...args) => dispatch(event, ...args))
    }
    return () => {
      const s = listeners.get(event)
      if (!s || !s.delete(reg)) return
      if (s.size === 0) {
        listeners.delete(event)
        subscribed.delete(event)
        EventsOff(event)
      }
    }
  },
  off(event) {
    listeners.delete(event)
    if (subscribed.delete(event)) EventsOff(event)
  },
}
