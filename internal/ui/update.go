package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/pym/luma/internal/ddc"
)

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case displaysLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.displays = msg.displays
		if m.selected >= len(m.displays) {
			m.selected = 0
		}
		return m, nil

	case brightnessUpdatedMsg:
		i := m.findDisplayByIndex(msg.displayIdx)
		if i < 0 {
			return m, nil
		}
		if msg.err != nil {
			// Mark display as timed out or just leave value as-is.
			return m, nil
		}
		m.displays[i].Brightness = msg.current
		m.displays[i].MaxVal = msg.max
		return m, nil

	case refreshDoneMsg:
		m.displays = msg.displays
		return m, nil

	case debounceFireMsg:
		return m.handleDebounceFire(msg)

	case tickMsg:
		return m.handleTick()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Quit always works.
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	switch m.mode {
	case ModeNormal:
		return m.handleNormalKey(msg)
	case ModeSlider:
		return m.handleSliderKey(msg)
	case ModeInput:
		return m.handleInputKey(msg)
	}
	return m, nil
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.displays) == 0 {
		if key.Matches(msg, m.keys.Refresh) {
			m.loading = true
			return m, detectDisplays(m.client)
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.selected > 0 {
			m.selected--
		}

	case key.Matches(msg, m.keys.Down):
		if m.selected < len(m.displays)-1 {
			m.selected++
		}

	case key.Matches(msg, m.keys.Refresh):
		return m, detectDisplays(m.client)

	case key.Matches(msg, m.keys.Slider):
		m.mode = ModeSlider
		m.sliderVal = m.avgBrightness()

	case key.Matches(msg, m.keys.Inc):
		return m.applyStep(m.cfg.Steps.Small)

	case key.Matches(msg, m.keys.Dec):
		return m.applyStep(-m.cfg.Steps.Small)

	case key.Matches(msg, m.keys.IncMed):
		return m.applyStep(m.cfg.Steps.Medium)

	case key.Matches(msg, m.keys.DecMed):
		return m.applyStep(-m.cfg.Steps.Medium)

	case key.Matches(msg, m.keys.IncLrg):
		return m.applyStep(m.cfg.Steps.Large)

	case key.Matches(msg, m.keys.DecLrg):
		return m.applyStep(-m.cfg.Steps.Large)

	default:
		k := msg.String()
		if len(k) == 1 && k[0] >= '0' && k[0] <= '9' && m.selected >= 0 && m.selected < len(m.displays) {
			m.inputOrigVal = m.displays[m.selected].Brightness
			m.mode = ModeInput
			m.inputBuf = k
		}
	}

	return m, nil
}

// applyStep updates local brightness optimistically and schedules a debounce timer.
func (m Model) applyStep(delta int) (tea.Model, tea.Cmd) {
	if m.selected < 0 || m.selected >= len(m.displays) {
		return m, nil
	}
	d := &m.displays[m.selected]
	maxVal := d.MaxVal
	if maxVal <= 0 {
		maxVal = 100
	}
	d.Brightness = clampInt(d.Brightness+delta, 0, maxVal)

	// Increment debounce sequence for this display.
	m.debounceSeq[d.Index]++
	seq := m.debounceSeq[d.Index]
	m.anyDebouncing = true

	return m, debounceCmd(d.Index, d.Brightness, seq, m.cfg.Guardrails.DebounceMs)
}

// handleTick processes auto-refresh ticks.
// Skips the refresh if the executor is busy, a mode is active, or a debounce is pending.
func (m Model) handleTick() (tea.Model, tea.Cmd) {
	// Schedule the next tick regardless.
	var nextTick tea.Cmd
	if m.cfg.Display.RefreshIntervalMs > 0 {
		nextTick = autoRefreshCmd(m.cfg.Display.RefreshIntervalMs)
	}

	// Guard: skip if any condition makes it unsafe.
	if m.mode != ModeNormal || m.executor.IsBusy() || m.anyDebouncing || m.loading {
		return m, nextTick
	}

	// Safe to refresh: fetch all display brightnesses.
	return m, tea.Batch(nextTick, refreshAllBrightness(m.client, m.displays))
}

