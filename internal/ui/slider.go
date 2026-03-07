package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderGradientBar renders a brightness bar with a gradient fill.
// Structure: ╸ + filled █ blocks (gradient from #000000→accent) + empty ─ blocks + ╺
// width is the number of fill/empty blocks (not counting the caps).
func renderGradientBar(current, max, width int, accent string) string {
	if max <= 0 {
		max = 100
	}
	filled := current * width / max
	if filled > width {
		filled = width
	}

	var sb strings.Builder
	sb.WriteString("╸")

	for i := 0; i < filled; i++ {
		t := float64(i) / float64(width)
		color := InterpolateColor("#000000", accent, t)
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("█"))
	}

	sb.WriteString(strings.Repeat("─", width-filled))
	sb.WriteString("╺")

	return sb.String()
}
