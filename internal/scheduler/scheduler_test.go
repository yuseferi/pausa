package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"pausa/internal/clock"
	"pausa/internal/config"
)

// testHarness wires up a scheduler with a fake clock and a drainable event
// channel. All durations are kept short so test math stays obvious.
type testHarness struct {
	t      *testing.T
	clk    *clock.FakeClock
	cfg    config.Config
	cfgCh  chan config.Config
	idle   *fakeIdle
	busy   *fakeBusy
	sched  *Scheduler
	cancel context.CancelFunc
	events []Event
}

type fakeIdle struct{ d time.Duration }

func (f *fakeIdle) IdleFor() time.Duration { return f.d }

type fakeBusy struct {
	mu    sync.Mutex
	label string
	busy  bool
}

func (f *fakeBusy) BusyState() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.label, f.busy
}

func (f *fakeBusy) set(busy bool, label string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.busy = busy
	f.label = label
}

func newHarness(t *testing.T) *testHarness {
	return newHarnessAt(t, time.Unix(1_700_000_000, 0), nil)
}

// newHarnessAt is like newHarness but anchors the fake clock at a specific
// instant and lets the test tweak the config before the scheduler starts.
func newHarnessAt(t *testing.T, start time.Time, modify func(*config.Config)) *testHarness {
	t.Helper()
	cfg := config.Default()
	// Squash everything to seconds for fast tests
	cfg.Schedule.ShortInterval = config.Duration(10 * time.Second)
	cfg.Schedule.ShortDuration = config.Duration(2 * time.Second)
	cfg.Schedule.LongDuration = config.Duration(5 * time.Second)
	cfg.Schedule.LongEvery = 3
	cfg.Schedule.PostponeShort = config.Duration(3 * time.Second)
	cfg.Schedule.PostponeLong = config.Duration(4 * time.Second)
	cfg.Notification.WarnShort = config.Duration(2 * time.Second)
	cfg.Notification.WarnLong = config.Duration(2 * time.Second)
	cfg.Idle.NaturalBreaks = false
	cfg.Idle.PauseWhenBusy = true
	if modify != nil {
		modify(&cfg)
	}

	h := &testHarness{
		t:     t,
		clk:   clock.NewFake(start),
		cfg:   cfg,
		cfgCh: make(chan config.Config, 1),
		idle:  &fakeIdle{},
		busy:  &fakeBusy{},
	}
	h.sched = New(h.clk, cfg, h.cfgCh, h.idle, h.busy)
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.sched.Start(ctx)
	t.Cleanup(func() {
		cancel()
		h.sched.Stop()
	})
	// drain the initial EventScheduled
	h.drain()
	return h
}

// drain pulls all events currently buffered on the channel into h.events.
// It also yields briefly so the actor can react to clock advances.
func (h *testHarness) drain() {
	// Yield repeatedly: the actor may emit events in cascades (e.g.
	// notify -> break -> tick).
	for i := 0; i < 50; i++ {
		select {
		case ev := <-h.sched.Events():
			h.events = append(h.events, ev)
		case <-time.After(5 * time.Millisecond):
			return
		}
	}
}

// advance moves the fake clock forward and drains resulting events.
func (h *testHarness) advance(d time.Duration) {
	h.clk.Advance(d)
	h.drain()
}

// last returns the most recent event, or fails if there are none.
func (h *testHarness) last() Event {
	h.t.Helper()
	if len(h.events) == 0 {
		h.t.Fatal("no events captured")
	}
	return h.events[len(h.events)-1]
}

// kinds returns the EventKinds in order, useful for asserting sequences.
func (h *testHarness) kinds() []EventKind {
	out := make([]EventKind, len(h.events))
	for i, e := range h.events {
		out[i] = e.Kind
	}
	return out
}

func (h *testHarness) clearEvents() { h.events = nil }

// ---------- Tests ----------

func TestInitialScheduleIsShortBreak(t *testing.T) {
	h := newHarness(t)
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled", snap.Phase)
	}
	if snap.NextKind != BreakShort {
		t.Errorf("nextKind=%s, want short", snap.NextKind)
	}
	if snap.NextBreakAt.Sub(h.clk.Now()) != 10*time.Second {
		t.Errorf("nextBreakAt=%v, want +10s", snap.NextBreakAt.Sub(h.clk.Now()))
	}
}

