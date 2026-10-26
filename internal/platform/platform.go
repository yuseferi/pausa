// Package platform is the single seam between Pausa's portable core and the
// host operating system.
//
// Everything above this package — the scheduler, breakapp, and main — is
// platform-agnostic: it talks only to the types and functions declared here.
// Each supported OS supplies an implementation in a build-tagged file:
//
//   - platform_darwin.go — the macOS implementation, backed by internal/macos
//     (AppKit / CoreAudio / MediaRemote via cgo).
//   - platform_other.go — the Windows/Linux lite implementation: Wails
//     notifications, no native overlays, tray, busy or idle detection.
//
// Callers must not assume a StatusBar, overlay, or notification is actually
// rendered: check Capabilities (or treat the no-ops as no-ops).
package platform

import "time"

// Capabilities describes what the running platform's implementation can
// actually do. The core reads it to choose a fallback behaviour; the UI reads
// it to hide preferences that would otherwise be silent no-ops.
type Capabilities struct {
	// NativeOverlays means the OS draws the break UI itself (macOS
	// NSPanels). When false, CreateOverlays is a no-op and the caller shows
	// the Wails window for the break instead.
	NativeOverlays bool `json:"nativeOverlays"`
	// SystemTray means a persistent menu-bar / tray surface exists, so the
	// window may hide on close and be restored from there.
	SystemTray bool `json:"systemTray"`
	// DockIcon means the app can switch between dock and accessory
	// activation policies.
	DockIcon bool `json:"dockIcon"`
	// LoginItem means launch-at-login can actually be toggled.
	LoginItem bool `json:"loginItem"`
	// BusyDetection means meetings and media playback can auto-pause the
	// schedule.
	BusyDetection bool `json:"busyDetection"`
	// IdleDetection means idle time is observable, so "away" pauses and
	// natural breaks can work.
	IdleDetection bool `json:"idleDetection"`
}

// Caps is the capability set for the running platform. The platform's init
// sets it once; nothing mutates it afterwards.
var Caps Capabilities

// MenuTag identifies a status-item / tray menu entry. Zero and negative tags
// are reserved.
type MenuTag int

// StatusIconState selects the built-in status-item glyph that reflects the
// scheduler state. Platforms that have no built-in glyphs may ignore it.
type StatusIconState int

const (
	StatusIconRunning StatusIconState = iota
	StatusIconPaused
	StatusIconBusy
	StatusIconBreak
)

// NotifAction is a notification button the user can press. The values mirror
// internal/macos and are asserted equal in platform_darwin_test.go.
type NotifAction int

const (
	NotifDefault NotifAction = iota + 1
	NotifSkip
	NotifPostpone
)

// OverlayAction is a button on the break overlay. Values mirror
// internal/macos and are asserted equal in platform_darwin_test.go.
type OverlayAction int

const (
	OverlaySkip     OverlayAction = 1
	OverlayPostpone OverlayAction = 2
)

// OverlayOptions describes a break overlay to show.
type OverlayOptions struct {
	Kind, Title, Timer, Tip, HexAccent         string
	ShowActions, Fullscreen, CurrentScreenOnly bool
}

// StatusBar is the menu-bar (macOS) / tray (Windows, Linux) surface. Handlers
// fire off the caller's goroutine; do real work on a channel.
type StatusBar interface {
	SetTitle(title string)
	SetBuiltinIcon(state StatusIconState)
	SetImage(pngData []byte)
	AddItem(tag MenuTag, title string, onClick func())
	AddDisabled(tag MenuTag, title string)
	AddSeparator()
	UpdateItem(tag MenuTag, title string)
	SetItemHidden(tag MenuTag, hidden bool)
	Teardown()
}

// BusySource reports whether the user is currently "busy" — in a meeting,
// playing a video, etc. — so the scheduler can pause instead of interrupting.
// It is polled every scheduler.BusyPollInterval, so BusyState must be cheap.
type BusySource interface {
	BusyState() (label string, busy bool)
	Close()
}

// IdleSource reports how long the user has been idle. A source that always
// returns 0 disables idle behavior.
type IdleSource interface {
	IdleFor() time.Duration
}
