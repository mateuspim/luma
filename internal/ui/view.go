package ui

import (
	"fmt"
	"strings"

	"github.com/pym/luma/internal/ddc"
)

const (
	defaultBoxWidth = 74 // used only before first WindowSizeMsg
	nameColWide     = 14
	nameColCompact  = 10
)

// boxWidth returns the actual terminal width, falling back to defaultBoxWidth
// before the first WindowSizeMsg arrives.
func (m Model) boxWidth() int {
	if m.width > 0 {
		return m.width
	}
	return defaultBoxWidth
}

// nameColWidth returns the display name column width adapted to the terminal.
func (m Model) nameColWidth() int {
	bw := m.boxWidth()
	switch {
	case bw >= 74:
		return nameColWide
	case bw >= 40:
		return nameColCompact
	default:
		// No bar: name fills the inner space minus cursor(2)+gap(2)+pct(4)+borders(2)
		w := bw - 10
		if w < 4 {
			return 4
		}
		return w
	}
}

// showBar returns true if there is enough room to render a gradient bar in list rows.
func (m Model) showBar() bool {
	return m.boxWidth() >= 40
}

// listGap returns the spacing string between row elements.
func (m Model) listGap() string {
	if m.boxWidth() >= 74 {
		return "   "
	}
	return "  "
}

// listBarWidth returns the inner block count for the list gradient bar.
// Only meaningful when showBar() is true.
// Row layout: borders(2) + cursor(2) + name(ncw) + gap + caps(2) + gap + pct(4) = ncw + 2*gap + 10
func (m Model) listBarWidth() int {
	bw := m.boxWidth()
	ncw := m.nameColWidth()
	gap := len(m.listGap())
	w := bw - ncw - 2*gap - 10
	if w < 1 {
		return 1
	}
	return w
}

// sliderGap returns the spacing string around the slider bar.
func (m Model) sliderGap() string {
	if m.boxWidth() >= 56 {
		return "   "
	}
	return "  "
}

// sliderBarWidth returns the inner block count for the slider gradient bar.
// Row layout: borders(2) + gap + caps(2) + gap + pct(4) = 2*gap + 8
func (m Model) sliderBarWidth() int {
	bw := m.boxWidth()
	gap := len(m.sliderGap())
	w := bw - 2*gap - 8
	if w < 1 {
		return 1
	}
	return w
}

// innerPad returns top and bottom empty-row counts to insert inside the box so
// the total rendered lines equals m.height-1 (one line reserved to prevent the
// top border from scrolling off screen).
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

// listFooter returns help text for the list view, adapted to terminal width.
func (m Model) listFooter() string {
	s := m.cfg.Steps
	sep := m.styles.Accent.Render(" ◆ ")
	bw := m.boxWidth()
	switch {
	case bw >= 74:
		return fmt.Sprintf("  ↑/↓ Enter · [a]ll · [q]uit%s+/- ±%d · [/] ±%d · {/} ±%d",
			sep, s.Small, s.Medium, s.Large)
	case bw >= 50:
		return fmt.Sprintf("  ↑/↓ [a] [q]%s±%d · ±%d · ±%d",
			sep, s.Small, s.Medium, s.Large)
	default:
		return fmt.Sprintf("  ↑↓ [a][q] ±%d ±%d ±%d", s.Small, s.Medium, s.Large)
	}
}

// sliderFooter returns help text for the slider view, adapted to terminal width.
func (m Model) sliderFooter() string {
	s := m.cfg.Steps
	bw := m.boxWidth()
	switch {
	case bw >= 74:
		return fmt.Sprintf("  ←/→ ±%d · [/] ±%d · {/} ±%d · Enter apply · Esc back",
			s.Small, s.Medium, s.Large)
	case bw >= 50:
		return fmt.Sprintf("  ←/→ ±%d · [/] ±%d · {/} ±%d  Enter Esc",
			s.Small, s.Medium, s.Large)
	default:
		return fmt.Sprintf("  ←/→ ±%d/±%d/±%d  Enter Esc", s.Small, s.Medium, s.Large)
	}
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
	sb.WriteString(m.boxLine(m.listFooter()))
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")

	return sb.String()
}

// viewListRow renders one display row in the list.
func (m Model) viewListRow(i int, d ddc.Display) string {
	cursor := "  "
	ncw := m.nameColWidth()
	name := truncate(d.Name, ncw)
	namePadded := fmt.Sprintf("%-*s", ncw, name)
	gap := m.listGap()

	selected := i == m.selected
	if selected {
		cursor = m.styles.Accent.Render("▸ ")
		namePadded = m.styles.AccentBold.Render(namePadded)
	}

	if d.Disconnected {
		return m.boxLine(cursor + namePadded + m.styles.Error.Render(" [off]"))
	}
	if d.TimedOut {
		return m.boxLine(cursor + namePadded + m.styles.Error.Render(" [?]"))
	}

	maxVal := d.MaxVal
	if maxVal <= 0 {
		maxVal = 100
	}
	pct := d.Brightness * 100 / maxVal
	pctStr := fmt.Sprintf("%3d%%", pct)
	if selected {
		pctStr = m.styles.AccentBold.Render(pctStr)
	} else {
		pctStr = m.styles.Accent.Render(pctStr)
	}

	if !m.showBar() {
		return m.boxLine(cursor + namePadded + gap + pctStr)
	}

	bar := renderGradientBar(d.Brightness, maxVal, m.listBarWidth(), m.cfg.Theme.AccentColor)
	return m.boxLine(cursor + namePadded + gap + bar + gap + pctStr)
}

// viewSliderScreen renders Screen 2 (per-display) or Screen 3 (all displays).
func (m Model) viewSliderScreen(all bool) string {
	bw := m.boxWidth()
	// fixed: top + header + sep + empty + bar + empty + sep + footer + bottom = 9
	padTop, padBot := m.innerPad(9)
	gap := m.sliderGap()
	var sb strings.Builder

	sb.WriteString("╭" + strings.Repeat("─", bw-2) + "╮\n")

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

	sb.WriteString(m.boxLine(""))
	bar := renderGradientBar(m.sliderVal, 100, m.sliderBarWidth(), m.cfg.Theme.AccentColor)
	pctStr := m.styles.Accent.Render(fmt.Sprintf("%3d%%", m.sliderVal))
	sb.WriteString(m.boxLine(gap + bar + gap + pctStr))
	sb.WriteString(m.boxLine(""))

	for range padBot {
		sb.WriteString(m.boxLine(""))
	}

	sb.WriteString("├" + strings.Repeat("─", bw-2) + "┤\n")
	sb.WriteString(m.boxLine(m.sliderFooter()))
	sb.WriteString("╰" + strings.Repeat("─", bw-2) + "╯\n")

	return sb.String()
}

func (m Model) viewError() string {
	bw := m.boxWidth()
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
