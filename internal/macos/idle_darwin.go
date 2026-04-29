//go:build darwin

package macos

/*
#include "bridge.h"
*/
import "C"
import "time"

// IdleSource returns the duration since the user's last input event.
// Implements scheduler.IdleSource.
type IdleSource struct{}

// IdleFor returns how long the user has been idle (no keyboard or mouse
// input), or 0 on failure.
func (IdleSource) IdleFor() time.Duration {
	secs := float64(C.pausa_idle_seconds())
	if secs <= 0 {
		return 0
	}
	return time.Duration(secs * float64(time.Second))
}
