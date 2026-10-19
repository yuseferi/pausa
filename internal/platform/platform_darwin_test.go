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
