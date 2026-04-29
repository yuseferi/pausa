//go:build darwin

package macos

/*
#include <stdlib.h>
#include "bridge.h"
*/
import "C"
import (
	"sync"
	"unsafe"
)

// ScreenCount returns the number of attached displays.
func ScreenCount() int { return int(C.pausa_screen_count()) }

// BringMainWindowForward moves the Wails main window to the current Space
// and makes it key.
func BringMainWindowForward() {
	C.pausa_window_bring_to_current_space()
}

// SetAccessoryActivationPolicy switches the app to "accessory" mode (no
// Dock icon, can float over fullscreen apps). This must be called before
// any break overlay is shown — without it, macOS won't allow our windows
// to appear on a fullscreen-app Space. Idempotent.
func SetAccessoryActivationPolicy() {
	C.pausa_app_set_accessory()
}

// OverlayAction enumerates the user actions exposed on the overlay UI.
type OverlayAction int

const (
	OverlaySkip     OverlayAction = 1
	OverlayPostpone OverlayAction = 2
)

var (
	overlayMu      sync.RWMutex
	overlayHandler func(OverlayAction)
)

// SetOverlayActionHandler registers a callback invoked when the user
// clicks one of the buttons (Skip / Postpone) inside the native break
// overlay. Pass nil to unregister.
func SetOverlayActionHandler(h func(OverlayAction)) {
	overlayMu.Lock()
	overlayHandler = h
	overlayMu.Unlock()
}

//export pausaOverlayAction
func pausaOverlayAction(action C.int) {
	overlayMu.RLock()
	h := overlayHandler
	overlayMu.RUnlock()
	if h == nil {
		return
	}
	go h(OverlayAction(action))
}

// OverlayOptions controls how break overlays are rendered.
type OverlayOptions struct {
	Kind              string // "short" or "long"
	Title             string // "Short Break" / "Long Break"
	Timer             string // initial countdown text ("00:20")
	Tip               string // exercise tip text
	HexAccent         string // gradient base color (e.g. "#0ea5e9")
	ShowActions       bool   // render Skip / Postpone buttons
	Fullscreen        bool   // true = edge-to-edge; false = ~720x520 centered card
	CurrentScreenOnly bool   // true = overlay only the screen with the cursor
}

// CreateOverlays opens a richly-styled break overlay panel. By default
// every connected screen is overlaid edge-to-edge; use OverlayOptions to
// switch to compact (windowed) or current-screen-only mode.
func CreateOverlays(opts OverlayOptions) {
	cKind := C.CString(opts.Kind)
	cTitle := C.CString(opts.Title)
	cTimer := C.CString(opts.Timer)
	cTip := C.CString(opts.Tip)
	cAccent := C.CString(opts.HexAccent)
	defer C.free(unsafe.Pointer(cKind))
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cTimer))
	defer C.free(unsafe.Pointer(cTip))
	defer C.free(unsafe.Pointer(cAccent))

	sa, fs, co := C.int(0), C.int(0), C.int(0)
	if opts.ShowActions {
		sa = 1
	}
	if opts.Fullscreen {
		fs = 1
	}
	if opts.CurrentScreenOnly {
		co = 1
	}
	C.pausa_overlays_create(cKind, cTitle, cTimer, cTip, cAccent, sa, fs, co)
}

// UpdateOverlayTimer changes the displayed timer on all overlay windows.
func UpdateOverlayTimer(timer string) {
	c := C.CString(timer)
	defer C.free(unsafe.Pointer(c))
	C.pausa_overlays_update_timer(c)
}

// CloseOverlays dismisses any open overlay windows.
func CloseOverlays() { C.pausa_overlays_close() }
