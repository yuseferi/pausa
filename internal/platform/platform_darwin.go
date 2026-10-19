//go:build darwin

package platform

import (
	"pausa/internal/macos"
)

// darwinStatusBar adapts *macos.StatusBar to the platform.StatusBar interface.
// The bridge uses its own MenuTag/StatusIconState types, so the adapter is the
// single place that translates between the two.
type darwinStatusBar struct{ inner *macos.StatusBar }

var _ StatusBar = darwinStatusBar{}

func (b darwinStatusBar) SetTitle(title string) { b.inner.SetTitle(title) }

func (b darwinStatusBar) SetBuiltinIcon(state StatusIconState) {
	b.inner.SetBuiltinIcon(macos.StatusIconState(state))
}

func (b darwinStatusBar) SetImage(pngData []byte) { b.inner.SetImage(pngData) }

func (b darwinStatusBar) AddItem(tag MenuTag, title string, onClick func()) {
	b.inner.AddItem(macos.MenuTag(tag), title, onClick)
}

func (b darwinStatusBar) AddDisabled(tag MenuTag, title string) {
	b.inner.AddDisabled(macos.MenuTag(tag), title)
}

func (b darwinStatusBar) AddSeparator() { b.inner.AddSeparator() }

func (b darwinStatusBar) UpdateItem(tag MenuTag, title string) {
	b.inner.UpdateItem(macos.MenuTag(tag), title)
}

func (b darwinStatusBar) SetItemHidden(tag MenuTag, hidden bool) {
	b.inner.SetItemHidden(macos.MenuTag(tag), hidden)
}

func (b darwinStatusBar) Teardown() { b.inner.Teardown() }

// Compile-time guarantees for the remaining platform interfaces.
var (
	_ BusySource = (*macos.BusySource)(nil)
	_ IdleSource = macos.IdleSource{}
)

// SetupStatusBar installs the menu-bar item and returns its handle.
func SetupStatusBar(title string) StatusBar {
	return darwinStatusBar{inner: macos.SetupStatusBar(title)}
}

// Notify posts a native notification, optionally with action buttons/sound.
func Notify(title, body string, withActions, withSound bool) {
	macos.Notify(title, body, withActions, withSound)
}

// SetNotificationHandler registers the callback for notification buttons.
func SetNotificationHandler(h func(NotifAction)) {
	macos.SetNotificationHandler(func(a macos.NotifAction) { h(NotifAction(a)) })
}

// FrontmostApp returns the bundle id of the frontmost app ("" if unknown).
func FrontmostApp() string { return macos.FrontmostApp() }

// HideSelfAndActivate deactivates Pausa and reactivates the given app.
func HideSelfAndActivate(bundleID string) { macos.HideSelfAndActivate(bundleID) }

// ScreenCount returns the number of attached displays.
func ScreenCount() int { return macos.ScreenCount() }

// SetAccessoryActivationPolicy hides Pausa from the Dock (menu-bar only).
func SetAccessoryActivationPolicy() { macos.SetAccessoryActivationPolicy() }

// SetRegularActivationPolicy shows Pausa in the Dock.
func SetRegularActivationPolicy() { macos.SetRegularActivationPolicy() }

// SetLoginItemEnabled toggles launch-at-login; returns false on failure.
func SetLoginItemEnabled(enabled bool) bool { return macos.SetLoginItemEnabled(enabled) }

// SetOverlayActionHandler registers the callback for overlay buttons.
func SetOverlayActionHandler(h func(OverlayAction)) {
	macos.SetOverlayActionHandler(func(a macos.OverlayAction) { h(OverlayAction(a)) })
}

// CreateOverlays renders the break overlay on the target screen(s).
func CreateOverlays(opts OverlayOptions) {
	macos.CreateOverlays(macos.OverlayOptions{
		Kind:              opts.Kind,
		Title:             opts.Title,
		Timer:             opts.Timer,
		Tip:               opts.Tip,
		HexAccent:         opts.HexAccent,
		ShowActions:       opts.ShowActions,
		Fullscreen:        opts.Fullscreen,
		CurrentScreenOnly: opts.CurrentScreenOnly,
	})
}

// UpdateOverlayTimer updates the countdown shown on the overlay.
func UpdateOverlayTimer(timer string) { macos.UpdateOverlayTimer(timer) }

// CloseOverlays dismisses every break overlay.
func CloseOverlays() { macos.CloseOverlays() }

// NewBusySource returns the macOS busy detector and starts its sampler.
func NewBusySource() BusySource { return macos.NewBusySource() }

// NewIdleSource returns the macOS idle probe.
func NewIdleSource() IdleSource { return macos.IdleSource{} }
