// Package breakapp wires together the scheduler, configuration store, tip
// catalog, and macOS bridge into a single Wails-bindable type. It is the
// thin facade exposed to JavaScript and the only file that imports the
// Wails runtime.
package breakapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pausa/internal/config"
	"pausa/internal/macos"
	"pausa/internal/scheduler"
	"pausa/internal/tips"
)

// Status-bar menu item tags. Negative numbers reserved for internal use.
const (
	tagNextBreak macos.MenuTag = 1
	tagTakeBreak macos.MenuTag = 2
	tagPause     macos.MenuTag = 3
	tagResume    macos.MenuTag = 4
	tagShow      macos.MenuTag = 5
	tagPrefs     macos.MenuTag = 6
	tagQuit      macos.MenuTag = 7
)

// App is the Wails-bound application object. It owns the scheduler,
// configuration store, and tip catalog; coordinates UI updates; and
// translates native macOS events (status-bar clicks, notification actions)
// into scheduler commands.
type App struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	cfg    *config.Store
	sched  *scheduler.Scheduler
	tips   *tips.Catalog
	busy   *macos.BusySource
	status *macos.StatusBar

	// nativeOnce guards setupStatusBar / setupNotifications so they're only
	// invoked the first time DomReady fires.
	nativeOnce sync.Once

	// Last-known break duration so the front-end can render the right
	// initial state if it loads mid-break.
	mu              sync.RWMutex
	currentBreakDur time.Duration
	currentTip      tips.Tip

	// lastBarTitle dedupes menu-bar title updates. Only touched by the
	// eventPump goroutine.
	lastBarTitle string
}

// New constructs an App. Call Startup from the Wails OnStartup callback.
func New(cfg *config.Store, sched *scheduler.Scheduler, catalog *tips.Catalog, busy *macos.BusySource) *App {
	return &App{
		cfg:   cfg,
		sched: sched,
		tips:  catalog,
		busy:  busy,
	}
}

// ---------- Wails lifecycle ----------

// Startup is invoked by Wails once the runtime context is ready. It only
// starts the pure-Go scheduler and event pump; native AppKit setup is
// deferred to DomReady so we know the run loop is fully spinning.
func (a *App) Startup(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.ctx = ctx
	a.cancelCtx = cancel

	a.sched.Start(ctx)
	go a.eventPump(ctx)
}

// DomReady is invoked once the JS side is ready to receive events. By this
// point AppKit has finished applicationDidFinishLaunching, so it's safe to
// install the status item and request notification permission.
func (a *App) DomReady(ctx context.Context) {
	a.nativeOnce.Do(func() {
		// Switch to accessory activation policy *before* the first break
		// fires. Without this, macOS treats Pausa as a regular app and
		// refuses to render its windows on top of fullscreen-app Spaces.
		// As a side effect, this also removes the Dock icon — Pausa is
		// menu-bar-driven anyway. Users who opted into ShowInDock keep
		// the regular policy (Dock icon) at the cost of that capability.
		if a.cfg.Get().General.ShowInDock {
			macos.SetRegularActivationPolicy()
		} else {
			macos.SetAccessoryActivationPolicy()
		}
		a.setupStatusBar()
		a.setupNotifications()
		a.setupOverlayHandlers()
	})
	wailsruntime.EventsEmit(ctx, "scheduler:hydrate", a.sched.Snapshot())
}

// Shutdown is invoked when the app quits.
func (a *App) Shutdown(ctx context.Context) {
	if a.cancelCtx != nil {
		a.cancelCtx()
	}
	a.sched.Stop()
	if a.status != nil {
		a.status.Teardown()
	}
	macos.CloseOverlays()
	a.cfg.Close()
}

// ---------- Bound methods (JS → Go) ----------

// GetConfig returns the current configuration.
func (a *App) GetConfig() config.Config { return a.cfg.Get() }

