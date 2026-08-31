// Package scheduler is the heart of Pausa. It owns the break state machine
// and runs as a single goroutine (actor pattern); all interaction happens
// through the methods on Scheduler, which post commands to the actor's
// channel. There are no mutexes — concurrency safety follows from "only one
// goroutine touches mutable state".
//
// The scheduler is driven by an injected Clock, so tests can advance time
// deterministically with FakeClock.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"pausa/internal/clock"
	"pausa/internal/config"
)

// IdleSource reports how long the user has been idle. Implementations must
// be safe to call from the scheduler goroutine; on macOS this wraps IOKit.
// A nil IdleSource (or one that always returns 0) disables idle behavior.
type IdleSource interface {
	IdleFor() time.Duration
}

type nopIdle struct{}

func (nopIdle) IdleFor() time.Duration { return 0 }

// BusySource reports whether the user is currently "busy" — in a meeting,
// playing a video, etc. — so the scheduler can pause itself instead of
// firing breaks during interruptable work. Returning ("", false) means
// the user is free. The string is a short label suitable for menu-bar
// display ("in a meeting", "media playing").
//
// Implementations should be cheap (~10ms) since the scheduler polls them
// every BusyPollInterval (default 5s). A nil BusySource disables
// auto-pause.
type BusySource interface {
	BusyState() (label string, busy bool)
}

type nopBusy struct{}

func (nopBusy) BusyState() (string, bool) { return "", false }

// BusyPollInterval is how often the scheduler asks the BusySource whether
// the user is busy. 5s is a good balance: snappy enough that breaks pause
// promptly when a call starts, slow enough not to wake the system from
// idle constantly.
const BusyPollInterval = 5 * time.Second

// Scheduler runs a break schedule. Construct with New, then call Start to
// launch the actor goroutine. Stop or context cancellation shuts it down.
type Scheduler struct {
	clk   clock.Clock
	idle  IdleSource
	busy  BusySource
	cfg   config.Config
	cfgCh <-chan config.Config

	// channels (set in New, never reassigned)
	cmds   chan command
	events chan Event
	done   chan struct{}
}

// New constructs a Scheduler. cfgCh, if non-nil, will be drained inside the
// actor loop so config changes propagate without external synchronization.
// idle may be nil for "never idle". busy may be nil to disable auto-pause.
func New(clk clock.Clock, cfg config.Config, cfgCh <-chan config.Config,
	idle IdleSource, busy BusySource) *Scheduler {
	if idle == nil {
		idle = nopIdle{}
	}
	if busy == nil {
		busy = nopBusy{}
	}
	return &Scheduler{
		clk:    clk,
		idle:   idle,
		busy:   busy,
		cfg:    cfg,
		cfgCh:  cfgCh,
		cmds:   make(chan command, 16),
		events: make(chan Event, 32),
		done:   make(chan struct{}),
	}
}

// Events returns the read side of the event stream. The channel is closed
// when the scheduler stops. Slow consumers will block the actor; the
// channel is buffered (32) to absorb short bursts.
func (s *Scheduler) Events() <-chan Event { return s.events }

