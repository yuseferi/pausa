# Pausa — Architecture

This document describes the rebuilt Pausa codebase. Keep it in sync with
the code; if you find a discrepancy, the code wins.

## Goals

1. **Correctness over cleverness.** Bugs in a break reminder are silent —
   the user just stops getting reminders. The state machine and tests are
   the most important code in the repo.
2. **One source of truth.** Time, configuration, and break state each have
   exactly one owner.
3. **Native macOS, properly.** No AppleScript shell-outs, no polling for
   menu clicks, real `UNUserNotificationCenter` notifications, real IOKit
   idle detection.
4. **Frontend renders, doesn't compute.** The Vue side never owns the break
   countdown — it just displays the value the backend ticks every second.

## Layout

```
pausa/
├── main.go                          # ~100 LOC entrypoint
├── internal/
│   ├── log/      logger.go          # slog → ~/Library/Logs/Pausa/pausa.log
│   ├── clock/    clock.go           # Clock interface + System + FakeClock
│   ├── config/   config.go          # typed nested config
│   │             duration.go        # Duration JSON marshalling
│   │             store.go           # atomic load/save + pub/sub
│   ├── tips/     tips.go            # exercise-tip catalog
│   ├── scheduler/state.go           # FSM types
│   │             scheduler.go       # actor goroutine
│   │             scheduler_test.go  # 17 tests w/ FakeClock
│   ├── macos/    bridge.h           # shared C declarations
│   │             bridge.m           # AppKit / UN / IOKit
│   │             cgo_darwin.go      # cgo flags
│   │             {statusbar,
│   │              notifications,
│   │              workspace,
│   │              idle,
│   │              windows}_darwin.go
│   │             stubs_other.go     # no-op stubs for !darwin
│   └── breakapp/ app.go             # Wails-bound facade
└── frontend/src/
    ├── main.js
    ├── App.vue                      # router (welcome | break | dashboard)
    ├── style.css
    ├── lib/
    │   ├── api.js                   # typed Wails RPC wrapper
    │   ├── store.js                 # reactive() store + actions
    │   └── duration.js              # Go duration ↔ {value, unit}
    ├── composables/
    │   ├── useCountdown.js          # rAF-based countdown
    │   └── useShortcuts.js          # keyboard shortcuts
    ├── views/
    │   ├── DashboardView.vue
    │   ├── BreakView.vue
    │   ├── PreferencesView.vue
    │   └── WelcomeView.vue
    └── components/
        ├── BreakCountdown.vue
        ├── BreathingGuide.vue
        ├── ExerciseTip.vue
        ├── LongBreakProgress.vue
        ├── StatsRow.vue
        └── preferences/Section{Schedule,Notifications,Display,
                                WorkingHours,Idle,General}.vue
```

## Concurrency Model

The scheduler is the only place mutable state lives, and it's owned by a
single goroutine. The "actor pattern" makes mutexes unnecessary:

```
                 ┌─────────────────────────────────────────┐
                 │           scheduler.run()               │
                 │  for {                                  │
                 │      select {                           │
                 │      case <-ctx.Done(): return          │
                 │      case cfg := <-cfgCh: …             │
                 │      case cmd := <-cmds: cmd.apply(…)   │
                 │      case <-breakTimer.C: …             │
                 │      case <-notifyTimer.C: …            │
                 │      case <-tickTimer.C: …              │
                 │      }                                  │
                 │  }                                      │
                 └─────────────────────────────────────────┘
                                ▲              │
                       commands │              │ events
                                │              ▼
        ┌──────────────────────┴───────┐  ┌──────────────────┐
        │ breakapp.App methods (Wails) │  │ breakapp.eventPump│
        │ - TakeBreakNow → cmds <-     │  │  reads events,    │
        │ - SaveConfig   → cfgCh <-    │  │  emits Wails ev,  │
        │ - PauseBreaks  → cmds <-     │  │  drives macOS     │
        └──────────────────────────────┘  └──────────────────┘
```

Anything that wants to mutate state sends a `command` over `s.cmds`. Timers
fire as channel reads on the same select. There are no mutexes inside the
actor.

External observers get a snapshot via `s.Snapshot()` (synchronous round-trip
through a response channel) or subscribe to `s.Events()`.

## State Machine

```
                ┌─────────────┐
       Start →  │   Idle      │ ←─────────────────────────┐
                └──────┬──────┘                            │
                       │ scheduleNext                      │
                       ▼                                   │
                ┌─────────────┐    Pause    ┌──────────┐   │
                │ Scheduled   │ ──────────→ │ Paused   │   │
                └──────┬──────┘ ←────Resume └──────────┘   │
                       │ T - notifyWarn                    │
                       ▼                                   │
                ┌─────────────┐                            │
                │ Notifying   │                            │
                └──────┬──────┘                            │
                       │ T - 0                             │
                       │  (or cmdTakeBreak from any state) │
                       ▼                                   │
                ┌─────────────┐                            │
                │ OnBreak     │ ── Skip / Postpone ────────┘
                └──────┬──────┘
                       │ duration elapsed (or cmdEndBreak)
                       └──────────────────────────────────→ Idle
```

