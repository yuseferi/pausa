//go:build !darwin

package platform

import "time"

// This file provides no-op implementations for platforms Pausa does not yet
// support natively. They exist so the module cross-compiles and so the core
// (scheduler, config, UI) can run headless in tests. Windows and Linux will
// replace these with real implementations.

type nopStatusBar struct{}

func (nopStatusBar) SetTitle(string)                 {}
func (nopStatusBar) SetBuiltinIcon(StatusIconState)  {}
func (nopStatusBar) SetImage([]byte)                 {}
func (nopStatusBar) AddItem(MenuTag, string, func()) {}
func (nopStatusBar) AddDisabled(MenuTag, string)     {}
func (nopStatusBar) AddSeparator()                   {}
func (nopStatusBar) UpdateItem(MenuTag, string)      {}
func (nopStatusBar) SetItemHidden(MenuTag, bool)     {}
func (nopStatusBar) Teardown()                       {}

// SetupStatusBar returns a no-op status bar.
func SetupStatusBar(string) StatusBar { return nopStatusBar{} }

// Notify is a no-op.
func Notify(string, string, bool, bool) {}

// SetNotificationHandler is a no-op.
func SetNotificationHandler(func(NotifAction)) {}

// FrontmostApp always reports unknown.
func FrontmostApp() string { return "" }

// HideSelfAndActivate is a no-op.
func HideSelfAndActivate(string) {}

// ScreenCount reports a single display.
func ScreenCount() int { return 1 }

// SetAccessoryActivationPolicy is a no-op.
func SetAccessoryActivationPolicy() {}

// SetRegularActivationPolicy is a no-op.
func SetRegularActivationPolicy() {}

// SetLoginItemEnabled always fails on unsupported platforms.
func SetLoginItemEnabled(bool) bool { return false }

// SetOverlayActionHandler is a no-op.
func SetOverlayActionHandler(func(OverlayAction)) {}

// CreateOverlays is a no-op.
func CreateOverlays(OverlayOptions) {}

// UpdateOverlayTimer is a no-op.
func UpdateOverlayTimer(string) {}

// CloseOverlays is a no-op.
func CloseOverlays() {}

type nopBusy struct{}

func (nopBusy) BusyState() (string, bool) { return "", false }
func (nopBusy) Close()                    {}

// NewBusySource returns a source that never reports busy.
func NewBusySource() BusySource { return nopBusy{} }

type nopIdle struct{}

func (nopIdle) IdleFor() time.Duration { return 0 }

// NewIdleSource returns a source that never reports idle.
func NewIdleSource() IdleSource { return nopIdle{} }
