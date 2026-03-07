package ui

import "testing"

func TestInterpolateColor(t *testing.T) {
	cases := []struct {
		from, to string
		tval     float64
		want     string
	}{
		{"#000000", "#FFFFFF", 0.0, "#000000"},
		{"#000000", "#FFFFFF", 1.0, "#FFFFFF"},
		{"#000000", "#FFFFFF", 0.5, "#7F7F7F"},
		{"#000000", "#9B59B6", 0.0, "#000000"},
		{"#000000", "#9B59B6", 1.0, "#9B59B6"},
	}
	for _, c := range cases {
		got := InterpolateColor(c.from, c.to, c.tval)
		if got != c.want {
			t.Errorf("InterpolateColor(%q, %q, %.1f) = %q, want %q",
				c.from, c.to, c.tval, got, c.want)
		}
	}
}

func TestParseHex(t *testing.T) {
	r, g, b := parseHex("#9B59B6")
	if r != 0x9B || g != 0x59 || b != 0xB6 {
		t.Errorf("parseHex(#9B59B6) = (%d,%d,%d), want (155,89,182)", r, g, b)
	}

	// Bad input returns zero
	r, g, b = parseHex("bad")
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("parseHex(bad) should return zeros")
	}
}
