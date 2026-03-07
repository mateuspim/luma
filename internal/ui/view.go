package ui

import (
	"fmt"
	"strings"

	"github.com/pym/luma/internal/ddc"
)

const (
	barWidth   = 20
	barFull    = '█'
	barEmpty   = '░'
	boxWidth   = 60
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

	var sb strings.Builder

	// Top border
	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")

	// Header
	busy := ""
	if m.executor.IsBusy() {
		busy = m.styles.Busy.Render(" ⟳")
	}
	header := m.styles.Accent.Render("◈ luma") + busy
	rhs := m.styles.Dim.Render("[r]efresh  [s]lider  [q]uit")
	sb.WriteString(m.boxRow(header, rhs))

	// Divider
	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")

	// Display rows
	if m.mode == ModeSlider {
		sb.WriteString(m.viewSlider())
	} else {
		for i, d := range m.displays {
			sb.WriteString(m.viewDisplayRow(i, d))
		}
	}

	// Divider + footer
	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	sb.WriteString(m.viewFooter())

	// Bottom border
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")

	return sb.String()
}

func (m Model) viewDisplayRow(i int, d ddc.Display) string {
	cursor := "  "
	nameStyle := m.styles.Dim
	if i == m.selected {
		cursor = m.styles.Selected.Render("▸ ")
		nameStyle = m.styles.Selected
	}

	name := d.Model
	if name == "" {
		name = fmt.Sprintf("Display %d", d.Index)
	}
	if len(name) > 16 {
		name = name[:16]
	}

	var valueStr string
	var bar string

	if m.mode == ModeInput && i == m.selected {
		valueStr = fmt.Sprintf("[ %s_ ]", m.inputBuf)
		bar = ""
	} else {
		brightness := d.Brightness
		maxVal := d.MaxVal
		if maxVal <= 0 {
			maxVal = 100
		}
		pct := brightness * 100 / maxVal
		bar = renderBar(brightness, maxVal, barWidth, m.styles)
		valueStr = fmt.Sprintf("%3d%%", pct)
	}

	status := ""
	if d.Disconnected {
		status = m.styles.Error.Render(" [disconnected]")
	} else if d.TimedOut {
		status = m.styles.Error.Render(" [?]")
	}

	line := cursor + nameStyle.Render(fmt.Sprintf("%-16s", name)) + "  " + bar + " " + valueStr + status
	return m.boxLine(line)
}

func (m Model) viewSlider() string {
	maxVal := 100
	bar := renderBar(m.sliderVal, maxVal, barWidth, m.styles)
	label := m.styles.Accent.Render("◈ Set All")
	line := label + "  " + bar + fmt.Sprintf("  %3d%%", m.sliderVal)
	result := m.boxLine(line)
	hint := m.styles.Dim.Render("  ←/→ adjust · Enter to apply · Esc to cancel")
	result += m.boxLine(hint)
	return result
}

func (m Model) viewFooter() string {
	var hint string
	switch m.mode {
	case ModeSlider:
		hint = "←/→ [/]  adjust · Enter apply · Esc cancel"
	case ModeInput:
		hint = "type value · Enter apply · Esc cancel"
	default:
		hint = "↑/↓ select · +/- [/] {/} adjust · s slider"
	}
	return m.boxLine(m.styles.Dim.Render("  " + hint))
}

func (m Model) viewError() string {
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", boxWidth-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma")))
	sb.WriteString("├" + strings.Repeat("─", boxWidth-2) + "┤\n")
	sb.WriteString(m.boxLine(m.styles.Error.Render("  Error: " + m.err.Error())))
	if isNotFound(m.err) {
		sb.WriteString(m.boxLine(m.styles.Dim.Render("  Install ddcutil: https://www.ddcutil.com/installation/")))
	}
	sb.WriteString(m.boxLine(m.styles.Dim.Render("  Press q to quit")))
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
	sb.WriteString(m.boxLine(m.styles.Dim.Render("  No DDC/CI displays found. Check connections.")))
	sb.WriteString(m.boxLine(m.styles.Dim.Render("  Press r to retry, q to quit.")))
	sb.WriteString("╰" + strings.Repeat("─", boxWidth-2) + "╯\n")
	return sb.String()
}

// boxLine wraps content in │ borders, padding to boxWidth.
func (m Model) boxLine(content string) string {
	// visible length for padding
	visible := visibleLen(content)
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	return "│" + content + strings.Repeat(" ", pad) + "│\n"
}

// boxRow renders a row with left and right content.
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

// renderBar produces a filled/empty block bar.
func renderBar(current, max, width int, s Styles) string {
	if max <= 0 {
		max = 100
	}
	filled := current * width / max
	if filled > width {
		filled = width
	}
	full := s.Bar.Render(strings.Repeat(string(barFull), filled))
	empty := s.BarEmpty.Render(strings.Repeat(string(barEmpty), width-filled))
	return full + empty
}

// visibleLen approximates the visible character width of a string
// by stripping ANSI escape sequences.
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