// refreshAllBrightness fetches brightness for all displays sequentially.
func refreshAllBrightness(client *ddc.Client, displays []ddc.Display) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		updated := make([]ddc.Display, len(displays))
		copy(updated, displays)
		changed := false
		for i := range updated {
			cur, max, err := client.GetBrightness(ctx, updated[i].Index)
			if err != nil {
				continue
			}
			if cur != updated[i].Brightness || max != updated[i].MaxVal {
				updated[i].Brightness = cur
				updated[i].MaxVal = max
				changed = true
			}
		}
		if !changed {
			return nil // no-op, avoid flicker
		}
		return refreshDoneMsg{displays: updated}
	}
}

// handleDebounceFire fires ddcutil only if the sequence still matches.
func (m Model) handleDebounceFire(msg debounceFireMsg) (tea.Model, tea.Cmd) {
	if m.debounceSeq[msg.displayIdx] != msg.seq {
		// Stale timer — a newer keypress already superseded this.
		return m, nil
	}
	// This is the latest timer for this display; clear debouncing state.
	delete(m.debounceSeq, msg.displayIdx)
	m.anyDebouncing = len(m.debounceSeq) > 0

	i := m.findDisplayByIndex(msg.displayIdx)
	if i < 0 {
		return m, nil
	}
	d := m.displays[i]
	return m, setBrightness(m.client, d, msg.value)
}

func (m Model) handleSliderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal

	case key.Matches(msg, m.keys.Confirm):
		val := clampInt(m.sliderVal, 0, 100)
		m.mode = ModeNormal
		return m, setAllBrightness(m.client, m.displays, val)

	case key.Matches(msg, m.keys.Inc), key.Matches(msg, m.keys.IncMed):
		step := m.cfg.Steps.Small
		if key.Matches(msg, m.keys.IncMed) {
			step = m.cfg.Steps.Medium
		}
		m.sliderVal = clampInt(m.sliderVal+step, 0, 100)

	case key.Matches(msg, m.keys.Dec), key.Matches(msg, m.keys.DecMed):
		step := m.cfg.Steps.Small
		if key.Matches(msg, m.keys.DecMed) {
			step = m.cfg.Steps.Medium
		}
		m.sliderVal = clampInt(m.sliderVal-step, 0, 100)

	case key.Matches(msg, m.keys.IncLrg):
		m.sliderVal = clampInt(m.sliderVal+m.cfg.Steps.Large, 0, 100)

	case key.Matches(msg, m.keys.DecLrg):
		m.sliderVal = clampInt(m.sliderVal-m.cfg.Steps.Large, 0, 100)
	}

	return m, nil
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		// Restore original brightness on cancel.
		if m.selected >= 0 && m.selected < len(m.displays) {
			m.displays[m.selected].Brightness = m.inputOrigVal
		}
		m.mode = ModeNormal
		m.inputBuf = ""

	case key.Matches(msg, m.keys.Confirm):
		val := parseInputVal(m.inputBuf)
		m.inputBuf = ""
		m.mode = ModeNormal
		if val < 0 {
			return m, nil
		}
		if m.selected >= 0 && m.selected < len(m.displays) {
			d := m.displays[m.selected]
			maxVal := d.MaxVal
			if maxVal <= 0 {
				maxVal = 100
			}
			val = clampInt(val, 0, maxVal)
			m.displays[m.selected].Brightness = val
			return m, setBrightness(m.client, d, val)
		}

	default:
		k := msg.String()
		if len(k) == 1 && k[0] >= '0' && k[0] <= '9' && len(m.inputBuf) < 3 {
			m.inputBuf += k
		} else if k == "backspace" && len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
	}

	return m, nil
}

// avgBrightness returns the average brightness of all displays.
func (m Model) avgBrightness() int {
	if len(m.displays) == 0 {
		return 50
	}
	sum := 0
	for _, d := range m.displays {
		sum += d.Brightness
	}
	return sum / len(m.displays)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func parseInputVal(s string) int {
	if s == "" {
		return -1
	}
	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		val = val*10 + int(c-'0')
	}
	return val
}
