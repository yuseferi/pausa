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

// MenuTag identifies a status-bar menu item. Negative or zero tags are
// reserved.
type MenuTag int

// StatusIconState controls the built-in native vector icon shown in the
// macOS menu bar.
type StatusIconState int

const (
	StatusIconRunning StatusIconState = iota
	StatusIconPaused
	StatusIconBusy
	StatusIconBreak
)

// StatusBar provides a typed wrapper over the native menu-bar item. Click
// handlers are invoked on the cgo callback goroutine; if they need to do
// real work, they should post to a channel.
type StatusBar struct {
	mu       sync.RWMutex
	handlers map[MenuTag]func()
}

var (
	statusBar    *StatusBar
	statusBarMu  sync.RWMutex
	statusBarSet bool
)

// SetupStatusBar initializes the menu-bar item with the given title (e.g.
// an emoji or short label). Returns the StatusBar handle. Idempotent — a
// second call replaces the click-handler registry but leaves the existing
// items in place.
func SetupStatusBar(title string) *StatusBar {
	sb := &StatusBar{handlers: make(map[MenuTag]func())}

	statusBarMu.Lock()
	statusBar = sb
	statusBarSet = true
	statusBarMu.Unlock()

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.pausa_statusbar_init(cTitle)
	return sb
}

// SetTitle replaces the menu-bar title.
func (s *StatusBar) SetTitle(title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.pausa_statusbar_set_title(cTitle)
}

// SetBuiltinIcon swaps the menu-bar icon to one of the native vector
// template glyphs (auto-tinted by macOS for light/dark mode).
func (s *StatusBar) SetBuiltinIcon(state StatusIconState) {
	C.pausa_statusbar_set_builtin_icon(C.int(state))
}

// SetImage replaces the menu-bar icon with the given PNG bytes. The image
// is rendered as a template (auto-tinted by macOS for light/dark mode).
// Pass an empty slice to clear the image.
func (s *StatusBar) SetImage(pngData []byte) {
	if len(pngData) == 0 {
		C.pausa_statusbar_set_image_data(nil, 0)
		return
	}
	C.pausa_statusbar_set_image_data(unsafe.Pointer(&pngData[0]), C.int(len(pngData)))
}

// AddItem appends a clickable menu item. Use AddDisabled for informational
// rows like a countdown display.
func (s *StatusBar) AddItem(tag MenuTag, title string, onClick func()) {
	s.mu.Lock()
	if onClick != nil {
		s.handlers[tag] = onClick
	}
	s.mu.Unlock()

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.pausa_statusbar_add_item(cTitle, C.int(tag), 1)
}

// AddDisabled appends a disabled menu item (no click handler).
func (s *StatusBar) AddDisabled(tag MenuTag, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.pausa_statusbar_add_item(cTitle, C.int(tag), 0)
}

// AddSeparator appends a divider.
func (s *StatusBar) AddSeparator() { C.pausa_statusbar_add_separator() }

// UpdateItem changes the title of an existing item.
func (s *StatusBar) UpdateItem(tag MenuTag, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.pausa_statusbar_update_item(C.int(tag), cTitle)
}

// SetItemHidden toggles visibility on an existing item.
func (s *StatusBar) SetItemHidden(tag MenuTag, hidden bool) {
	v := C.int(0)
	if hidden {
		v = 1
	}
	C.pausa_statusbar_set_item_hidden(C.int(tag), v)
}

// Teardown removes the menu-bar item entirely.
func (s *StatusBar) Teardown() {
	statusBarMu.Lock()
	statusBar = nil
	statusBarSet = false
	statusBarMu.Unlock()
	C.pausa_statusbar_remove()
}

//export pausaStatusBarClicked
func pausaStatusBarClicked(tag C.int) {
	statusBarMu.RLock()
	sb := statusBar
	statusBarMu.RUnlock()
	if sb == nil {
		return
	}
	sb.mu.RLock()
	h := sb.handlers[MenuTag(tag)]
	sb.mu.RUnlock()
	if h != nil {
		// Run the handler on a fresh goroutine so the callback returns
		// quickly to the AppKit main queue.
		go h()
	}
}
