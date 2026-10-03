package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
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

func TestValidateRepairsInvertedWorkingHours(t *testing.T) {
	c := Default()
	c.WorkingHours.Enabled = true
	c.WorkingHours.StartMinute = 20 * 60 // 20:00
	c.WorkingHours.EndMinute = 0         // midnight
	if !c.Validate() {
		t.Fatal("Validate should report the repaired window")
	}
	if c.WorkingHours.EndMinute <= c.WorkingHours.StartMinute {
		t.Fatalf("window still inverted: start=%d end=%d",
			c.WorkingHours.StartMinute, c.WorkingHours.EndMinute)
	}
	// A time inside the repaired window must be reported as working;
	// before the fix this could never be true and all breaks stopped.
	inside := time.Date(2024, 1, 3, c.WorkingHours.StartMinute/60, 0, 0, 0, time.Local)
	if !c.IsWorkingNow(inside) {
		t.Errorf("expected working at start of repaired window")
	}
}

func TestStoreOpenReadErrorFallsBackToDefaults(t *testing.T) {
	// Passing a directory makes os.ReadFile fail with a non-NotExist error.
	s, err := Open(t.TempDir())
	if s == nil {
		t.Fatal("Open returned a nil store on read error; callers would panic")
	}
	if err == nil {
		t.Fatal("expected a read error to be reported")
	}
	if got := s.Get().Schedule.ShortInterval; got != Default().Schedule.ShortInterval {
		t.Errorf("store did not fall back to defaults: %v", got)
	}
}

func TestStoreConcurrentSetAndClose(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "cfg.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				c := s.Get()
				c.Display.Theme = "dark"
				_, _ = s.Set(c)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Close()
	}()
	wg.Wait()

	// Subscribing after Close must yield an already-closed channel.
	ch := s.Subscribe()
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected closed subscriber channel after Close")
		}
	case <-time.After(time.Second):
		t.Error("subscriber channel was not closed after Close")
	}
}

func TestDefaultEnablesMediaCountsAsActivity(t *testing.T) {
	if !Default().Idle.MediaCountsAsActivity {
		t.Error("Default().Idle.MediaCountsAsActivity should be true")
	}
	if c := Default(); c.Validate() {
		t.Fatal("Default() should already be valid")
	}
}

// Config files written before mediaCountsAsActivity existed must migrate to
// true (the fix); an explicitly saved false must be preserved.
func TestStoreOpenMigratesMissingMediaCountsAsActivity(t *testing.T) {
	legacy := Default()
	legacy.Idle.MediaCountsAsActivity = false // value is irrelevant; key will be stripped
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	idle, ok := doc["idle"].(map[string]any)
	if !ok {
		t.Fatal("expected idle object in marshalled config")
	}
	delete(idle, "mediaCountsAsActivity")
	stripped, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "cfg.json")
	if err := os.WriteFile(path, stripped, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Get().Idle.MediaCountsAsActivity {
		t.Error("legacy config without the key should migrate to true")
	}

	// Explicit false survives a round-trip.
	c := s.Get()
	c.Idle.MediaCountsAsActivity = false
	if _, err := s.Set(c); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s2.Get().Idle.MediaCountsAsActivity {
		t.Error("explicitly saved false must be preserved, not re-migrated")
	}
}
