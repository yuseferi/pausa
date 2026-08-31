package scheduler

import "time"

// Phase is the high-level state of the scheduler.
type Phase string

const (
	// PhaseIdle is the initial state, before scheduling has started.
	PhaseIdle Phase = "idle"
	// PhaseScheduled means a future break is timed.
	PhaseScheduled Phase = "scheduled"
	// PhaseNotifying means the pre-break warning has been emitted; the break
	// itself is imminent.
	PhaseNotifying Phase = "notifying"
	// PhaseOnBreak means a break is currently underway.
	PhaseOnBreak Phase = "on_break"
	// PhasePaused means the user manually paused breaks. Resumes only on
	// an explicit Resume command.
	PhasePaused Phase = "paused"
	// PhaseAutoPaused means the scheduler paused itself because a busy
	// signal (microphone in use, media playing, ...) is active. Resumes
	// automatically when the busy signal clears.
	PhaseAutoPaused Phase = "auto_paused"
)

// BreakKind distinguishes short and long breaks.
type BreakKind string

const (
	BreakShort BreakKind = "short"
	BreakLong  BreakKind = "long"
)

// Snapshot is an immutable view of the scheduler's current state, suitable
// for sending to the UI. Time fields are absolute wall-clock instants.
type Snapshot struct {
	Phase           Phase     `json:"phase"`
	NextKind        BreakKind `json:"nextKind"`
	NextBreakAt     time.Time `json:"nextBreakAt"`
	CurrentKind     BreakKind `json:"currentKind"`
	BreakEndsAt     time.Time `json:"breakEndsAt"`
	ShortsCompleted int       `json:"shortsCompleted"` // since last long break
	ShortsUntilLong int       `json:"shortsUntilLong"` // remaining shorts before next long
	Stats           Stats     `json:"stats"`
	// AutoPauseReason is a human-readable label for why the scheduler is
	// auto-paused (e.g. "in a meeting"). Empty when not auto-paused.
	AutoPauseReason string `json:"autoPauseReason,omitempty"`
}

// Stats counts user behavior over the current process lifetime.
type Stats struct {
	BreaksTaken     int `json:"breaksTaken"`
	BreaksSkipped   int `json:"breaksSkipped"`
	BreaksPostponed int `json:"breaksPostponed"`
	NaturalBreaks   int `json:"naturalBreaks"`
}

// EventKind enumerates UI-relevant scheduler events.
type EventKind string

const (
	EventScheduled  EventKind = "scheduled"  // a new break is scheduled
	EventNotify     EventKind = "notify"     // pre-break warning
	EventBreakStart EventKind = "breakStart" // break has begun
	EventBreakTick  EventKind = "breakTick"  // 1Hz tick during a break
	EventBreakEnd   EventKind = "breakEnd"   // break completed normally
	EventSkipped    EventKind = "skipped"
	EventPostponed  EventKind = "postponed"
	EventPaused     EventKind = "paused"     // user-initiated pause
	EventResumed    EventKind = "resumed"    // user-initiated resume
	EventAutoPaused EventKind = "autoPaused" // busy signal triggered pause
	EventAutoResume EventKind = "autoResume" // busy signal cleared
	EventNatural    EventKind = "natural"    // a break was satisfied by idle time
)

// Event is published on the scheduler's Events channel.
type Event struct {
	Kind     EventKind `json:"kind"`
	Snapshot Snapshot  `json:"snapshot"`
	// BreakKind is set on break-related events.
	BreakKind BreakKind `json:"breakKind,omitempty"`
	// SecondsLeft is set on EventBreakTick.
	SecondsLeft int `json:"secondsLeft,omitempty"`
	// SecondsUntil is set on EventNotify.
	SecondsUntil int `json:"secondsUntil,omitempty"`
}
