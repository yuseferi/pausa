package update

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.4", "1.0.4", 0},
		{"v1.0.5", "1.0.4", 1},
		{"1.0.3", "1.0.4", -1},
		{"1.1.0", "1.0.9", 1},
		{"2.0", "1.9.9", 1},
		{"1.0.4", "1.0.4-rc.1", 0},
		{"1.0.10", "1.0.9", 1},
		{"  v1.2.3  ", "1.2.3", 0},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}