// SaveConfig validates and persists a new configuration. Returns the
// (possibly clamped) config that was saved.
func (a *App) SaveConfig(c config.Config) (config.Config, error) {
	prev := a.cfg.Get()
	saved, err := a.cfg.Set(c)
	if err != nil {
		slog.Error("save config", "err", err)
		return saved, err
	}
	if a.busy != nil {
		a.busy.SetMediaDebounce(saved.Idle.BusyMediaDebounce.AsDuration())
	}
	if saved.General.StartAtLogin != prev.General.StartAtLogin {
		if ok := macos.SetLoginItemEnabled(saved.General.StartAtLogin); !ok {
			slog.Warn("login item change failed", "enabled", saved.General.StartAtLogin)
		}
	}
	if saved.General.ShowInDock != prev.General.ShowInDock {
		if saved.General.ShowInDock {
			macos.SetRegularActivationPolicy()
		} else {
			macos.SetAccessoryActivationPolicy()
		}
	}
	wailsruntime.EventsEmit(a.ctx, "config:updated", saved)
	return saved, nil
}

// GetSnapshot returns the scheduler's current snapshot.
func (a *App) GetSnapshot() scheduler.Snapshot { return a.sched.Snapshot() }

// GetTips returns all known exercise tips.
func (a *App) GetTips() []tips.Tip { return a.tips.All() }

// IsFirstLaunch reports whether the configuration file existed at startup.
func (a *App) IsFirstLaunch() bool { return !a.cfg.Exists() }

// MarkFirstLaunchComplete persists the (possibly default) config so future
// launches know first-run is done.
func (a *App) MarkFirstLaunchComplete() error {
	_, err := a.cfg.Set(a.cfg.Get())
	return err
}

// TakeBreakNow triggers an immediate break.
func (a *App) TakeBreakNow() { a.sched.TakeBreakNow() }

// EndBreak ends the current break.
func (a *App) EndBreak() { a.sched.EndBreak() }

// SkipBreak skips the current break.
func (a *App) SkipBreak() {
	a.sched.SkipBreak()
}

// PostponeBreak postpones the current or upcoming break.
func (a *App) PostponeBreak() { a.sched.PostponeBreak() }

// PauseBreaks pauses the schedule.
func (a *App) PauseBreaks() { a.sched.Pause() }

// ResumeBreaks resumes the schedule.
func (a *App) ResumeBreaks() { a.sched.Resume() }

// ResetBreaks resets the schedule.
func (a *App) ResetBreaks() { a.sched.Reset() }

// ShowMainWindow brings the main window into the foreground.
func (a *App) ShowMainWindow() { wailsruntime.WindowShow(a.ctx) }

// QuitApp exits the application.
func (a *App) QuitApp() { wailsruntime.Quit(a.ctx) }

// OpenPreferences shows the main window and emits the open-preferences UI
// event. Used by the macOS top-bar menu.
func (a *App) OpenPreferences() {
	wailsruntime.WindowShow(a.ctx)
	wailsruntime.EventsEmit(a.ctx, "ui:open-preferences")
}

// ---------- Internal: status bar ----------

func (a *App) setupStatusBar() {
	// Use a native vector template icon. This avoids raster/template
	// rendering glitches (the old gray-square issue) and lets us reflect
	// running/paused/busy/break states directly in the menu bar.
	sb := macos.SetupStatusBar("")
	a.status = sb
	sb.SetBuiltinIcon(macos.StatusIconRunning)

	sb.AddDisabled(tagNextBreak, "Loading…")
	sb.AddSeparator()
	sb.AddItem(tagTakeBreak, "Take a Break Now", a.sched.TakeBreakNow)
	sb.AddSeparator()
	sb.AddItem(tagPause, "Pause Breaks", a.sched.Pause)
	sb.AddItem(tagResume, "Resume Breaks", a.sched.Resume)
	sb.SetItemHidden(tagResume, true)
	sb.AddSeparator()
	sb.AddItem(tagShow, "Show Pausa", func() {
		wailsruntime.WindowShow(a.ctx)
	})
	sb.AddItem(tagPrefs, "Preferences…", func() {
		wailsruntime.WindowShow(a.ctx)
		wailsruntime.EventsEmit(a.ctx, "ui:open-preferences")
	})
	sb.AddSeparator()
	sb.AddItem(tagQuit, "Quit Pausa", func() { wailsruntime.Quit(a.ctx) })
}

