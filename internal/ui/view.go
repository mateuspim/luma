package ui

import (
	"fmt"
	"strings"

	"github.com/pym/luma/internal/ddc"
)

const (
	defaultBoxWidth       = 74
	defaultListBarWidth   = 44 // = defaultBoxWidth - 30
	defaultSliderBarWidth = 60 // = defaultBoxWidth - 14
	nameColWidth          = 14
)

// boxWidth returns the terminal-adaptive box width.
func (m Model) boxWidth() int {
	if m.width >= defaultBoxWidth {
		return m.width
	}
	return defaultBoxWidth
}

// listBarWidth returns the gradient bar width for the list view.
func (m Model) listBarWidth() int {
	w := m.boxWidth() - 30
	if w < defaultListBarWidth {
		return defaultListBarWidth
	}
	return w
}

// sliderBarWidth returns the gradient bar width for the slider view.
func (m Model) sliderBarWidth() int {
	w := m.boxWidth() - 14
	if w < defaultSliderBarWidth {
		return defaultSliderBarWidth
	}
	return w
}

// innerPad returns top and bottom empty-row counts to insert inside the box
// so the total rendered lines equals m.height-1 (one line reserved to prevent
// the top border from scrolling off screen). fixedLines is the count without padding.
func (m Model) innerPad(fixedLines int) (top, bottom int) {
	effective := m.height - 1
	if effective <= fixedLines {
		return 0, 0
	}
	extra := effective - fixedLines
	top = extra / 2
	bottom = extra - top
	return
}

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
	bw := m.boxWidth()
	// fixed: top + header + sep + N displays + sep + footer + bottom = N + 6
	padTop, padBot := m.innerPad(len(m.displays) + 6)
	var sb strings.Builder

	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")

	// Header: ◈ luma with version and spinner when ddcutil is running
	header := m.styles.Accent.Render("◈ luma " + m.version)
	if m.executor.IsBusy() {
		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		header += " " + spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	}
	sb.WriteString(m.boxLine(header))

	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")

	for range padTop {
		sb.WriteString(m.boxLine(""))
	}
	for i, d := range m.displays {
		sb.WriteString(m.viewListRow(i, d))
	}
	for range padBot {
		sb.WriteString(m.boxLine(""))
	}

	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")
	separator := m.styles.Accent.Render(" ◆ ")
	footer := fmt.Sprintf("  ↑/↓ select · Enter slider · [a]ll · [q]uit%s+/- ±%d · [/] ±%d · {/} ±%d",
		separator, m.cfg.Steps.Small, m.cfg.Steps.Medium, m.cfg.Steps.Large)
	sb.WriteString(m.boxLine(footer))
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")

	return sb.String()
}

// viewListRow renders one display row in the list.
func (m Model) viewListRow(i int, d ddc.Display) string {
	cursor := "  "
	name := truncate(d.Name, nameColWidth)
	namePadded := fmt.Sprintf("%-*s", nameColWidth, name)

	selected := i == m.selected
	if selected {
		cursor = m.styles.Accent.Render("▸ ")
		namePadded = m.styles.AccentBold.Render(namePadded)
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
	bar := renderGradientBar(d.Brightness, maxVal, m.listBarWidth(), m.cfg.Theme.AccentColor)
	pctStr := fmt.Sprintf("%3d%%", pct)
	if selected {
		pctStr = m.styles.AccentBold.Render(pctStr)
	} else {
		pctStr = m.styles.Accent.Render(pctStr)
	}

	return m.boxLine(cursor + namePadded + "   " + bar + "   " + pctStr)
}

// viewSliderScreen renders Screen 2 (per-display) or Screen 3 (all displays).
func (m Model) viewSliderScreen(all bool) string {
	bw := m.boxWidth()
	// fixed: top + header + sep + empty + bar + empty + sep + footer + bottom = 9
	padTop, padBot := m.innerPad(9)
	var sb strings.Builder

	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")

	// Header
	var title string
	if all {
		title = "◈ luma " + m.version + " · All Displays"
	} else if m.selected >= 0 && m.selected < len(m.displays) {
		title = "◈ luma " + m.version + " · " + m.displays[m.selected].Name
	} else {
		title = "◈ luma " + m.version
	}
	lhs := m.styles.Accent.Render(title)
	rhs := "[q]uit"
	sb.WriteString(m.boxRow(lhs, rhs))

	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")

	for range padTop {
		sb.WriteString(m.boxLine(""))
	}

	// Empty padding row + bar + empty padding row (always present for visual breathing room)
	sb.WriteString(m.boxLine(""))
	bar := renderGradientBar(m.sliderVal, 100, m.sliderBarWidth(), m.cfg.Theme.AccentColor)
	pctStr := m.styles.Accent.Render(fmt.Sprintf("%3d%%", m.sliderVal))
	sb.WriteString(m.boxLine("   " + bar + "   " + pctStr))
	sb.WriteString(m.boxLine(""))

	for range padBot {
		sb.WriteString(m.boxLine(""))
	}

	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")
	footer := fmt.Sprintf("  ←/→ ±%d · [/] ±%d · {/} ±%d · Enter apply · Esc back",
		m.cfg.Steps.Small, m.cfg.Steps.Medium, m.cfg.Steps.Large)
	sb.WriteString(m.boxLine(footer))
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")

	return sb.String()
}

func (m Model) viewError() string {
	bw := m.boxWidth()
	// fixed: top + header + sep + error + [install] + quit + bottom
	fixedLines := 6
	if isNotFound(m.err) {
		fixedLines = 7
	}
	padTop, padBot := m.innerPad(fixedLines)
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma " + m.version)))
	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")
	for range padTop {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString(m.boxLine(m.styles.Error.Render("  Error: " + m.err.Error())))
	if isNotFound(m.err) {
		sb.WriteString(m.boxLine("  Install ddcutil: https://www.ddcutil.com"))
	}
	sb.WriteString(m.boxLine("  Press q to quit"))
	for range padBot {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")
	return sb.String()
}

func (m Model) viewLoading() string {
	bw := m.boxWidth()
	// fixed: top + header + bottom = 3
	padTop, padBot := m.innerPad(3)
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")
	for range padTop {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma "+m.version) + m.styles.Busy.Render(" ⟳ detecting displays...")))
	for range padBot {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")
	return sb.String()
}

func (m Model) viewNoDisplays() string {
	bw := m.boxWidth()
	// fixed: top + header + sep + msg1 + msg2 + bottom = 6
	padTop, padBot := m.innerPad(6)
	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")
	sb.WriteString(m.boxLine(m.styles.Accent.Render("◈ luma " + m.version)))
	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")
	for range padTop {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString(m.boxLine("  No DDC/CI displays found. Check connections."))
	sb.WriteString(m.boxLine("  Press q to quit."))
	for range padBot {
		sb.WriteString(m.boxLine(""))
	}
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")
	return sb.String()
}

// boxLine wraps content in │ borders, padding to boxWidth.
func (m Model) boxLine(content string) string {
	visible := visibleLen(content)
	pad := m.boxWidth() - 2 - visible
	if pad < 0 {
		pad = 0
	}
	return "│" + content + strings.Repeat(" ", pad) + "│\n"
}

// boxRow renders a header row with left and right content.
func (m Model) boxRow(left, right string) string {
	lv := visibleLen(left)
	rv := visibleLen(right)
	inner := m.boxWidth() - 2
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
