// Package clock provides a Clock interface so time-dependent code can be
// tested deterministically with a FakeClock. Production code uses System.
package clock

import (
	"sync"
	"time"
)

// Clock abstracts the parts of the time package the scheduler uses.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
	// NewTimer creates a timer that fires after d. If d <= 0, the timer
	// fires (almost) immediately.
	NewTimer(d time.Duration) Timer
}

// Timer abstracts time.Timer so the fake implementation can drive it
// deterministically. Stop is idempotent.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// ---------- System (production) ----------

// System is the real wall-clock implementation.
type System struct{}

// New returns the system clock.
func New() Clock { return System{} }

// Now returns the current wall time.
func (System) Now() time.Time { return time.Now() }

// NewTimer wraps time.NewTimer.
func (System) NewTimer(d time.Duration) Timer {
	if d <= 0 {
		d = time.Nanosecond
	}
	return &systemTimer{t: time.NewTimer(d)}
}

type systemTimer struct{ t *time.Timer }

func (s *systemTimer) C() <-chan time.Time { return s.t.C }
func (s *systemTimer) Stop() bool          { return s.t.Stop() }
func (s *systemTimer) Reset(d time.Duration) bool {
	if !s.t.Stop() {
		select {
		case <-s.t.C:
		default:
		}
	}
	return s.t.Reset(d)
}

// ---------- Fake (test) ----------

// FakeClock is a deterministic clock for tests. It maintains a virtual time
// that only advances via Advance; timers fire synchronously when their
// deadline is reached.
type FakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

// NewFake returns a FakeClock anchored at the given start time.
func NewFake(start time.Time) *FakeClock {
	return &FakeClock{now: start}
}

// Now returns the virtual time.
func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// NewTimer creates a fake timer scheduled relative to virtual now.
func (f *FakeClock) NewTimer(d time.Duration) Timer {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := &fakeTimer{
		c:        make(chan time.Time, 1),
		deadline: f.now.Add(d),
		clock:    f,
		active:   true,
	}
	f.timers = append(f.timers, t)
	if d <= 0 {
		t.fireLocked(f.now)
	}
	return t
}

// Advance moves the virtual clock forward by d, firing any timers whose
// deadline falls within (oldNow, oldNow+d]. Timers fire synchronously and
// in deadline order; the test goroutine should yield (e.g. read from a
// channel) between Advance calls if it expects scheduler reactions.
func (f *FakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	target := f.now.Add(d)
	for {
		var next *fakeTimer
		for _, t := range f.timers {
			if !t.active {
				continue
			}
			if !t.deadline.After(target) {
				if next == nil || t.deadline.Before(next.deadline) {
					next = t
				}
			}
		}
		if next == nil {
			break
		}
		f.now = next.deadline
		next.fireLocked(f.now)
	}
	f.now = target
	f.mu.Unlock()
}

type fakeTimer struct {
	c        chan time.Time
	deadline time.Time
	clock    *FakeClock
	active   bool
}

func (t *fakeTimer) C() <-chan time.Time { return t.c }

func (t *fakeTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	wasActive := t.active
	t.active = false
	return wasActive
}

func (t *fakeTimer) Reset(d time.Duration) bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	wasActive := t.active
	t.active = true
	t.deadline = t.clock.now.Add(d)
	// drain a previously-fired value
	select {
	case <-t.c:
	default:
	}
	return wasActive
}

// fireLocked must be called with f.mu held.
func (t *fakeTimer) fireLocked(now time.Time) {
	t.active = false
	select {
	case t.c <- now:
	default:
	}
}
