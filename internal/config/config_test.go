package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultIsValid(t *testing.T) {
	c := Default()
	if c.Validate() {
		t.Fatal("Default() should already be valid")
	}
}

func TestValidateClampsOutOfRange(t *testing.T) {
	c := Default()
	c.Schedule.ShortInterval = Duration(-1 * time.Second)
	c.Schedule.LongEvery = 0
	c.Display.Theme = "neon"
	if !c.Validate() {
		t.Fatal("Validate should report changes")
	}
	if c.Schedule.ShortInterval <= 0 {
		t.Errorf("ShortInterval not clamped: %v", c.Schedule.ShortInterval)
	}
	if c.Schedule.LongEvery < 1 {
		t.Errorf("LongEvery not clamped: %d", c.Schedule.LongEvery)
	}
	if c.Display.Theme != "system" {
		t.Errorf("Theme not reset: %s", c.Display.Theme)
	}
}

func TestDurationRoundTrip(t *testing.T) {
	type wrap struct{ D Duration }
	in := wrap{D: Duration(90 * time.Second)}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out wrap
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.D != in.D {
		t.Errorf("round-trip mismatch: got %v want %v", out.D, in.D)
	}
}

func TestStoreAtomicWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Exists() {
		t.Fatal("expected !Exists on first open")
	}
	c := s.Get()
	c.Schedule.ShortInterval = Duration(7 * time.Minute)
	if _, err := s.Set(c); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s2.Exists() {
		t.Fatal("expected Exists after Set")
	}
	if s2.Get().Schedule.ShortInterval != Duration(7*time.Minute) {
		t.Errorf("did not persist: %v", s2.Get().Schedule.ShortInterval)
	}
}

func TestStoreSubscribeBroadcasts(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(filepath.Join(dir, "cfg.json"))
	ch := s.Subscribe()

	c := s.Get()
	c.Display.Theme = "dark"
	if _, err := s.Set(c); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-ch:
		if got.Display.Theme != "dark" {
			t.Errorf("subscriber got wrong theme: %s", got.Display.Theme)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive update")
	}
}

func TestIsWorkingNow(t *testing.T) {
	c := Default()
	c.WorkingHours.Enabled = true
	c.WorkingHours.Days = [7]bool{false, true, true, true, true, true, false}
	c.WorkingHours.StartMinute = 9 * 60
	c.WorkingHours.EndMinute = 17 * 60

	// Wednesday at 10:00
	wed10 := time.Date(2024, 1, 3, 10, 0, 0, 0, time.Local)
	if !c.IsWorkingNow(wed10) {
		t.Error("expected working at Wed 10:00")
	}
	// Wednesday at 18:00
	wed18 := time.Date(2024, 1, 3, 18, 0, 0, 0, time.Local)
	if c.IsWorkingNow(wed18) {
		t.Error("expected NOT working at Wed 18:00")
	}
	// Sunday
	sun := time.Date(2024, 1, 7, 10, 0, 0, 0, time.Local)
	if c.IsWorkingNow(sun) {
		t.Error("expected NOT working on Sunday")
	}

	// Disabled => always working
	c.WorkingHours.Enabled = false
	if !c.IsWorkingNow(sun) {
		t.Error("expected always-working when disabled")
	}
}
