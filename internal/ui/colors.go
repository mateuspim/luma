package ui

import (
	"fmt"
	"strconv"
)

// InterpolateColor linearly interpolates between two #RRGGBB hex colors.
// t=0.0 returns from, t=1.0 returns to.
func InterpolateColor(from, to string, t float64) string {
	fr, fg, fb := parseHex(from)
	tr, tg, tb := parseHex(to)
	return fmt.Sprintf("#%02X%02X%02X",
		lerp(fr, tr, t),
		lerp(fg, tg, t),
		lerp(fb, tb, t),
	)
}

func parseHex(color string) (r, g, b int) {
	if len(color) != 7 || color[0] != '#' {
		return 0, 0, 0
	}
	rv, _ := strconv.ParseInt(color[1:3], 16, 32)
	gv, _ := strconv.ParseInt(color[3:5], 16, 32)
	bv, _ := strconv.ParseInt(color[5:7], 16, 32)
	return int(rv), int(gv), int(bv)
}

func lerp(a, b int, t float64) int {
	v := float64(a) + (float64(b)-float64(a))*t
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return int(v)
}