func TestNotifyFiresBeforeBreak(t *testing.T) {
	h := newHarness(t)
	h.advance(8 * time.Second) // notify at T-2s of 10s interval
	if h.last().Kind != EventNotify {
		t.Fatalf("expected EventNotify, got %s", h.last().Kind)
	}
	if h.last().SecondsUntil != 2 {
		t.Errorf("secondsUntil=%d, want 2", h.last().SecondsUntil)
	}
}

func TestBreakStartsAfterInterval(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second)
	// We should see notify + breakStart + initial tick
	saw := h.kinds()
	hasStart := false
	for _, k := range saw {
		if k == EventBreakStart {
			hasStart = true
			break
		}
	}
	if !hasStart {
		t.Fatalf("no breakStart in %v", saw)
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseOnBreak {
		t.Errorf("phase=%s, want on_break", snap.Phase)
	}
}

func TestBreakEndsNaturallyAfterDuration(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second) // start break
	h.clearEvents()
	h.advance(2 * time.Second) // duration is 2s
	saw := h.kinds()
	hasEnd := false
	for _, k := range saw {
		if k == EventBreakEnd {
			hasEnd = true
			break
		}
	}
	if !hasEnd {
		t.Fatalf("expected EventBreakEnd in %v", saw)
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled (next break already armed)", snap.Phase)
	}
	if snap.Stats.BreaksTaken != 1 {
		t.Errorf("BreaksTaken=%d, want 1", snap.Stats.BreaksTaken)
	}
}

func TestThirdShortPromotesToLong(t *testing.T) {
	h := newHarness(t)
	// Complete 3 short breaks (LongEvery=3). After the 3rd, next should be long.
	for i := 0; i < 3; i++ {
		h.advance(10 * time.Second) // break starts
		h.advance(2 * time.Second)  // break ends
	}
	snap := h.sched.Snapshot()
	if snap.NextKind != BreakLong {
		t.Errorf("after 3 shorts, NextKind=%s, want long", snap.NextKind)
	}
	if snap.Stats.BreaksTaken != 3 {
		t.Errorf("BreaksTaken=%d, want 3", snap.Stats.BreaksTaken)
	}
}

func TestLongBreakResetsCounter(t *testing.T) {
	h := newHarness(t)
	// 3 shorts to promote
	for i := 0; i < 3; i++ {
		h.advance(10 * time.Second)
		h.advance(2 * time.Second)
	}
	// Now run the long break (5s duration)
	h.advance(10 * time.Second)
	snap := h.sched.Snapshot()
	if snap.CurrentKind != BreakLong {
		t.Fatalf("on-break kind=%s, want long", snap.CurrentKind)
	}
	h.advance(5 * time.Second)
	snap = h.sched.Snapshot()
	if snap.NextKind != BreakShort {
		t.Errorf("after long, NextKind=%s, want short", snap.NextKind)
	}
	if snap.ShortsCompleted != 0 {
		t.Errorf("ShortsCompleted=%d, want 0", snap.ShortsCompleted)
	}
}

func TestSkipBreak(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second) // start break
	h.clearEvents()
	h.sched.SkipBreak()
	h.drain()
	saw := h.kinds()
	hasSkip := false
	for _, k := range saw {
		if k == EventSkipped {
			hasSkip = true
		}
	}
	if !hasSkip {
		t.Fatalf("expected EventSkipped in %v", saw)
	}
	snap := h.sched.Snapshot()
	if snap.Stats.BreaksSkipped != 1 {
		t.Errorf("BreaksSkipped=%d, want 1", snap.Stats.BreaksSkipped)
	}
	if snap.Stats.BreaksTaken != 0 {
		t.Errorf("BreaksTaken=%d, want 0", snap.Stats.BreaksTaken)
	}
}

func TestPostponeFromScheduled(t *testing.T) {
	h := newHarness(t)
	h.sched.PostponeBreak() // postpone before break starts
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled", snap.Phase)
	}
	delta := snap.NextBreakAt.Sub(h.clk.Now())
	if delta != 3*time.Second {
		t.Errorf("postponed by %v, want 3s", delta)
	}
	if snap.Stats.BreaksPostponed != 1 {
		t.Errorf("BreaksPostponed=%d, want 1", snap.Stats.BreaksPostponed)
	}
}

