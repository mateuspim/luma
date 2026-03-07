package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles.
type Styles struct {
	Accent    lipgloss.Style
	Dim       lipgloss.Style
	Bold      lipgloss.Style
	Selected  lipgloss.Style
	Error     lipgloss.Style
	Busy      lipgloss.Style
	Bar       lipgloss.Style
	BarEmpty  lipgloss.Style
}

func newStyles(accentColor string) Styles {
	accent := lipgloss.Color(accentColor)
	return Styles{
		Accent:   lipgloss.NewStyle().Foreground(accent),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Bold:     lipgloss.NewStyle().Bold(true),
		Selected: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Busy:     lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		Bar:      lipgloss.NewStyle().Foreground(accent),
		BarEmpty: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
	}
}