func (a *App) updateStatusBar(snap scheduler.Snapshot) {
	if a.status == nil {
		return
	}
	cfg := a.cfg.Get()
	title := "" // menu-bar countdown text (only while scheduled)

	switch snap.Phase {
	case scheduler.PhasePaused:
		a.status.SetBuiltinIcon(macos.StatusIconPaused)
		a.status.UpdateItem(tagNextBreak, "Breaks paused")
		a.status.SetItemHidden(tagPause, true)
		a.status.SetItemHidden(tagResume, false)
	case scheduler.PhaseAutoPaused:
		a.status.SetBuiltinIcon(macos.StatusIconBusy)
		// Auto-paused (busy detection). Show why so the user doesn't
		// wonder why their break didn't fire.
		label := snap.AutoPauseReason
		if label == "" {
			label = "busy"
		}
		a.status.UpdateItem(tagNextBreak, "Paused — "+label)
		// Allow manual pause/resume even while auto-paused; manual pause
		// takes precedence and survives busy-clear.
		a.status.SetItemHidden(tagPause, false)
		a.status.SetItemHidden(tagResume, true)
	case scheduler.PhaseOnBreak:
		a.status.SetBuiltinIcon(macos.StatusIconBreak)
		a.status.UpdateItem(tagNextBreak, "Break in progress")
		a.status.SetItemHidden(tagPause, false)
		a.status.SetItemHidden(tagResume, true)
	default:
		a.status.SetBuiltinIcon(macos.StatusIconRunning)
		left := time.Until(snap.NextBreakAt)
		if left < 0 {
			left = 0
		}
		kind := "Short"
		if snap.NextKind == scheduler.BreakLong {
			kind = "Long"
		}
		a.status.UpdateItem(tagNextBreak, fmt.Sprintf("%s break in %s", kind, formatCountdown(left)))
		a.status.SetItemHidden(tagPause, false)
		a.status.SetItemHidden(tagResume, true)
		if cfg.General.StatusBarTitle {
			title = formatCountdown(left)
		}
	}

	// Only touch the title when it actually changed; this runs every second.
	if title != a.lastBarTitle {
		a.status.SetTitle(title)
		a.lastBarTitle = title
	}
}

// ---------- Internal: notifications ----------

func (a *App) setupNotifications() {
	macos.SetNotificationHandler(func(action macos.NotifAction) {
		switch action {
		case macos.NotifSkip:
			a.sched.SkipBreak()
		case macos.NotifPostpone:
			a.sched.PostponeBreak()
		case macos.NotifDefault:
			a.sched.TakeBreakNow()
		}
	})
}

// setupOverlayHandlers wires button clicks inside the native overlay
// (Skip / Postpone) into the scheduler.
func (a *App) setupOverlayHandlers() {
	macos.SetOverlayActionHandler(func(action macos.OverlayAction) {
		switch action {
		case macos.OverlaySkip:
			a.sched.SkipBreak()
		case macos.OverlayPostpone:
			a.sched.PostponeBreak()
		}
	})
}

// ---------- Internal: event pump ----------

// eventPump translates scheduler events into Wails events for the UI and
// orchestrates native overlay/notification side effects. It also runs a
// 1-second status-bar refresh so the menu-bar countdown stays current.
func (a *App) eventPump(ctx context.Context) {
	statusTicker := time.NewTicker(time.Second)
	defer statusTicker.Stop()

	previousApp := ""

	for {
		select {
		case <-ctx.Done():
			return

		case <-statusTicker.C:
			a.updateStatusBar(a.sched.Snapshot())

		case ev, ok := <-a.sched.Events():
			if !ok {
				return
			}
			a.handleSchedulerEvent(ev, &previousApp)
		}
	}
}