func TestPostponeFromOnBreak(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second) // start break
	h.sched.PostponeBreak()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled", snap.Phase)
	}
	if snap.Stats.BreaksPostponed != 1 {
		t.Errorf("BreaksPostponed=%d, want 1", snap.Stats.BreaksPostponed)
	}
}

func TestPauseEndsOngoingBreak(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second) // on break
	h.sched.Pause()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhasePaused {
		t.Fatalf("phase=%s, want paused", snap.Phase)
	}
	// Advancing time while paused should produce no new events
	h.clearEvents()
	h.advance(60 * time.Second)
	if len(h.events) > 0 {
		t.Errorf("events fired while paused: %v", h.kinds())
	}
}

func TestResumeReschedules(t *testing.T) {
	h := newHarness(t)
	h.sched.Pause()
	h.drain()
	h.sched.Resume()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled", snap.Phase)
	}
}

func TestTakeBreakNowFromIdle(t *testing.T) {
	h := newHarness(t)
	h.sched.TakeBreakNow()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseOnBreak {
		t.Errorf("phase=%s, want on_break", snap.Phase)
	}
	if snap.CurrentKind != BreakShort {
		t.Errorf("kind=%s, want short", snap.CurrentKind)
	}
}

func TestNaturalBreakSkipsScheduledBreak(t *testing.T) {
	h := newHarness(t)
	h.cfg.Idle.NaturalBreaks = true
	h.cfgCh <- h.cfg
	h.drain()

	// Pretend the user has been idle for 30s
	h.idle.d = 30 * time.Second
	h.advance(10 * time.Second)
	saw := h.kinds()
	hasNatural := false
	hasStart := false
	for _, k := range saw {
		if k == EventNatural {
			hasNatural = true
		}
		if k == EventBreakStart {
			hasStart = true
		}
	}
	if !hasNatural {
		t.Errorf("expected EventNatural in %v", saw)
	}
	if hasStart {
		t.Errorf("expected NO EventBreakStart in %v (was satisfied by idle)", saw)
	}
	snap := h.sched.Snapshot()
	if snap.Stats.NaturalBreaks != 1 {
		t.Errorf("NaturalBreaks=%d, want 1", snap.Stats.NaturalBreaks)
	}
}

func TestConfigChangeReschedulesPendingBreak(t *testing.T) {
	h := newHarness(t)
	h.advance(2 * time.Second) // 8s left
	// Change interval to something else
	h.cfg.Schedule.ShortInterval = config.Duration(20 * time.Second)
	h.cfgCh <- h.cfg
	h.drain()
	snap := h.sched.Snapshot()
	delta := snap.NextBreakAt.Sub(h.clk.Now())
	if delta != 20*time.Second {
		t.Errorf("after config change delta=%v, want 20s", delta)
	}
}

func TestResetReturnsToShortFromAnyState(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 2; i++ {
		h.advance(10 * time.Second)
		h.advance(2 * time.Second)
	}
	if got := h.sched.Snapshot().ShortsCompleted; got != 2 {
		t.Fatalf("setup: ShortsCompleted=%d, want 2", got)
	}
	h.sched.Reset()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.NextKind != BreakShort {
		t.Errorf("after reset NextKind=%s, want short", snap.NextKind)
	}
	if snap.ShortsCompleted != 0 {
		t.Errorf("after reset ShortsCompleted=%d, want 0", snap.ShortsCompleted)
	}
}

func TestEventBreakTickReportsCountdown(t *testing.T) {
	h := newHarness(t)
	h.advance(10 * time.Second) // start break (2s duration)
	h.clearEvents()
	h.advance(time.Second)
	// Find a tick event with secondsLeft <= 1
	foundTick := false
	for _, e := range h.events {
		if e.Kind == EventBreakTick {
			foundTick = true
			if e.SecondsLeft < 0 || e.SecondsLeft > 2 {
				t.Errorf("secondsLeft=%d out of range", e.SecondsLeft)
			}
		}
	}
	if !foundTick {
		t.Errorf("expected at least one EventBreakTick in %v", h.kinds())
	}
}

