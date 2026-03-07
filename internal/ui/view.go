package ui

import (
	"fmt"
	"strings"

	"github.com/pym/luma/internal/ddc"
)

const (
	boxWidth       = 74
	listBarWidth   = 44 // fill blocks inside ╸...╺ caps, with 3-space padding each side
	sliderBarWidth = 60
	nameColWidth   = 14
)

func (m Model) View() string {
	if m.err != nil {
		return m.viewError()
	}
	if m.loading {
		return m.viewLoading()
	}
	if len(m.displays) == 0 {
		return m.viewNoDisplays()
	}

	switch m.mode {
	case ModeSlider:
		return m.viewSliderScreen(false)
	case ModeAllSlider:
		return m.viewSliderScreen(true)
	default:
		return m.viewList()
	}
}

// viewList renders Screen 1: the display list.
func (m Model) viewList() string {
	var sb strings.Builder

	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")

	// Header: ◈ luma with spinner when ddcutil is running
	header := m.styles.Accent.Render("◈ luma")
	if m.executor.IsBusy() {
		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		header += " " + spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	}
	sb.WriteString(m.boxLine(header))

	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")

	for i, d := range m.displays {
		sb.WriteString(m.viewListRow(i, d))
	}

	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	footer := fmt.Sprintf("  ↑/↓ select · +/- ±%d · [/] ±%d · {/} ±%d · Enter slider · [a]ll · [q]uit",
		m.cfg.Steps.Small, m.cfg.Steps.Medium, m.cfg.Steps.Large)
	sb.WriteString(m.boxLine(footer))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")

	return sb.String()
}

// viewListRow renders one display row in the list.
func (m Model) viewListRow(i int, d ddc.Display) string {
	cursor := "  "
	name := truncate(d.Name, nameColWidth)
	namePadded := fmt.Sprintf("%-*s", nameColWidth, name)

	if i == m.selected {
		cursor = m.styles.Accent.Render("▸ ")
		namePadded = m.styles.Accent.Render(namePadded)
	}

	status := ""
	if d.Disconnected {
		status = m.styles.Error.Render(" [off]")
		return m.boxLine(cursor + namePadded + status)
	}
	if d.TimedOut {
		status = m.styles.Error.Render(" [?]")
		return m.boxLine(cursor + namePadded + status)
	}

	maxVal := d.MaxVal
	if maxVal <= 0 {
		maxVal = 100
	}
	pct := d.Brightness * 100 / maxVal
	bar := renderGradientBar(d.Brightness, maxVal, listBarWidth, m.cfg.Theme.AccentColor)
	pctStr := m.styles.Accent.Render(fmt.Sprintf("%3d%%", pct))

	return m.boxLine(cursor + namePadded + "   " + bar + "   " + pctStr)
}

// viewSliderScreen renders Screen 2 (per-display) or Screen 3 (all displays).
func (m Model) viewSliderScreen(all bool) string {
	var sb strings.Builder

	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")

	// Header
	var title string
	if all {
		title = "◈ luma · All Displays"
	} else if m.selected >= 0 && m.selected < len(m.displays) {
		title = "◈ luma · " + m.displays[m.selected].Name
	} else {
		title = "◈ luma"
	}
	lhs := m.styles.Accent.Render(title)
	rhs := "[q]uit"
	sb.WriteString(m.boxRow(lhs, rhs))

	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")

	// Empty padding row
	sb.WriteString(m.boxLine(""))

	// Slider bar row
	bar := renderGradientBar(m.sliderVal, 100, sliderBarWidth, m.cfg.Theme.AccentColor)
	pctStr := m.styles.Accent.Render(fmt.Sprintf("%3d%%", m.sliderVal))
	sb.WriteString(m.boxLine("   " + bar + "   " + pctStr))

	// Empty padding row
	sb.WriteString(m.boxLine(""))

	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	footer := fmt.Sprintf("  ←/→ ±%d · [/] ±%d · {/} ±%d · Enter apply · Esc back",
		m.cfg.Steps.Small, m.cfg.Steps.Medium, m.cfg.Steps.Large)
	sb.WriteString(m.boxLine(footer))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")

	return sb.String()
}

func (m Model) viewError() string {
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma")))
	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	sb.WriteString(m.boxLine(m.styles.Error.Render("  Error: " + m.err.Error())))
	if isNotFound(m.err) {
		sb.WriteString(m.boxLine("  Install ddcutil: https://www.ddcutil.com"))
	}
	sb.WriteString(m.boxLine("  Press q to quit"))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")
	return sb.String()
}

func (m Model) viewLoading() string {
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma") + m.styles.Busy.Render(" ⟳ detecting displays...")))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")
	return sb.String()
}

func (m Model) viewNoDisplays() string {
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma")))
	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	sb.WriteString(m.boxLine("  No DDC/CI displays found. Check connections."))
	sb.WriteString(m.boxLine("  Press q to quit."))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")
	return sb.String()
}

// boxLine wraps content in │ borders, padding to boxWidth.
func (m Model) boxLine(content string) string {
	visible := visibleLen(content)
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	return "│" + content + strings.Repeat(" ", pad) + "│\n"
}

// boxRow renders a header row with left and right content.
func (m Model) boxRow(left, right string) string {
	lv := visibleLen(left)
	rv := visibleLen(right)
	inner := boxWidth - 2
	space := inner - lv - rv
	if space < 1 {
		space = 1
	}
	return "│" + left + strings.Repeat(" ", space) + right + "│\n"
}

// truncate clips s to maxLen runes.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

// visibleLen returns the visible character width by stripping ANSI escape sequences.
func visibleLen(s string) int {
	inEsc := false
	count := 0
	for _, r := range s {
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			continue
		}
		count++
	}
	return count
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "ddc: ddcutil not found, please install it"
}
