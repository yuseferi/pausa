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

// NotifAction is the user response to an actionable notification.
type NotifAction int

const (
	// NotifDefault is the user clicking the notification body.
	NotifDefault NotifAction = iota + 1
	// NotifSkip is the "Skip" button.
	NotifSkip
	// NotifPostpone is the "Postpone" button.
	NotifPostpone
)

var (
	notifMu      sync.RWMutex
	notifHandler func(NotifAction)
)

// SetNotificationHandler registers a callback invoked whenever the user
// interacts with one of Pausa's notifications. Pass nil to unregister.
// The handler runs on a fresh goroutine so it can do real work safely.
func SetNotificationHandler(h func(NotifAction)) {
	notifMu.Lock()
	notifHandler = h
	notifMu.Unlock()
	C.pausa_notify_request_auth()
}

// Notify posts a notification. If withActions is true, the notification
// includes "Postpone" and "Skip" buttons that route to the registered
// handler. If withSound is true, the system notification sound plays.
func Notify(title, body string, withActions, withSound bool) {
	cT := C.CString(title)
	cB := C.CString(body)
	defer C.free(unsafe.Pointer(cT))
	defer C.free(unsafe.Pointer(cB))
	wa := C.int(0)
	if withActions {
		wa = 1
	}
	ws := C.int(0)
	if withSound {
		ws = 1
	}
	C.pausa_notify_post(cT, cB, wa, ws)
}

//export pausaNotificationAction
func pausaNotificationAction(action C.int) {
	notifMu.RLock()
	h := notifHandler
	notifMu.RUnlock()
	if h == nil {
		return
	}
	go h(NotifAction(action))
}
