//go:build darwin

package macos

/*
#include "bridge.h"
*/
import "C"

import (
	"context"
	"log/slog"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// BusyKind enumerates the reasons the user is considered "busy" and breaks
// should be auto-paused. Multiple reasons can be true simultaneously.
type BusyKind uint8

const (
	BusyMicrophone   BusyKind = 1 << iota // a meeting / call / dictation
	BusyMediaPlaying                      // YouTube / Spotify / video tab playing
)

// BusyState reports which busy signals are currently active. Returns 0
// when the user is free (no busy signal active).
type BusyState uint8

// IsBusy reports whether any busy signal is active.
func (s BusyState) IsBusy() bool { return s != 0 }

// Has reports whether a particular kind is set.
func (s BusyState) Has(k BusyKind) bool { return uint8(s)&uint8(k) != 0 }

// String returns a short human label for status display.
func (s BusyState) String() string {
	if s == 0 {
		return ""
	}
	switch {
	case s.Has(BusyMicrophone) && s.Has(BusyMediaPlaying):
		return "in call · media"
	case s.Has(BusyMicrophone):
		return "in a meeting"
	case s.Has(BusyMediaPlaying):
		return "media playing"
	}
	return ""
}

// BusySource implements scheduler.BusySource by polling a few macOS-native
// signals:
//  1. microphone in use (strong meeting/call signal)
//  2. system Now Playing API (when browsers/apps publish there)
//  3. sustained output-device activity (fallback for apps that don't)
//  4. frontmost-browser URL heuristic (muted YouTube/Meet/Netflix tabs)
//
// The output-device activity is debounced to avoid short notification sounds
// from auto-pausing the schedule.
type BusySource struct {
	mu sync.Mutex
	// outputActiveSince is non-zero while the output-device active signal has
	// been continuously true. Once it has been true for >= mediaDebounce, we
	// treat it as real media playback.
	outputActiveSince time.Time
	mediaDebounce     time.Duration

	// change-only logging latch
	havePrev    bool
	prevMic     bool
	prevNowPlay bool
	prevOutput  bool
	prevBrowser bool
	prevReason  string
	prevSince   time.Time
}

// DefaultMediaDebounce is the output-audio fallback debounce window.
const DefaultMediaDebounce = 15 * time.Second

// NewBusySource returns a configured BusySource.
func NewBusySource(mediaDebounce time.Duration) *BusySource {
	bs := &BusySource{}
	bs.SetMediaDebounce(mediaDebounce)
	return bs
}

// SetMediaDebounce updates the audio-output debounce duration. The browser
// URL heuristic and the Now Playing signal remain immediate; only the raw
// output-device fallback is delayed by this amount.
func (b *BusySource) SetMediaDebounce(d time.Duration) {
	if d < 0 {
		d = 0
	}
	b.mu.Lock()
	b.mediaDebounce = d
	b.mu.Unlock()
}

// BusyState returns the current busy label and whether the user is busy.
func (b *BusySource) BusyState() (string, bool) {
	// Native detectors.
	mic := C.pausa_busy_microphone_active() != 0
	nowPlaying := C.pausa_busy_now_playing_active() != 0
	rawOutput := C.pausa_busy_output_active() != 0

	// Browser heuristic: catches muted YouTube / Meet tabs when the browser is
	// the frontmost app. This is intentionally conservative — we don't inspect
	// background tabs for privacy and reliability reasons.
	browserMedia, browserURL := frontmostBrowserMediaCandidate()
	browserHost := hostOnly(browserURL)

	media := nowPlaying || b.debouncedOutputActive(rawOutput) || browserMedia

	var state BusyState
	if mic {
		state |= BusyState(BusyMicrophone)
	}
	if media {
		state |= BusyState(BusyMediaPlaying)
	}
	reason := state.String()

	// Per-poll detail (verbose; enable with PAUSA_LOG_LEVEL=debug).
	slog.Debug("busy poll",
		"mic", mic,
		"nowPlaying", nowPlaying,
		"output", rawOutput,
		"browserMedia", browserMedia,
		"browserHost", browserHost,
		"combined", reason)

	b.logChange(mic, nowPlaying, rawOutput, browserMedia, reason, browserHost)

	if !state.IsBusy() {
		return "", false
	}
	return reason, true
}

func (b *BusySource) debouncedOutputActive(rawOutput bool) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rawOutput {
		if b.outputActiveSince.IsZero() {
			b.outputActiveSince = time.Now()
		}
		if b.mediaDebounce == 0 {
			return true
		}
		return time.Since(b.outputActiveSince) >= b.mediaDebounce
	}
	b.outputActiveSince = time.Time{}
	return false
}

func (b *BusySource) logChange(mic, nowPlaying, rawOutput, browserMedia bool, reason, browserHost string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if b.havePrev &&
		b.prevMic == mic &&
		b.prevNowPlay == nowPlaying &&
		b.prevOutput == rawOutput &&
		b.prevBrowser == browserMedia &&
		b.prevReason == reason {
		return
	}
	attrs := []any{
		"mic", mic,
		"nowPlaying", nowPlaying,
		"output", rawOutput,
		"browserMedia", browserMedia,
		"browserHost", browserHost,
		"label", reason,
	}
	if b.havePrev {
		attrs = append(attrs, "prevHeldFor", now.Sub(b.prevSince).Round(time.Second).String())
		slog.Info("busy state changed", attrs...)
	} else {
		slog.Info("busy state initial", attrs...)
	}
	b.havePrev = true
	b.prevMic = mic
	b.prevNowPlay = nowPlaying
	b.prevOutput = rawOutput
	b.prevBrowser = browserMedia
	b.prevReason = reason
	b.prevSince = now
}

// frontmostBrowserMediaCandidate returns (true, url) if the frontmost app is
// a supported browser and its active tab URL looks like a video/meeting page.
// This is deliberately conservative and frontmost-only.
func frontmostBrowserMediaCandidate() (bool, string) {
	front := FrontmostApp()
	switch front {
	case "com.google.Chrome", "Google Chrome":
		url := activeChromeLikeURL("Google Chrome")
		return mediaURL(url), url
	case "company.thebrowser.Browser", "Arc":
		url := activeChromeLikeURL("Arc")
		return mediaURL(url), url
	case "com.apple.Safari", "Safari":
		url := activeSafariURL()
		return mediaURL(url), url
	case "com.brave.Browser", "Brave Browser":
		url := activeChromeLikeURL("Brave Browser")
		return mediaURL(url), url
	default:
		return false, ""
	}
}

func activeChromeLikeURL(app string) string {
	script := `tell application "` + app + `"
    if not (exists front window) then return ""
    try
        return URL of active tab of front window
    on error
        return ""
    end try
end tell`
	return runAppleScript(script)
}

func activeSafariURL() string {
	script := `tell application "Safari"
    if not (exists front window) then return ""
    try
        return URL of current tab of front window
    on error
        return ""
    end try
end tell`
	return runAppleScript(script)
}

func runAppleScript(script string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func mediaURL(raw string) bool {
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	path := strings.ToLower(u.Path)
	switch host {
	case "youtube.com", "m.youtube.com", "music.youtube.com", "youtu.be", "tv.youtube.com":
		return true
	case "meet.google.com":
		return true
	case "netflix.com", "vimeo.com", "player.vimeo.com", "twitch.tv", "player.twitch.tv", "disneyplus.com", "hulu.com", "primevideo.com":
		return true
	case "loom.com":
		return strings.Contains(path, "/share/") || strings.Contains(path, "/embed/")
	}
	return false
}

func hostOnly(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(u.Host, "www."))
}
