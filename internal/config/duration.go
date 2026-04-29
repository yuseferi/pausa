package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a time.Duration with human-readable JSON encoding ("10m",
// "20s") so the config file is editable by hand.
type Duration time.Duration

// MarshalJSON encodes d as a Go duration string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON accepts either a duration string ("10m") or a number of
// nanoseconds (for forward compatibility with hand-edited files).
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		*d = Duration(time.Duration(x))
		return nil
	case string:
		td, err := time.ParseDuration(x)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", x, err)
		}
		*d = Duration(td)
		return nil
	default:
		return fmt.Errorf("invalid duration value %v", v)
	}
}

// AsDuration returns d as a time.Duration.
func (d Duration) AsDuration() time.Duration { return time.Duration(d) }