// Start launches the actor goroutine. It returns immediately. The actor
// runs until ctx is cancelled or Stop is called.
func (s *Scheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop signals the actor to exit and waits for it to finish.
func (s *Scheduler) Stop() {
	select {
	case s.cmds <- &cmdStop{}:
	case <-s.done:
		return
	}
	<-s.done
}

// ---------- Public API (each method posts a command) ----------

// Snapshot returns the current state. Synchronous.
func (s *Scheduler) Snapshot() Snapshot {
	resp := make(chan Snapshot, 1)
	if !s.send(&cmdSnapshot{resp: resp}) {
		return Snapshot{Phase: PhaseIdle}
	}
	select {
	case snap := <-resp:
		return snap
	case <-s.done:
		return Snapshot{Phase: PhaseIdle}
	}
}

// TakeBreakNow triggers an immediate break.
func (s *Scheduler) TakeBreakNow() { s.send(&cmdTakeBreak{}) }

// EndBreak ends the current break (idempotent).
func (s *Scheduler) EndBreak() { s.send(&cmdEndBreak{reason: endNatural}) }

// SkipBreak skips the current break.
func (s *Scheduler) SkipBreak() { s.send(&cmdEndBreak{reason: endSkip}) }

// PostponeBreak postpones the current or upcoming break.
func (s *Scheduler) PostponeBreak() { s.send(&cmdPostpone{}) }

// Pause pauses the schedule. Ends an in-progress break first.
func (s *Scheduler) Pause() { s.send(&cmdPause{}) }

// Resume resumes the schedule.
func (s *Scheduler) Resume() { s.send(&cmdResume{}) }

// Reset returns the schedule to its initial state.
func (s *Scheduler) Reset() { s.send(&cmdReset{}) }

func (s *Scheduler) send(c command) bool {
	select {
	case s.cmds <- c:
		return true
	case <-s.done:
		return false
	}
}

// ---------- Actor loop ----------

type actorState struct {
	phase           Phase
	nextKind        BreakKind
	nextAt          time.Time
	currentKind     BreakKind
	breakEndsAt     time.Time
	shortsCompleted int
	stats           Stats
	// Timers (always exactly one of breakTimer or notifyTimer is "active" at
	// a time when in PhaseScheduled/PhaseNotifying; both nil otherwise).
	breakTimer  clock.Timer // fires at break start
	notifyTimer clock.Timer // fires at pre-break warning
	tickTimer   clock.Timer // fires every second while OnBreak
	// Auto-pause state.
	// pausedRemaining is the time left on the next-break countdown when we
	// entered PhaseAutoPaused; on resume we re-arm with this remaining time
	// rather than the full interval.
	pausedRemaining time.Duration
	autoPauseReason string
	// pausedByIdle marks an auto-pause caused by user idleness (as opposed
	// to a busy signal). pausedIdleFor is how long the user was already idle
	// when the pause started, and pausedAt when the pause started; together
	// they approximate total away time for natural-break credit on resume.
	pausedByIdle  bool
	pausedIdleFor time.Duration
	pausedAt      time.Time
	// manualPaused remembers an explicit user pause so that a manual "take
	// a break now" during pause doesn't silently re-arm the schedule.
	manualPaused bool
}

func (s *Scheduler) run(ctx context.Context) {
	defer close(s.done)
	defer close(s.events)

	st := &actorState{phase: PhaseIdle, nextKind: BreakShort}
	s.scheduleNext(st)

	// Busy-detection timer. Re-armed each iteration so config changes that
	// disable PauseWhenBusy take effect (we check cfg inside onBusyPoll).
	busyTimer := s.clk.NewTimer(BusyPollInterval)

	for {
		var (
			breakC  <-chan time.Time
			notifyC <-chan time.Time
			tickC   <-chan time.Time
		)
		if st.breakTimer != nil {
			breakC = st.breakTimer.C()
		}
		if st.notifyTimer != nil {
			notifyC = st.notifyTimer.C()
		}
		if st.tickTimer != nil {
			tickC = st.tickTimer.C()
		}

		select {
		case <-ctx.Done():
			s.stopAllTimers(st)
			busyTimer.Stop()
			return

		case cfg, ok := <-s.cfgCh:
			if !ok {
				s.cfgCh = nil
				continue
			}
			s.cfg = cfg
			s.handleConfigChange(st)

		case c := <-s.cmds:
			if _, stop := c.(*cmdStop); stop {
				s.stopAllTimers(st)
				busyTimer.Stop()
				return
			}
			c.apply(s, st)

		case <-notifyC:
			s.onNotifyFire(st)

		case <-breakC:
			s.onBreakFire(st)

		case <-tickC:
			s.onBreakTick(st)

		case <-busyTimer.C():
			s.onBusyPoll(st)
			busyTimer.Reset(BusyPollInterval)
		}
	}
}

// ---------- Commands (typed messages) ----------

type command interface {
	apply(s *Scheduler, st *actorState)
}

type cmdStop struct{}

func (*cmdStop) apply(*Scheduler, *actorState) {}

type cmdSnapshot struct{ resp chan Snapshot }

func (c *cmdSnapshot) apply(s *Scheduler, st *actorState) {
	c.resp <- s.snapshot(st)
}

type cmdTakeBreak struct{}

func (*cmdTakeBreak) apply(s *Scheduler, st *actorState) {
	if st.phase == PhaseOnBreak {
		return
	}
	// Manual "take break now" overrides any auto-pause; user explicitly
	// wants the break despite being marked as busy.
	if st.phase == PhaseAutoPaused {
		st.autoPauseReason = ""
		st.pausedRemaining = 0
		st.pausedByIdle = false
		st.pausedIdleFor = 0
		st.pausedAt = time.Time{}
	}
	s.startBreak(st, st.nextKind)
}

type endReason int

const (
	endNatural endReason = iota
	endSkip
)

type cmdEndBreak struct{ reason endReason }

func (c *cmdEndBreak) apply(s *Scheduler, st *actorState) {
	if st.phase != PhaseOnBreak {
		return
	}
	completed := st.currentKind
	s.stopBreakTimers(st)
	st.phase = PhaseIdle
	st.currentKind = ""
	st.breakEndsAt = time.Time{}

	if c.reason == endSkip {
		st.stats.BreaksSkipped++
		s.publish(EventSkipped, st, withKind(completed))
	} else {
		st.stats.BreaksTaken++
		s.publish(EventBreakEnd, st, withKind(completed))
	}

	s.advanceCounters(st, completed)
	if st.manualPaused {
		st.phase = PhasePaused
		s.publish(EventPaused, st)
		return
	}
	s.scheduleNext(st)
}

type cmdPostpone struct{}

func (*cmdPostpone) apply(s *Scheduler, st *actorState) {
	wasOnBreak := st.phase == PhaseOnBreak
	var kind BreakKind
	if wasOnBreak {
		kind = st.currentKind
	} else {
		kind = st.nextKind
	}

	var delay time.Duration
	if kind == BreakShort {
		delay = s.cfg.Schedule.PostponeShort.AsDuration()
	} else {
		delay = s.cfg.Schedule.PostponeLong.AsDuration()
	}

	if wasOnBreak {
		s.stopBreakTimers(st)
		st.phase = PhaseIdle
		st.currentKind = ""
		st.breakEndsAt = time.Time{}
		s.publish(EventBreakEnd, st, withKind(kind))
	}

	st.stats.BreaksPostponed++
	st.nextKind = kind
	st.nextAt = s.clk.Now().Add(delay)
	// Postponing always lands in PhaseScheduled; drop any auto-pause
	// bookkeeping so stale state can't leak into a future pause/resume.
	st.autoPauseReason = ""
	st.pausedRemaining = 0
	st.pausedByIdle = false
	st.pausedIdleFor = 0
	st.pausedAt = time.Time{}
	s.armBreakTimer(st, delay)
	s.armNotifyTimer(st, delay, kind)
	st.phase = PhaseScheduled
	s.publish(EventPostponed, st)
}

type cmdPause struct{}

func (*cmdPause) apply(s *Scheduler, st *actorState) {
	if st.phase == PhasePaused {
		return
	}
	if st.phase == PhaseOnBreak {
		s.stopBreakTimers(st)
		s.publish(EventBreakEnd, st, withKind(st.currentKind))
		st.currentKind = ""
		st.breakEndsAt = time.Time{}
	}
	// Manual pause overrides auto-pause: clear any preserved remaining
	// time so a future Resume restarts cleanly from a full interval.
	st.autoPauseReason = ""
	st.pausedRemaining = 0
	st.pausedByIdle = false
	st.pausedIdleFor = 0
	st.pausedAt = time.Time{}
	st.manualPaused = true
	s.stopAllTimers(st)
	st.phase = PhasePaused
	s.publish(EventPaused, st)
}

type cmdResume struct{}

func (*cmdResume) apply(s *Scheduler, st *actorState) {
	// Allow Resume to clear both manual pause and auto-pause states.
	if st.phase != PhasePaused && st.phase != PhaseAutoPaused {
		return
	}
	st.autoPauseReason = ""
	st.pausedRemaining = 0
	st.pausedByIdle = false
	st.pausedIdleFor = 0
	st.pausedAt = time.Time{}
	st.manualPaused = false
	st.phase = PhaseIdle
	s.scheduleNext(st)
	s.publish(EventResumed, st)
}

type cmdReset struct{}

func (*cmdReset) apply(s *Scheduler, st *actorState) {
	s.stopAllTimers(st)
	st.shortsCompleted = 0
	st.nextKind = BreakShort
	st.currentKind = ""
	st.breakEndsAt = time.Time{}
	st.manualPaused = false
	st.phase = PhaseIdle
	s.scheduleNext(st)
}

// ---------- Internal helpers (run on actor goroutine only) ----------

func (s *Scheduler) snapshot(st *actorState) Snapshot {
	until := s.cfg.Schedule.LongEvery - st.shortsCompleted
	if until < 0 {
		until = 0
	}
	return Snapshot{
		Phase:           st.phase,
		NextKind:        st.nextKind,
		NextBreakAt:     st.nextAt,
		CurrentKind:     st.currentKind,
		BreakEndsAt:     st.breakEndsAt,
		ShortsCompleted: st.shortsCompleted,
		ShortsUntilLong: until,
		Stats:           st.stats,
		AutoPauseReason: st.autoPauseReason,
	}
}

type eventOpt func(*Event)

func withKind(k BreakKind) eventOpt { return func(e *Event) { e.BreakKind = k } }
func withSecondsLeft(n int) eventOpt {
	return func(e *Event) { e.SecondsLeft = n }
}
func withSecondsUntil(n int) eventOpt {
	return func(e *Event) { e.SecondsUntil = n }
}

func (s *Scheduler) publish(kind EventKind, st *actorState, opts ...eventOpt) {
	ev := Event{Kind: kind, Snapshot: s.snapshot(st)}
	for _, o := range opts {
		o(&ev)
	}
	select {
	case s.events <- ev:
	default:
		// Drop event rather than block; UI will catch up via Snapshot.
	}
}

func (s *Scheduler) scheduleNext(st *actorState) {
	if st.phase == PhasePaused || st.phase == PhaseAutoPaused {
		return
	}
	st.phase = PhaseScheduled
	now := s.clk.Now()
	next := now.Add(s.cfg.Schedule.ShortInterval.AsDuration())
	// Working hours: breaks only fire inside the configured window. If we
	// are currently outside it, or the interval would land outside it,
	// defer to the next window start.
	if s.cfg.WorkingHours.Enabled {
		if !s.cfg.IsWorkingNow(now) || !s.cfg.IsWorkingNow(next) {
			next = s.cfg.NextWorkingStart(now)
		}
	}
	st.nextAt = next
	s.armBreakTimer(st, next.Sub(now))
	s.armNotifyTimer(st, next.Sub(now), st.nextKind)
	s.publish(EventScheduled, st)
}

func (s *Scheduler) armBreakTimer(st *actorState, d time.Duration) {
	if st.breakTimer != nil {
		st.breakTimer.Stop()
	}
	st.breakTimer = s.clk.NewTimer(d)
}

func (s *Scheduler) armNotifyTimer(st *actorState, total time.Duration, kind BreakKind) {
	if st.notifyTimer != nil {
		st.notifyTimer.Stop()
		st.notifyTimer = nil
	}
	if !s.cfg.Notification.Enabled {
		return
	}
	var warn time.Duration
	if kind == BreakShort {
		warn = s.cfg.Notification.WarnShort.AsDuration()
	} else {
		warn = s.cfg.Notification.WarnLong.AsDuration()
	}
	if warn <= 0 || warn >= total {
		return
	}
	st.notifyTimer = s.clk.NewTimer(total - warn)
}

func (s *Scheduler) onNotifyFire(st *actorState) {
	if st.phase != PhaseScheduled {
		return
	}
	st.phase = PhaseNotifying
	st.notifyTimer = nil

	var warn time.Duration
	if st.nextKind == BreakShort {
		warn = s.cfg.Notification.WarnShort.AsDuration()
	} else {
		warn = s.cfg.Notification.WarnLong.AsDuration()
	}
	s.publish(EventNotify, st,
		withKind(st.nextKind),
		withSecondsUntil(int(warn.Seconds())))
}

func (s *Scheduler) onBreakFire(st *actorState) {
	if st.phase != PhaseScheduled && st.phase != PhaseNotifying {
		return
	}
	st.breakTimer = nil
	st.notifyTimer = nil

	// Natural-break detection: if the user has been idle ≥ break duration,
	// consider it taken.
	if s.cfg.Idle.NaturalBreaks {
		dur := s.breakDuration(st.nextKind)
		if s.idle.IdleFor() >= dur {
			st.stats.NaturalBreaks++
			s.publish(EventNatural, st, withKind(st.nextKind))
			s.advanceCounters(st, st.nextKind)
			s.scheduleNext(st)
			return
		}
	}
	s.startBreak(st, st.nextKind)
}

func (s *Scheduler) startBreak(st *actorState, kind BreakKind) {
	dur := s.breakDuration(kind)
	st.phase = PhaseOnBreak
	st.currentKind = kind
	st.breakEndsAt = s.clk.Now().Add(dur)

	if st.breakTimer != nil {
		st.breakTimer.Stop()
	}
	st.breakTimer = s.clk.NewTimer(dur)

	if st.tickTimer != nil {
		st.tickTimer.Stop()
	}
	st.tickTimer = s.clk.NewTimer(time.Second)

	s.publish(EventBreakStart, st, withKind(kind))
}

func (s *Scheduler) onBreakTick(st *actorState) {
	if st.phase != PhaseOnBreak {
		st.tickTimer = nil
		return
	}
	left := int(st.breakEndsAt.Sub(s.clk.Now()).Round(time.Second).Seconds())
	if left < 0 {
		left = 0
	}
	s.publish(EventBreakTick, st,
		withKind(st.currentKind),
		withSecondsLeft(left))

	if left <= 0 {
		// onBreakFire (the *break-end* timer) will run too; stop ticking.
		st.tickTimer.Stop()
		st.tickTimer = nil
		// Break-end is signalled by breakTimer firing again, which calls
		// onBreakFire -- but in OnBreak, breakTimer represents the end of
		// the break. Handle that here:
		s.naturalBreakEnd(st)
		return
	}
	st.tickTimer.Reset(time.Second)
}

// naturalBreakEnd is called when the break duration has elapsed. Equivalent
// to the user letting the break complete naturally.
func (s *Scheduler) naturalBreakEnd(st *actorState) {
	completed := st.currentKind
	s.stopBreakTimers(st)
	st.phase = PhaseIdle
	st.currentKind = ""
	st.breakEndsAt = time.Time{}
	st.stats.BreaksTaken++
	s.publish(EventBreakEnd, st, withKind(completed))
	s.advanceCounters(st, completed)
	if st.manualPaused {
		st.phase = PhasePaused
		s.publish(EventPaused, st)
		return
	}
	s.scheduleNext(st)
}

func (s *Scheduler) advanceCounters(st *actorState, completed BreakKind) {
	if completed == BreakShort {
		st.shortsCompleted++
		if st.shortsCompleted >= s.cfg.Schedule.LongEvery {
			st.nextKind = BreakLong
		} else {
			st.nextKind = BreakShort
		}
	} else { // long
		st.shortsCompleted = 0
		st.nextKind = BreakShort
	}
}

func (s *Scheduler) breakDuration(k BreakKind) time.Duration {
	if k == BreakShort {
		return s.cfg.Schedule.ShortDuration.AsDuration()
	}
	return s.cfg.Schedule.LongDuration.AsDuration()
}

func (s *Scheduler) stopBreakTimers(st *actorState) {
	if st.breakTimer != nil {
		st.breakTimer.Stop()
		st.breakTimer = nil
	}
	if st.tickTimer != nil {
		st.tickTimer.Stop()
		st.tickTimer = nil
	}
}

func (s *Scheduler) stopAllTimers(st *actorState) {
	if st.breakTimer != nil {
		st.breakTimer.Stop()
		st.breakTimer = nil
	}
	if st.notifyTimer != nil {
		st.notifyTimer.Stop()
		st.notifyTimer = nil
	}
	if st.tickTimer != nil {
		st.tickTimer.Stop()
		st.tickTimer = nil
	}
}

func (s *Scheduler) handleConfigChange(st *actorState) {
	switch st.phase {
	case PhaseScheduled, PhaseNotifying:
		// Reschedule the upcoming break with the new interval.
		s.stopAllTimers(st)
		s.scheduleNext(st)
	case PhaseAutoPaused:
		// If the user disabled busy-pause while we were auto-paused,
		// resume immediately. Same for idle-pause with PauseWhenIdle off.
		if (st.pausedByIdle && !s.cfg.Idle.PauseWhenIdle) ||
			(!st.pausedByIdle && !s.cfg.Idle.PauseWhenBusy) {
			s.exitAutoPause(st)
		}
	default:
		// PhaseIdle/Paused/OnBreak: leave timers alone; the next scheduleNext
		// will pick up the new config.
	}
}

// onBusyPoll runs every BusyPollInterval. It transitions us into or out of
// PhaseAutoPaused based on the busy source and user idleness, each gated by
// its own config flag.
func (s *Scheduler) onBusyPoll(st *actorState) {
	// Outside working hours no breaks will fire anyway, so auto-pausing
	// would only produce confusing menu-bar state overnight.
	if s.cfg.WorkingHours.Enabled && !s.cfg.IsWorkingNow(s.clk.Now()) {
		return
	}

	// Only poll the (relatively expensive) busy source when it can change
	// something this tick.
	var label string
	var busy bool
	if s.cfg.Idle.PauseWhenBusy || (st.phase == PhaseAutoPaused && !st.pausedByIdle) {
		label, busy = s.busy.BusyState()
	}
	slog.Debug("scheduler busy poll", "phase", st.phase, "busy", busy, "label", label)

	idleThreshold := s.cfg.Idle.IdleThreshold.AsDuration()

	switch st.phase {
	case PhaseScheduled, PhaseNotifying, PhaseOnBreak:
		switch {
		case s.cfg.Idle.PauseWhenBusy && busy:
			s.enterAutoPause(st, label)
		case st.phase != PhaseOnBreak && s.cfg.Idle.PauseWhenIdle &&
			s.idle.IdleFor() >= idleThreshold:
			// User walked away mid-countdown. Pause so away time isn't
			// charged against the break timer.
			s.enterIdlePause(st)
		}
	case PhaseAutoPaused:
		if st.pausedByIdle {
			// Resume when the user returns (idle drops) or the feature
			// was switched off.
			if !s.cfg.Idle.PauseWhenIdle || s.idle.IdleFor() < idleThreshold {
				s.exitAutoPause(st)
			}
		} else if !busy {
			s.exitAutoPause(st)
		} else if label != st.autoPauseReason {
			// Reason changed (e.g. mic dropped but media still playing);
			// update the visible label without resuming.
			st.autoPauseReason = label
			s.publish(EventAutoPaused, st)
		}
	case PhaseIdle, PhasePaused:
		// Don't auto-pause from these states.
	}
}

// enterAutoPause transitions to PhaseAutoPaused, preserving the time
// remaining on the current countdown so we can resume from where we left
// off.
func (s *Scheduler) enterAutoPause(st *actorState, reason string) {
	slog.Info("entering auto-pause", "fromPhase", st.phase, "reason", reason)
	switch st.phase {
	case PhaseScheduled, PhaseNotifying:
		remaining := st.nextAt.Sub(s.clk.Now())
		if remaining < 0 {
			remaining = 0
		}
		st.pausedRemaining = remaining
	case PhaseOnBreak:
		// User started watching a video mid-break; just end the break and
		// auto-pause. They'll get a new break when the video stops.
		completed := st.currentKind
		s.stopBreakTimers(st)
		st.currentKind = ""
		st.breakEndsAt = time.Time{}
		s.publish(EventBreakEnd, st, withKind(completed))
		// We treat the in-flight break as "taken" since it ran for some
		// portion. Conservative; doesn't double-count.
		st.stats.BreaksTaken++
		s.advanceCounters(st, completed)
		// Reschedule from now as if we just finished a break, so when the
		// busy state clears we have a fresh full-interval countdown.
		st.pausedRemaining = s.cfg.Schedule.ShortInterval.AsDuration()
	default:
		return
	}
	s.stopAllTimers(st)
	st.autoPauseReason = reason
	st.pausedAt = s.clk.Now()
	st.phase = PhaseAutoPaused
	slog.Info("auto-paused", "reason", reason, "remaining", st.pausedRemaining.Round(time.Second).String())
	s.publish(EventAutoPaused, st)
}

// enterIdlePause auto-pauses because the user has been away from the
// keyboard for at least IdleThreshold. Away time may later be credited as
// a natural break (see exitAutoPause).
func (s *Scheduler) enterIdlePause(st *actorState) {
	st.pausedByIdle = true
	st.pausedIdleFor = s.idle.IdleFor()
	s.enterAutoPause(st, "away")
}

// exitAutoPause leaves PhaseAutoPaused. Idle-caused pauses that covered at
// least a full break are credited as natural breaks; otherwise the next
// break is re-armed with the remaining time captured at pause-time.
func (s *Scheduler) exitAutoPause(st *actorState) {
	slog.Info("exiting auto-pause", "reason", st.autoPauseReason, "remaining", st.pausedRemaining.Round(time.Second).String())
	wasIdle := st.pausedByIdle
	idleAtEntry := st.pausedIdleFor
	awayTime := idleAtEntry + s.clk.Now().Sub(st.pausedAt)
	remaining := st.pausedRemaining
	st.autoPauseReason = ""
	st.pausedRemaining = 0
	st.pausedByIdle = false
	st.pausedIdleFor = 0
	st.pausedAt = time.Time{}

	// Working hours may have ended while we were paused; defer everything
	// to the next window (scheduleNext handles the deferral).
	if s.cfg.WorkingHours.Enabled && !s.cfg.IsWorkingNow(s.clk.Now()) {
		st.phase = PhaseIdle
		s.scheduleNext(st)
		s.publish(EventAutoResume, st)
		return
	}

	// Away long enough to cover the upcoming break? Count it as taken.
	if wasIdle && s.cfg.Idle.NaturalBreaks && awayTime >= s.breakDuration(st.nextKind) {
		slog.Info("away time credited as natural break", "away", awayTime.Round(time.Second).String())
		st.stats.NaturalBreaks++
		s.publish(EventNatural, st, withKind(st.nextKind))
		s.advanceCounters(st, st.nextKind)
		s.scheduleNext(st)
		return
	}

	if remaining <= 0 {
		// Nothing was preserved (e.g. paused mid-break); start a fresh
		// full-interval countdown.
		st.phase = PhaseIdle
		s.scheduleNext(st)
		s.publish(EventAutoResume, st)
		return
	}

	// Re-arm a custom-duration scheduling from "now + remaining" without
	// going through scheduleNext (which always uses the full interval).
	st.phase = PhaseScheduled
	st.nextAt = s.clk.Now().Add(remaining)
	s.armBreakTimer(st, remaining)
	s.armNotifyTimer(st, remaining, st.nextKind)
	slog.Info("auto-resumed", "nextBreakIn", remaining.Round(time.Second).String(), "nextKind", st.nextKind)
	s.publish(EventAutoResume, st)
}
