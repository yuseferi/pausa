//go:build !darwin

package platform

import (
	"context"
	"log/slog"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// This file provides the baseline implementation for platforms without a
// native bridge. On Windows and Linux the app still runs — the Wails window is
// the whole UI, and notifications go through the Wails notification service
// (D-Bus on Linux, WinRT/toast on Windows) — but tray, busy detection, idle
// detection and multi-screen overlays are unavailable. Other GOOS values fall
// back to the same no-ops so the module keeps cross-compiling.

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

// InitNotifications opens the platform notification channel. On non-macOS this
// is the Wails notification service, which must be initialized before the
// first send (on Linux it holds the D-Bus connection).
func InitNotifications(ctx context.Context) error {
	return wailsruntime.InitializeNotifications(ctx)
}

// CloseNotifications releases the notification channel.
func CloseNotifications(ctx context.Context) {
	wailsruntime.CleanupNotifications(ctx)
}

// Notify posts a desktop notification. Action buttons and notification
// actions are macOS-only, so withActions and withSound are ignored here.
func Notify(ctx context.Context, title, body string, withActions, withSound bool) {
	if err := wailsruntime.SendNotification(ctx, wailsruntime.NotificationOptions{
		ID:    "pausa-break-warning",
		Title: title,
		Body:  body,
	}); err != nil {
		slog.Warn("desktop notification failed", "err", err)
	}
}

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