func TestNotifyDisabledDoesNotEmitNotifyEvent(t *testing.T) {
	h := newHarness(t)
	h.cfg.Notification.Enabled = false
	h.cfgCh <- h.cfg
	h.drain()
	h.clearEvents()
	h.advance(10 * time.Second)
	for _, k := range h.kinds() {
		if k == EventNotify {
			t.Errorf("EventNotify fired despite notifications disabled: %v", h.kinds())
		}
	}
}

// ---------- Busy / auto-pause tests ----------

// hasKind reports whether the captured event stream contains the given
// event kind.
func (h *testHarness) hasKind(k EventKind) bool {
	for _, e := range h.events {
		if e.Kind == k {
			return true
		}
	}
	return false
}

func TestBusySignalAutoPausesScheduledBreak(t *testing.T) {
	h := newHarness(t)
	// User starts a meeting before the break interval elapses.
	h.busy.set(true, "in a meeting")

	// The busy poll fires every 5s. Advance enough to trigger a poll.
	h.advance(BusyPollInterval + time.Second)

	if !h.hasKind(EventAutoPaused) {
		t.Fatalf("expected EventAutoPaused in %v", h.kinds())
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseAutoPaused {
		t.Errorf("phase=%s, want auto_paused", snap.Phase)
	}
	if snap.AutoPauseReason != "in a meeting" {
		t.Errorf("autoPauseReason=%q, want %q", snap.AutoPauseReason, "in a meeting")
	}
}

func TestBusySignalDoesNotFireBreakWhilePaused(t *testing.T) {
	h := newHarness(t)
	h.busy.set(true, "")
	h.advance(BusyPollInterval + time.Second) // enter auto-pause
	h.clearEvents()

	// Advance well past when the break would have fired.
	h.advance(60 * time.Second)
	if h.hasKind(EventBreakStart) {
		t.Errorf("break fired during auto-pause: %v", h.kinds())
	}
}

func TestBusyClearedResumesWithRemainingTime(t *testing.T) {
	// This test verifies the *invariant* that auto-pause preserves the
	// remaining time on the countdown. Exact numbers depend on when the
	// busy poll happens relative to the break interval; we just assert
	// the resume time is meaningfully shorter than a full interval but
	// non-trivial.
	h := newHarness(t)
	// Get well into the interval before triggering busy.
	h.advance(2 * time.Second) // 8s remaining of 10s
	h.busy.set(true, "")
	// Trigger the busy poll (next fires at the next BusyPollInterval
	// boundary). Some interval has elapsed since the actor saw the start.
	h.advance(BusyPollInterval)
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("expected auto-paused after busy poll, got %s",
			h.sched.Snapshot().Phase)
	}

	// Stay busy for a long time; no break should fire.
	h.advance(60 * time.Second)
	if h.hasKind(EventBreakStart) {
		t.Errorf("break fired during long busy stretch: %v", h.kinds())
	}

	// Clear busy; the next busy poll should auto-resume.
	h.busy.set(false, "")
	h.clearEvents()
	h.advance(BusyPollInterval + time.Second)
	if !h.hasKind(EventAutoResume) {
		t.Fatalf("expected EventAutoResume in %v", h.kinds())
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Fatalf("phase=%s, want scheduled", snap.Phase)
	}
	// The remaining time should be POSITIVE and SHORTER than a fresh
	// full-interval (10s). We don't assert the exact value because it
	// depends on FakeClock's batch-firing semantics relative to the
	// actor goroutine's scheduling.
	delta := snap.NextBreakAt.Sub(h.clk.Now())
	if delta <= 0 {
		t.Errorf("after resume, time-to-break=%v, want > 0", delta)
	}
	if delta >= h.cfg.Schedule.ShortInterval.AsDuration() {
		t.Errorf("after resume, time-to-break=%v, should be less than full interval %v",
			delta, h.cfg.Schedule.ShortInterval.AsDuration())
	}
}

