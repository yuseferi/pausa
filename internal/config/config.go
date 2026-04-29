// Package config holds the typed application configuration and an atomic
// on-disk store. Configuration is grouped by concern; runtime state (e.g.
// "is the user currently paused") is intentionally NOT persisted here.
package config

import (
	"time"
)

// Config is the user-facing settings for Pausa. All durations are stored as
// time.Duration in memory; on-disk JSON uses string forms (e.g. "10m", "20s")
// via custom marshalling so the file is human-editable.
type Config struct {
	Schedule     ScheduleConfig     `json:"schedule"`
	Notification NotificationConfig `json:"notification"`
	Display      DisplayConfig      `json:"display"`
	WorkingHours WorkingHoursConfig `json:"workingHours"`
	Idle         IdleConfig         `json:"idle"`
	General      GeneralConfig      `json:"general"`
}

// ScheduleConfig controls when breaks fire.
type ScheduleConfig struct {
	// ShortInterval is the time between short breaks.
	ShortInterval Duration `json:"shortInterval"`
	// ShortDuration is how long a short break lasts.
	ShortDuration Duration `json:"shortDuration"`
	// LongEvery is the number of short breaks before a long break replaces one.
	LongEvery int `json:"longEvery"`
	// LongDuration is how long a long break lasts.
	LongDuration Duration `json:"longDuration"`
	// PostponeShort is the delay applied when postponing a short break.
	PostponeShort Duration `json:"postponeShort"`
	// PostponeLong is the delay applied when postponing a long break.
	PostponeLong Duration `json:"postponeLong"`
}

// NotificationConfig controls the pre-break warning notification.
type NotificationConfig struct {
	Enabled        bool     `json:"enabled"`
	WarnShort      Duration `json:"warnShort"` // how long before a short break to warn
	WarnLong       Duration `json:"warnLong"`
	PlaySound      bool     `json:"playSound"`
	ShowActions    bool     `json:"showActions"` // Skip / Postpone buttons in notification
}

// DisplayConfig controls how break screens look.
type DisplayConfig struct {
	Theme            string `json:"theme"` // "system", "light", "dark"
	Fullscreen       bool   `json:"fullscreen"`
	AllMonitors      bool   `json:"allMonitors"`
	ShowExerciseTips bool   `json:"showExerciseTips"`
	ShowBreathing    bool   `json:"showBreathing"` // breathing guide on long breaks
	AccentColor      string `json:"accentColor"`   // hex like "#0ea5e9"
}

// WorkingHoursConfig limits when breaks fire to certain hours/days.
type WorkingHoursConfig struct {
	Enabled bool `json:"enabled"`
	// Days is a 7-element bool slice, index 0 = Sunday.
	Days [7]bool `json:"days"`
	// StartMinute / EndMinute are minutes-since-midnight in local time.
	StartMinute int `json:"startMinute"`
	EndMinute   int `json:"endMinute"`
}

// IdleConfig controls idle-aware behavior.
type IdleConfig struct {
	// PauseWhenIdle pauses the schedule when the system is idle for IdleThreshold.
	PauseWhenIdle bool     `json:"pauseWhenIdle"`
	IdleThreshold Duration `json:"idleThreshold"`
	// NaturalBreaks: if the user is idle for at least the upcoming break's
	// duration before it fires, count it as a "natural break" and reset.
	NaturalBreaks bool `json:"naturalBreaks"`
	// PauseWhenBusy pauses the schedule when the user is detected to be in
	// a meeting (microphone in use) or watching/listening to media. Resumes
	// automatically when those signals clear. Catches Zoom/Meet/Teams
	// calls, YouTube/Netflix playback, etc.
	PauseWhenBusy bool `json:"pauseWhenBusy"`
	// BusyMediaDebounce is how long sustained audio output must continue
	// before Pausa treats it as real media playback. This avoids false
	// positives from short notification sounds. Browser/video URL detection
	// (YouTube, Meet, Netflix, etc.) is immediate and does not wait for the
	// debounce window.
	BusyMediaDebounce Duration `json:"busyMediaDebounce"`
}

// GeneralConfig holds miscellaneous settings.
type GeneralConfig struct {
	StartAtLogin    bool `json:"startAtLogin"`
	ShowInDock      bool `json:"showInDock"`
	StatusBarTitle  bool `json:"statusBarTitle"` // show countdown text in menu bar
}

