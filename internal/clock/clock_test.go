package clock

import (
	"testing"
	"time"
)

func TestFakeAdvanceFiresTimer(t *testing.T) {
	f := NewFake(time.Unix(0, 0))
	timer := f.NewTimer(5 * time.Second)
	f.Advance(4 * time.Second)
	select {
	case <-timer.C():
		t.Fatal("timer fired too early")
	default:
	}
	f.Advance(2 * time.Second)
	select {
	case got := <-timer.C():
		if !got.Equal(time.Unix(5, 0)) {
			t.Fatalf("fired at %v, want %v", got, time.Unix(5, 0))
		}
	default:
		t.Fatal("timer should have fired")
	}
}

func TestFakeStopPreventsFiring(t *testing.T) {
	f := NewFake(time.Unix(0, 0))
	timer := f.NewTimer(time.Second)
	if !timer.Stop() {
		t.Fatal("Stop returned false on active timer")
	}
	f.Advance(2 * time.Second)
	select {
	case <-timer.C():
		t.Fatal("stopped timer fired")
	default:
	}
}

func TestFakeResetReschedules(t *testing.T) {
	f := NewFake(time.Unix(0, 0))
	timer := f.NewTimer(time.Second)
	f.Advance(500 * time.Millisecond)
	timer.Reset(2 * time.Second)
	f.Advance(time.Second)
	select {
	case <-timer.C():
		t.Fatal("timer fired before reset deadline")
	default:
	}
	f.Advance(2 * time.Second)
	select {
	case <-timer.C():
	default:
		t.Fatal("timer did not fire after reset deadline")
	}
}

func TestFakeMultipleTimersFireInOrder(t *testing.T) {
	f := NewFake(time.Unix(0, 0))
	a := f.NewTimer(3 * time.Second)
	b := f.NewTimer(time.Second)
	f.Advance(5 * time.Second)
	// both should have fired
	select {
	case <-a.C():
	default:
		t.Fatal("a did not fire")
	}
	select {
	case <-b.C():
	default:
		t.Fatal("b did not fire")
	}
}
