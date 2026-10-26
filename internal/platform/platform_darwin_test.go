//go:build darwin

package platform

import (
	"testing"

	"pausa/internal/macos"
)

// The darwin adapter casts between platform.* and macos.* enum values, so the
// two sets must stay numerically identical. If the bridge ever renumbers its
// constants, this test fails instead of silently misrouting a click or action.
func TestEnumsMatchBridge(t *testing.T) {
	cases := []struct {
		name      string
		got, want int
	}{
		{"StatusIconRunning", int(StatusIconRunning), int(macos.StatusIconRunning)},
		{"StatusIconPaused", int(StatusIconPaused), int(macos.StatusIconPaused)},
		{"StatusIconBusy", int(StatusIconBusy), int(macos.StatusIconBusy)},
		{"StatusIconBreak", int(StatusIconBreak), int(macos.StatusIconBreak)},

		{"NotifDefault", int(NotifDefault), int(macos.NotifDefault)},
		{"NotifSkip", int(NotifSkip), int(macos.NotifSkip)},
		{"NotifPostpone", int(NotifPostpone), int(macos.NotifPostpone)},

		{"OverlaySkip", int(OverlaySkip), int(macos.OverlaySkip)},
		{"OverlayPostpone", int(OverlayPostpone), int(macos.OverlayPostpone)},
	}

	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, but the bridge has %d", c.name, c.got, c.want)
		}
	}
}

// breakapp decides whether to show the Wails window for a break based on
// NativeOverlays, and whether closing the window may hide it based on
// SystemTray. macOS must advertise both, or breaks would render twice (once
// in a panel and once in the window) and the close button would quit.
func TestDarwinCapabilities(t *testing.T) {
	if !Caps.NativeOverlays {
		t.Error("NativeOverlays must be true on darwin")
	}
	if !Caps.SystemTray {
		t.Error("SystemTray must be true on darwin")
	}
}

// The preferences UI hides a control when its capability is false. If macOS
// ever advertised a capability it cannot honour, users would lose a setting
// they can actually use; this pins the full native set.
func TestDarwinSupportsEveryCapability(t *testing.T) {
	all := map[string]bool{
		"NativeOverlays": Caps.NativeOverlays,
		"SystemTray":     Caps.SystemTray,
		"DockIcon":       Caps.DockIcon,
		"LoginItem":      Caps.LoginItem,
		"BusyDetection":  Caps.BusyDetection,
		"IdleDetection":  Caps.IdleDetection,
	}
	for name, ok := range all {
		if !ok {
			t.Errorf("%s must be true on darwin", name)
		}
	}
}