// Default returns a sensible default configuration.
func Default() Config {
	allWeekdays := [7]bool{false, true, true, true, true, true, false}
	return Config{
		Schedule: ScheduleConfig{
			ShortInterval: Duration(10 * time.Minute),
			ShortDuration: Duration(20 * time.Second),
			LongEvery:     3,
			LongDuration:  Duration(5 * time.Minute),
			PostponeShort: Duration(2 * time.Minute),
			PostponeLong:  Duration(5 * time.Minute),
		},
		Notification: NotificationConfig{
			Enabled:     true,
			WarnShort:   Duration(10 * time.Second),
			WarnLong:    Duration(30 * time.Second),
			PlaySound:   false,
			ShowActions: true,
		},
		Display: DisplayConfig{
			Theme:            "system",
			Fullscreen:       true,
			AllMonitors:      true,
			ShowExerciseTips: true,
			ShowBreathing:    true,
			AccentColor:      "#0ea5e9",
		},
		WorkingHours: WorkingHoursConfig{
			Enabled:     false,
			Days:        allWeekdays,
			StartMinute: 9 * 60,
			EndMinute:   17 * 60,
		},
		Idle: IdleConfig{
			PauseWhenIdle: true,
			IdleThreshold: Duration(2 * time.Minute),
			NaturalBreaks: true,
			PauseWhenBusy: true,
			BusyMediaDebounce: Duration(15 * time.Second),
		},
		General: GeneralConfig{
			StartAtLogin:   false,
			ShowInDock:     false,
			StatusBarTitle: true,
		},
	}
}

// Validate clamps each field into a sane range and applies defaults for
// any obviously-bad values. Returns true if anything was changed.
func (c *Config) Validate() bool {
	d := Default()
	changed := false

	clampDur := func(v *Duration, min, max, def time.Duration) {
		td := time.Duration(*v)
		if td < min || td > max {
			*v = Duration(def)
			changed = true
		}
	}
	clampInt := func(v *int, min, max, def int) {
		if *v < min || *v > max {
			*v = def
			changed = true
		}
	}

	// Schedule
	clampDur(&c.Schedule.ShortInterval, time.Minute, 6*time.Hour, time.Duration(d.Schedule.ShortInterval))
	clampDur(&c.Schedule.ShortDuration, 5*time.Second, 10*time.Minute, time.Duration(d.Schedule.ShortDuration))
	clampInt(&c.Schedule.LongEvery, 1, 20, d.Schedule.LongEvery)
	clampDur(&c.Schedule.LongDuration, 30*time.Second, time.Hour, time.Duration(d.Schedule.LongDuration))
	clampDur(&c.Schedule.PostponeShort, time.Minute, time.Hour, time.Duration(d.Schedule.PostponeShort))
	clampDur(&c.Schedule.PostponeLong, time.Minute, 2*time.Hour, time.Duration(d.Schedule.PostponeLong))

	// Notification
	clampDur(&c.Notification.WarnShort, 0, 5*time.Minute, time.Duration(d.Notification.WarnShort))
	clampDur(&c.Notification.WarnLong, 0, 10*time.Minute, time.Duration(d.Notification.WarnLong))

	// Display
	switch c.Display.Theme {
	case "light", "dark", "system":
	default:
		c.Display.Theme = d.Display.Theme
		changed = true
	}
	if c.Display.AccentColor == "" {
		c.Display.AccentColor = d.Display.AccentColor
		changed = true
	}

	// Working hours
	clampInt(&c.WorkingHours.StartMinute, 0, 24*60-1, d.WorkingHours.StartMinute)
	clampInt(&c.WorkingHours.EndMinute, 0, 24*60, d.WorkingHours.EndMinute)
	if c.WorkingHours.EndMinute <= c.WorkingHours.StartMinute {
		c.WorkingHours.EndMinute = d.WorkingHours.EndMinute
		changed = true
	}

	// Idle
	clampDur(&c.Idle.IdleThreshold, 30*time.Second, time.Hour, time.Duration(d.Idle.IdleThreshold))
	clampDur(&c.Idle.BusyMediaDebounce, 0, 5*time.Minute, time.Duration(d.Idle.BusyMediaDebounce))

	return changed
}

// IsWorkingNow returns true if the given time falls within configured working
// hours, or if working hours are disabled (always true).
func (c *Config) IsWorkingNow(t time.Time) bool {
	if !c.WorkingHours.Enabled {
		return true
	}
	if !c.WorkingHours.Days[int(t.Weekday())] {
		return false
	}
	mins := t.Hour()*60 + t.Minute()
	return mins >= c.WorkingHours.StartMinute && mins < c.WorkingHours.EndMinute
}