func TestManualPauseOverridesAutoPause(t *testing.T) {
	h := newHarness(t)
	h.busy.set(true, "")
	h.advance(BusyPollInterval + time.Second) // enter auto-pause
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("setup: expected auto-paused")
	}

	h.sched.Pause()
	h.drain()
	if h.sched.Snapshot().Phase != PhasePaused {
		t.Errorf("phase=%s, want paused (manual)", h.sched.Snapshot().Phase)
	}

	// Now clear busy; manual pause should remain — no auto-resume.
	h.busy.set(false, "")
	h.clearEvents()
	h.advance(BusyPollInterval + time.Second)
	if h.hasKind(EventAutoResume) {
		t.Errorf("EventAutoResume fired despite manual pause")
	}
	if h.sched.Snapshot().Phase != PhasePaused {
		t.Errorf("phase changed away from manual pause: %s", h.sched.Snapshot().Phase)
	}
}

func TestTakeBreakNowOverridesAutoPause(t *testing.T) {
	h := newHarness(t)
	h.busy.set(true, "")
	h.advance(BusyPollInterval + time.Second)
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("setup: expected auto-paused")
	}

	h.sched.TakeBreakNow()
	h.drain()
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseOnBreak {
		t.Errorf("phase=%s, want on_break", snap.Phase)
	}
}

func TestBusyDisabledInConfigSkipsAutoPause(t *testing.T) {
	h := newHarness(t)
	h.cfg.Idle.PauseWhenBusy = false
	h.cfgCh <- h.cfg
	h.drain()

	h.busy.set(true, "")
	h.clearEvents()
	h.advance(BusyPollInterval + time.Second)
	if h.hasKind(EventAutoPaused) {
		t.Errorf("EventAutoPaused fired despite PauseWhenBusy=false")
	}
}

func TestConfigDisableBusyExitsAutoPause(t *testing.T) {
	h := newHarness(t)
	h.busy.set(true, "")
	h.advance(BusyPollInterval + time.Second)
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("setup: expected auto-paused")
	}

	h.cfg.Idle.PauseWhenBusy = false
	h.cfgCh <- h.cfg
	h.drain()

	if h.sched.Snapshot().Phase == PhaseAutoPaused {
		t.Errorf("still auto-paused after PauseWhenBusy disabled")
	}
}

func TestBusyReasonChangeUpdatesLabel(t *testing.T) {
	h := newHarness(t)
	h.busy.set(true, "in a meeting")
	h.advance(BusyPollInterval + time.Second)
	if got := h.sched.Snapshot().AutoPauseReason; got != "in a meeting" {
		t.Fatalf("initial label=%q", got)
	}

	h.busy.set(true, "media playing")
	h.advance(BusyPollInterval + time.Second)
	if got := h.sched.Snapshot().AutoPauseReason; got != "media playing" {
		t.Errorf("after change, label=%q, want media playing", got)
	}
}

// ---------- Working-hours tests ----------

// whConfig enables a Mon–Fri 09:00–17:00 working window.
func whConfig(c *config.Config) {
	c.WorkingHours.Enabled = true
	c.WorkingHours.Days = [7]bool{false, true, true, true, true, true, false}
	c.WorkingHours.StartMinute = 9 * 60
	c.WorkingHours.EndMinute = 17 * 60
}

func TestWorkingHoursInsideWindowSchedulesNormally(t *testing.T) {
	// Wednesday 2025-06-04 10:00 local — inside the window.
	start := time.Date(2025, 6, 4, 10, 0, 0, 0, time.Local)
	h := newHarnessAt(t, start, whConfig)
	snap := h.sched.Snapshot()
	if delta := snap.NextBreakAt.Sub(h.clk.Now()); delta != 10*time.Second {
		t.Errorf("next break in %v, want 10s (normal interval inside working hours)", delta)
	}
}

func TestWorkingHoursDefersBreakOutsideWindow(t *testing.T) {
	// Wednesday 2025-06-04 22:00 local — outside the window.
	start := time.Date(2025, 6, 4, 22, 0, 0, 0, time.Local)
	h := newHarnessAt(t, start, whConfig)
	snap := h.sched.Snapshot()
	want := time.Date(2025, 6, 5, 9, 0, 0, 0, time.Local)
	if !snap.NextBreakAt.Equal(want) {
		t.Errorf("nextBreakAt=%v, want %v (next working-window start)", snap.NextBreakAt, want)
	}
}

