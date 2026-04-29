//go:build darwin

// Package macos is the Pausa macOS native bridge. All AppKit/UN/IOKit work
// happens through cgo into bridge.m. Public Go functions here are safe to
// call from any goroutine; the Objective-C side dispatches onto the main
// queue.
//
// Build flags:
//   -fobjc-arc enables Automatic Reference Counting in bridge.m.
//   -fmodules lets us @import frameworks instead of #import.
//
// Non-darwin builds get the stubs in stubs_other.go so the rest of the
// codebase remains portable.
package macos

/*
#cgo CFLAGS: -fobjc-arc -fmodules -Wno-deprecated-declarations
#cgo LDFLAGS: -framework Cocoa -framework UserNotifications -framework QuartzCore -framework CoreGraphics -framework CoreAudio

#include "bridge.h"
*/
import "C"
