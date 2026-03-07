package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles.
type Styles struct {
	Accent   lipgloss.Style // applied to: ◈ header, ▸ cursor, filled bar
	Error    lipgloss.Style
	Busy     lipgloss.Style
	Bar      lipgloss.Style // filled bar segments (same color as Accent)
	BarEmpty lipgloss.Style // empty bar segments
}

func newStyles(accentColor string) Styles {
	accent := lipgloss.Color(accentColor)
	return Styles{
		Accent:   lipgloss.NewStyle().Foreground(accent),
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Busy:     lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		Bar:      lipgloss.NewStyle().Foreground(accent),
		BarEmpty: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
	}
}