func TestWorkingHoursPushesBreakLandingOutsideWindow(t *testing.T) {
	// Wednesday 16:59:55 — the 10s interval would land at 17:00:05, one
	// tick past the window end. Expect deferral to Thursday 09:00.
	start := time.Date(2025, 6, 4, 16, 59, 55, 0, time.Local)
	h := newHarnessAt(t, start, whConfig)
	snap := h.sched.Snapshot()
	want := time.Date(2025, 6, 5, 9, 0, 0, 0, time.Local)
	if !snap.NextBreakAt.Equal(want) {
		t.Errorf("nextBreakAt=%v, want %v", snap.NextBreakAt, want)
	}
}

func TestWorkingHoursNoBreakFiresOutsideWindow(t *testing.T) {
	start := time.Date(2025, 6, 4, 22, 0, 0, 0, time.Local)
	h := newHarnessAt(t, start, whConfig)
	h.clearEvents()
	h.advance(4 * time.Hour) // deep into the night — nothing should fire
	if h.hasKind(EventBreakStart) {
		t.Errorf("break fired outside working hours: %v", h.kinds())
	}
}

// ---------- Idle auto-pause tests ----------

func TestIdlePausesCountdown(t *testing.T) {
	h := newHarness(t)
	h.idle.d = 3 * time.Minute // >= IdleThreshold (2m default)
	h.advance(BusyPollInterval + time.Second)

	if !h.hasKind(EventAutoPaused) {
		t.Fatalf("expected EventAutoPaused in %v", h.kinds())
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseAutoPaused {
		t.Errorf("phase=%s, want auto_paused", snap.Phase)
	}
	if snap.AutoPauseReason != "away" {
		t.Errorf("autoPauseReason=%q, want %q", snap.AutoPauseReason, "away")
	}
}

func TestIdleResumeKeepsRemainingTime(t *testing.T) {
	h := newHarness(t)
	h.advance(2 * time.Second) // 8s left of the interval
	h.idle.d = 3 * time.Minute
	h.advance(BusyPollInterval) // poll pauses the countdown
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("setup: expected auto-paused")
	}

	h.idle.d = 0 // user returns
	h.clearEvents()
	h.advance(BusyPollInterval) // next poll resumes

	if !h.hasKind(EventAutoResume) {
		t.Fatalf("expected EventAutoResume in %v", h.kinds())
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Fatalf("phase=%s, want scheduled", snap.Phase)
	}
	delta := snap.NextBreakAt.Sub(h.clk.Now())
	if delta <= 0 || delta >= h.cfg.Schedule.ShortInterval.AsDuration() {
		t.Errorf("after idle-resume, time-to-break=%v, want (0, full interval)", delta)
	}
}

func TestIdleAwayLongEnoughCreditsNaturalBreak(t *testing.T) {
	h := newHarness(t)
	h.cfg.Idle.NaturalBreaks = true
	h.cfgCh <- h.cfg
	h.drain()

	h.idle.d = 3 * time.Minute
	h.advance(BusyPollInterval + time.Second) // pause as "away"
	h.idle.d = 0                              // user returns
	h.clearEvents()
	h.advance(BusyPollInterval)

	if !h.hasKind(EventNatural) {
		t.Fatalf("expected EventNatural (away time covers the break) in %v", h.kinds())
	}
	if h.hasKind(EventBreakStart) {
		t.Errorf("break started despite away-time credit: %v", h.kinds())
	}
	if got := h.sched.Snapshot().Stats.NaturalBreaks; got != 1 {
		t.Errorf("NaturalBreaks=%d, want 1", got)
	}
}

func TestIdlePauseDisabledKeepsCounting(t *testing.T) {
	h := newHarness(t)
	h.cfg.Idle.PauseWhenIdle = false
	h.cfgCh <- h.cfg
	h.drain()

	h.idle.d = 3 * time.Minute
	h.clearEvents()
	h.advance(BusyPollInterval + time.Second)
	if h.hasKind(EventAutoPaused) {
		t.Errorf("EventAutoPaused fired despite PauseWhenIdle=false")
	}
}

