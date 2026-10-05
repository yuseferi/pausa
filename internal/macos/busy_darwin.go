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
//  3. frontmost-browser URL heuristic (muted YouTube/Meet/Netflix tabs)
//
// Output-device activity (kAudioDevicePropertyDeviceIsRunningSomewhere) is
// deliberately NOT used: it reports whether any process holds the output
// device open, which apps like Spotube keep doing while idle, so it produced
// false "media playing" readings. The Now Playing API reflects actual
// playback and is the reliable signal.
type BusySource struct {
	mu sync.Mutex

	// cached reading served to BusyState(). Updated by the sampler
	// goroutine so the expensive native/AppleScript detection never runs on
	// a caller's goroutine (notably the scheduler actor).
	cachedLabel string
	cachedBusy  bool

	// change-only logging latch
	havePrev    bool
	prevMic     bool
	prevNowPlay bool
	prevBrowser bool
	prevReason  string
	prevSince   time.Time

	// Background sampler lifecycle.
	startOnce sync.Once
	closeOnce sync.Once
	cancel    context.CancelFunc
	done      chan struct{}
}

// busySampleInterval is how often the background sampler refreshes the
// cached reading. It is shorter than the scheduler's BusyPollInterval so a
// poll never sees a reading older than one poll interval.
const busySampleInterval = 2 * time.Second

// NewBusySource returns a configured BusySource and starts its background
// sampler. Call Close to stop the sampler.
func NewBusySource() *BusySource {
	bs := &BusySource{}
	bs.start()
	return bs
}

// start launches the background sampler exactly once.
func (b *BusySource) start() {
	b.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		b.cancel = cancel
		b.done = make(chan struct{})
		go b.loop(ctx)
	})
}

// loop periodically samples busy signals on its own goroutine.
func (b *BusySource) loop(ctx context.Context) {
	defer close(b.done)
	// Sample immediately so the first scheduler poll has data.
	b.sample()
	t := time.NewTicker(busySampleInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			b.sample()
		}
	}
}

// Close stops the background sampler. Safe to call multiple times and from
// any goroutine.
func (b *BusySource) Close() {
	b.closeOnce.Do(func() {
		if b.cancel != nil {
			b.cancel()
			<-b.done
		}
	})
}

// BusyState returns the most recent sampled busy label and whether the user
// is busy. It is a cheap, non-blocking read: the expensive native and
// AppleScript detection happens on the sampler goroutine, so the scheduler
// actor never stalls waiting on IOKit/osascript.
func (b *BusySource) BusyState() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cachedLabel, b.cachedBusy
}

// sample performs one detection pass and refreshes the cached reading.
func (b *BusySource) sample() {
	// Native detectors.
	mic := C.pausa_busy_microphone_active() != 0
	nowPlaying := C.pausa_busy_now_playing_active() != 0

	// Browser heuristic: catches muted YouTube / Meet tabs when the browser is
	// the frontmost app. This is intentionally conservative — we don't inspect
	// background tabs for privacy and reliability reasons.
	browserMedia, browserURL := frontmostBrowserMediaCandidate()
	browserHost := hostOnly(browserURL)

	media := nowPlaying || browserMedia

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
		"browserMedia", browserMedia,
		"browserHost", browserHost,
		"combined", reason)

	b.logChange(mic, nowPlaying, browserMedia, reason, browserHost)

	b.mu.Lock()
	b.cachedLabel = reason
	b.cachedBusy = state.IsBusy()
	b.mu.Unlock()
}

func (b *BusySource) logChange(mic, nowPlaying, browserMedia bool, reason, browserHost string) {
	b.mu.Lock()
	now := time.Now()
	if b.havePrev &&
		b.prevMic == mic &&
		b.prevNowPlay == nowPlaying &&
		b.prevBrowser == browserMedia &&
		b.prevReason == reason {
		b.mu.Unlock()
		return
	}
	attrs := []any{
		"mic", mic,
		"nowPlaying", nowPlaying,
		"browserMedia", browserMedia,
		"browserHost", browserHost,
		"label", reason,
	}
	hadPrev := b.havePrev
	if hadPrev {
		attrs = append(attrs, "prevHeldFor", now.Sub(b.prevSince).Round(time.Second).String())
	}
	b.havePrev = true
	b.prevMic = mic
	b.prevNowPlay = nowPlaying
	b.prevBrowser = browserMedia
	b.prevReason = reason
	b.prevSince = now
	b.mu.Unlock()

	// Log outside the lock so a slow log sink can never stall BusyState()
	// (and therefore the scheduler actor) on the cache lock.
	if hadPrev {
		slog.Info("busy state changed", attrs...)
	} else {
		slog.Info("busy state initial", attrs...)
	}
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
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
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
	return strings.TrimPrefix(strings.ToLower(u.Host), "www.")
}