func (a *App) handleSchedulerEvent(ev scheduler.Event, previousApp *string) {
	cfg := a.cfg.Get()

	switch ev.Kind {
	case scheduler.EventNotify:
		// Capture the frontmost app NOW, before the user can interact with
		// the notification (which would change focus). If we wait until
		// EventBreakStart we may have already activated Pausa and lose the
		// real previous-app context, especially when the previous app was
		// in a fullscreen Space.
		if name := macos.FrontmostApp(); name != "" {
			*previousApp = name
			slog.Info("captured previous app", "name", name)
		}
		if cfg.Notification.Enabled {
			label := "short"
			if ev.BreakKind == scheduler.BreakLong {
				label = "long"
			}
			macos.Notify(
				fmt.Sprintf("Your %s break starts soon", label),
				fmt.Sprintf("In %d seconds. Tap to start now.", ev.SecondsUntil),
				cfg.Notification.ShowActions,
				cfg.Notification.PlaySound,
			)
		}

	case scheduler.EventBreakStart:
		// Last chance to capture the previous app, in case there was no
		// pre-break notification (warning disabled, or user invoked
		// "Take a break now" directly).
		if *previousApp == "" {
			if name := macos.FrontmostApp(); name != "" {
				*previousApp = name
				slog.Info("captured previous app at break start", "name", name)
			}
		}

		tip := a.pickTip(ev.BreakKind)
		dur := ev.Snapshot.BreakEndsAt.Sub(time.Now())
		if dur < 0 {
			dur = 0
		}
		a.mu.Lock()
		a.currentBreakDur = dur
		a.currentTip = tip
		a.mu.Unlock()

		// Render the break as native NSPanel overlays on EVERY screen.
		// We deliberately do NOT show the Wails main window — standard
		// NSWindows can't appear on fullscreen-app Spaces, but our panels
		// can. This way every monitor gets the same UI regardless of
		// which Space the user is on.
		title := "Short Break"
		kindStr := "short"
		if ev.BreakKind == scheduler.BreakLong {
			title = "Long Break"
			kindStr = "long"
		}
		slog.Info("break started",
			"kind", ev.BreakKind,
			"screenCount", macos.ScreenCount(),
			"allMonitors", cfg.Display.AllMonitors,
			"fullscreen", cfg.Display.Fullscreen)

		macos.CreateOverlays(macos.OverlayOptions{
			Kind:              kindStr,
			Title:             title,
			Timer:             formatTimer(int(dur.Seconds())),
			Tip:               tip.Text,
			HexAccent:         cfg.Display.AccentColor,
			ShowActions:       true,
			Fullscreen:        cfg.Display.Fullscreen,
			CurrentScreenOnly: !cfg.Display.AllMonitors,
		})

	case scheduler.EventBreakTick:
		macos.UpdateOverlayTimer(formatTimer(ev.SecondsLeft))

	case scheduler.EventBreakEnd, scheduler.EventSkipped, scheduler.EventNatural:
		// Close every overlay and return focus to the previous app.
		macos.CloseOverlays()

		prev := *previousApp
		*previousApp = ""
		slog.Info("break ended, restoring previous app", "name", prev)
		macos.HideSelfAndActivate(prev)
	}

	// Always emit a typed event for the frontend.
	wailsruntime.EventsEmit(a.ctx, "scheduler:event", a.frontendEvent(ev))
	a.updateStatusBar(ev.Snapshot)
}

// frontendEvent enriches the scheduler event with anything the UI needs but
// the scheduler doesn't carry (tips, durations).
func (a *App) frontendEvent(ev scheduler.Event) map[string]interface{} {
	out := map[string]interface{}{
		"kind":         string(ev.Kind),
		"snapshot":     ev.Snapshot,
		"breakKind":    string(ev.BreakKind),
		"secondsLeft":  ev.SecondsLeft,
		"secondsUntil": ev.SecondsUntil,
	}
	a.mu.RLock()
	if ev.Kind == scheduler.EventBreakStart {
		out["tip"] = a.currentTip
		out["durationSeconds"] = int(a.currentBreakDur.Seconds())
	}
	a.mu.RUnlock()
	return out
}

func (a *App) pickTip(kind scheduler.BreakKind) tips.Tip {
	if kind == scheduler.BreakLong {
		return a.tips.PickForLong()
	}
	return a.tips.PickForShort()
}

func formatTimer(secs int) string {
	if secs < 0 {
		secs = 0
	}
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// formatCountdown renders a time-until-break for the menu bar. Waits over an
// hour (e.g. deferred to the next working window) render as "16h05m".
func formatCountdown(d time.Duration) string {
	total := int(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
