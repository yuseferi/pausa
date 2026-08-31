//go:build darwin

package macos

/*
#include <stdlib.h>
#include "bridge.h"
*/
import "C"
import "unsafe"

// FrontmostApp returns the bundle identifier (preferred) or localized name
// of the currently focused application. Returns "" on failure or if Pausa
// itself is frontmost.
func FrontmostApp() string {
	c := C.pausa_workspace_frontmost_app()
	if c == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(c))
	return C.GoString(c)
}

// ActivateApp brings the given application to the foreground. The
// identifier should be a bundle identifier (e.g. "com.apple.Safari") or a
// localized application name. Returns true on success.
//
// Note: this uses the legacy NSRunningApplication API which can struggle
// with apps that live in a fullscreen Space. For end-of-break restoration,
// prefer HideSelfAndActivate which combines a deactivation step with the
// modern NSWorkspace.openApplicationAtURL: API.
func ActivateApp(identifier string) bool {
	if identifier == "" {
		return false
	}
	c := C.CString(identifier)
	defer C.free(unsafe.Pointer(c))
	return C.pausa_workspace_activate(c) != 0
}

// HideSelfAndActivate is the correct way to return focus to a previous app
// after a break ends. The order is:
//  1. Activate the previous app via NSWorkspace.openApplicationAtURL —
//     this lets macOS switch Spaces to follow the app, which is the only
//     reliable way to return to an app that lives in a fullscreen Space.
//  2. Then hide the Pausa app, removing it from view without forcing
//     another Space switch.
//
// Reversing the order (hide first, activate second) causes macOS to pick
// whatever app is next in the activation stack — usually Finder on
// Desktop 1 — which is wrong when the user was in a fullscreen Safari/IDE.
//
// bundleIdentifier should ideally be a bundle ID like "com.apple.Safari";
// localized names also work as a fallback. Pass "" to just hide Pausa.
func HideSelfAndActivate(bundleIdentifier string) {
	if bundleIdentifier != "" {
		c := C.CString(bundleIdentifier)
		C.pausa_workspace_activate_url(c)
		C.free(unsafe.Pointer(c))
	}
	// Brief delay would be ideal here to let the Space switch animation
	// settle, but Cocoa serializes both calls on the main queue so by the
	// time pausa_app_hide_self runs, openApplicationAtURL has already
	// dispatched its underlying request.
	C.pausa_app_hide_self()
}

// SetLoginItemEnabled registers or unregisters Pausa as a login item
// (macOS 13+, bundled app only). Returns true on success.
func SetLoginItemEnabled(enabled bool) bool {
	v := C.int(0)
	if enabled {
		v = 1
	}
	return C.pausa_login_item_set_enabled(v) != 0
}
