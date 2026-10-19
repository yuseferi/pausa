package platform

import "testing"

// The per-OS files assert interface conformance at compile time. Here we check
// the value invariants the app relies on: callers switch on these constants, so
// they must be distinct and stable.
func TestStatusIconValuesAreDistinct(t *testing.T) {
	icons := []StatusIconState{
		StatusIconRunning,
		StatusIconPaused,
		StatusIconBusy,
		StatusIconBreak,
	}
	seen := make(map[StatusIconState]bool, len(icons))
	for _, v := range icons {
		if seen[v] {
			t.Fatalf("duplicate StatusIconState value %d", v)
		}
		seen[v] = true
	}
}

func TestNotificationAndOverlayActionsAreDistinct(t *testing.T) {
	if NotifDefault == NotifSkip || NotifSkip == NotifPostpone || NotifDefault == NotifPostpone {
		t.Fatal("notification actions must be distinct")
	}
	if OverlaySkip == OverlayPostpone {
		t.Fatal("overlay actions must be distinct")
	}
}