func TestAutoPauseSettlesAfterWorkingHoursEnd(t *testing.T) {
	// Wednesday 16:59:50 — busy pause starts inside the window but the
	// signal clears only after closing time. The scheduler must settle the
	// pause and defer the next break to Thursday 09:00, not stay
	// auto-paused until the next window.
	start := time.Date(2025, 6, 4, 16, 59, 50, 0, time.Local)
	h := newHarnessAt(t, start, whConfig)

	h.busy.set(true, "in a meeting")
	h.advance(BusyPollInterval)
	if h.sched.Snapshot().Phase != PhaseAutoPaused {
		t.Fatalf("setup: expected auto-paused, got %s", h.sched.Snapshot().Phase)
	}

	h.busy.set(false, "")
	h.clearEvents()
	h.advance(BusyPollInterval) // now past 17:00

	if !h.hasKind(EventAutoResume) {
		t.Fatalf("expected EventAutoResume after busy cleared post-closing, got %v", h.kinds())
	}
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseScheduled {
		t.Fatalf("phase=%s, want scheduled", snap.Phase)
	}
	want := time.Date(2025, 6, 5, 9, 0, 0, 0, time.Local)
	if !snap.NextBreakAt.Equal(want) {
		t.Errorf("nextBreakAt=%v, want %v", snap.NextBreakAt, want)
	}
}

func TestIdlePauseConvertsToBusyPause(t *testing.T) {
	h := newHarness(t)
	h.idle.d = 3 * time.Minute
	h.advance(BusyPollInterval) // idle-pause kicks in
	snap := h.sched.Snapshot()
	if snap.Phase != PhaseAutoPaused || snap.AutoPauseReason != "away" {
		t.Fatalf("setup: phase=%s reason=%q, want away idle-pause", snap.Phase, snap.AutoPauseReason)
	}

	// User returns but is immediately in a meeting.
	h.idle.d = 0
	h.busy.set(true, "in a meeting")
	h.clearEvents()
	h.advance(BusyPollInterval)

	snap = h.sched.Snapshot()
	if snap.Phase != PhaseAutoPaused {
		t.Fatalf("phase=%s, want still auto-paused (converted to busy)", snap.Phase)
	}
	if snap.AutoPauseReason != "in a meeting" {
		t.Errorf("reason=%q, want converted busy label", snap.AutoPauseReason)
	}
	if h.hasKind(EventAutoResume) {
		t.Errorf("resumed mid-conversion: %v", h.kinds())
	}

	// Meeting ends; countdown resumes.
	h.busy.set(false, "")
	h.clearEvents()
	h.advance(BusyPollInterval)
	if !h.hasKind(EventAutoResume) {
		t.Fatalf("expected EventAutoResume after meeting ended, got %v", h.kinds())
	}
	if got := h.sched.Snapshot().Phase; got != PhaseScheduled {
		t.Errorf("phase=%s, want scheduled", got)
	}
}

// ---------- Misc regression tests ----------

func TestManualPauseSurvivesManualBreak(t *testing.T) {
	h := newHarness(t)
	h.sched.Pause()
	h.drain()
	h.sched.TakeBreakNow()
	h.drain()
	if got := h.sched.Snapshot().Phase; got != PhaseOnBreak {
		t.Fatalf("phase=%s, want on_break", got)
	}

	h.clearEvents()
	h.advance(2 * time.Second) // let the break finish

	snap := h.sched.Snapshot()
	if snap.Phase != PhasePaused {
		t.Errorf("phase=%s, want paused (manual pause must survive a manual break)", snap.Phase)
	}
	if !h.hasKind(EventPaused) {
		t.Errorf("expected EventPaused in %v", h.kinds())
	}

	// And the schedule must stay dormant afterwards.
	h.clearEvents()
	h.advance(60 * time.Second)
	if h.hasKind(EventBreakStart) {
		t.Errorf("break fired after manually-paused manual break: %v", h.kinds())
	}
	if got := h.sched.Snapshot().Phase; got != PhasePaused {
		t.Errorf("phase drifted to %s, want paused", got)
	}
}

func TestSnapshotReportsShortsUntilLong(t *testing.T) {
	h := newHarness(t)
	if got := h.sched.Snapshot().ShortsUntilLong; got != 3 {
		t.Errorf("initial ShortsUntilLong=%d, want 3", got)
	}
	h.advance(10 * time.Second) // break starts
	h.advance(2 * time.Second)  // break ends
	if got := h.sched.Snapshot().ShortsUntilLong; got != 2 {
		t.Errorf("after one short, ShortsUntilLong=%d, want 2", got)
	}
}
