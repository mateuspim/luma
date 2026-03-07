package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles.
type Styles struct {
	Accent     lipgloss.Style // ◈ header, ▸ cursor, percentage readout
	AccentBold lipgloss.Style // bold accent for selected row
	Error      lipgloss.Style
	Busy       lipgloss.Style
}

func newStyles(accentColor string) Styles {
	accent := lipgloss.Color(accentColor)
	return Styles{
		Accent:     lipgloss.NewStyle().Foreground(accent),
		AccentBold: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Busy:       lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
	}
}