`shortsCompleted` increments after every short. When it reaches
`config.LongEvery`, the *next* break becomes a long break (then the counter
resets when the long break ends). Bug-for-bug compatibility with the old
implementation is *not* preserved here — the previous code scheduled long
breaks at `interval × count` minutes; the new code correctly schedules them
at `interval` minutes after the threshold short break.

## Key design choices

### `Clock` interface
Production uses `clock.System`; tests use `clock.FakeClock` whose `Advance`
method drives virtual time deterministically. Every timer the scheduler
creates goes through this interface. Result: 17 scheduler tests run in
~150ms with no `time.Sleep`s.

### Config: structured + validated
`config.Config` is a tree of small structs (`ScheduleConfig`,
`NotificationConfig`, …). `Validate()` clamps every numeric field to a sane
range before persistence. Durations marshal as Go duration strings ("10m",
"20s") so the file is hand-editable. Atomic write (`*.tmp` then `rename`)
guarantees the file is never half-written.

Runtime state (`is the user paused right now?`) lives only in scheduler
memory, not in the config file. A crash mid-break can't strand the user in
a paused state.

### Native macOS bridge (no AppleScript)
- **Status bar**: `NSStatusItem` with menu items whose target is a Go
  callback exported via `//export pausaStatusBarClicked`. No polling.
- **Notifications**: `UNUserNotificationCenter` with action buttons (Skip /
  Postpone). User taps go through `//export pausaNotificationAction`
  straight into the scheduler.
- **Frontmost app / activation**: `NSWorkspace`. Returns bundle IDs (more
  reliable than localized names) and uses
  `activateWithOptions:NSApplicationActivateIgnoringOtherApps`.
- **Idle**: `CGEventSourceSecondsSinceLastEventType` (no polling required;
  IOKit caches it).
- **Multi-monitor overlays**: borderless `NSWindow` per non-primary
  `NSScreen` with a CAGradient background. The Wails main window handles the
  primary screen; overlays handle the rest.

All AppKit calls dispatch onto the main queue inside `bridge.m`, so Go
callers can safely invoke them from any goroutine.

### Frontend: backend-driven
- The store (`lib/store.js`) is a `reactive()` Vue object. There's no Pinia,
  no Vuex — just `import { state, api } from './lib/store'`.
- The break countdown is driven by `EventBreakTick` events from the
  scheduler. Vue *never* runs its own setInterval for the break — only the
  next-break-in display uses a `useCountdown` composable, and that
  recomputes from wall-clock every animation frame so it survives sleep.
- Components are small (BreakCountdown, BreathingGuide, ExerciseTip, …) and
  pure — they take props and emit events.

### Cross-platform stubs
Every macOS bridge file has a `//go:build darwin` tag and a sibling
`stubs_other.go` with `//go:build !darwin`. The rest of the codebase doesn't
care; `GOOS=linux CGO_ENABLED=0 go build ./...` succeeds. The app won't
*work* on Linux, but it compiles, which keeps CI fast and discourages
darwin-specific imports leaking outside `internal/macos`.

## Testing

```
go test -race ./...        # 27 tests, ~1s
go vet ./...               # zero issues
GOOS=linux CGO_ENABLED=0 go build ./...  # cross-compile sanity
npm --prefix frontend run build           # frontend type-check + bundle
```

Coverage by package:
- `internal/clock` — 4 tests covering fake-clock semantics
- `internal/config` — 6 tests: defaults, validation, atomic store, pub/sub,
  working-hours predicate
- `internal/scheduler` — 17 tests: every transition, postpone, pause,
  natural breaks, config changes mid-flight

The scheduler tests are the most valuable test suite in the project — they
exercise every state transition without ever calling `time.Sleep`. If a
future change breaks the long-break math again, a test fires immediately.

## Adding features

Common changes:

| Change | Where | Roughly |
|--------|-------|---------|
| New config setting | `internal/config/config.go` + a section component | 10–20 LOC |
| New scheduler behavior | new test in `scheduler_test.go`, new transition | 30–50 LOC |
| New macOS integration | new `*_darwin.go` + bridge function in `.h`/`.m` | 40–80 LOC |
| New view | `frontend/src/views/*.vue`, route from `App.vue` | 50–100 LOC |
| New status-bar menu item | a `tag*` const + `sb.AddItem()` call in `breakapp.app.go` | 5 LOC |

## What's not in scope (yet)

- **Cursor-following countdown** (à la DeskRest) — needs a tiny transparent
  always-on-top NSWindow tracked to the cursor; would live in `macos`.
- **Focus/Do Not Disturb integration** — there's no public macOS API for
  this; possible workarounds use private symbols or scriptableness.
- **Statistics persistence** — currently in-memory only. Would add a
  `internal/stats` package that writes ndjson to
  `~/Library/Application Support/Pausa/stats.log`.
- **Custom exercises UI** — the `tips.Catalog` already supports `Add` /
  `Remove`; just need a frontend section.
