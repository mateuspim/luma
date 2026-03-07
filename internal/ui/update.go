package ui

import (
	"context"
	"errors"

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
			if isTimeout(msg.err) {
				m.displays[i].TimedOut = true
			} else if !isBusy(msg.err) {
				m.displays[i].Disconnected = true
			}
			return m, nil
		}
		m.displays[i].TimedOut = false
		m.displays[i].Disconnected = false
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
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}
	switch m.mode {
	case ModeNormal:
		return m.handleNormalKey(msg)
	case ModeSlider:
		return m.handleSliderKey(msg, false)
	case ModeAllSlider:
		return m.handleSliderKey(msg, true)
	}
	return m, nil
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.displays) == 0 {
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

	case key.Matches(msg, m.keys.Confirm):
		// Enter → per-display slider
		if m.selected >= 0 && m.selected < len(m.displays) {
			m.sliderVal = m.displays[m.selected].Brightness
			m.mode = ModeSlider
		}

	case key.Matches(msg, m.keys.AllSlider):
		// a → set-all slider
		m.sliderVal = m.avgBrightness()
		m.mode = ModeAllSlider

	case key.Matches(msg, m.keys.Inc):
		return m.applyStep(m.cfg.Steps.Small)

	case key.Matches(msg, m.keys.Dec):
		return m.applyStep(-m.cfg.Steps.Small)

	case key.Matches(msg, m.keys.IncMed):
		return m.applyStep(m.cfg.Steps.Medium)

	case key.Matches(msg, m.keys.DecMed):
		return m.applyStep(-m.cfg.Steps.Medium)
	}

	return m, nil
}

// handleSliderKey handles keys for both ModeSlider and ModeAllSlider.
func (m Model) handleSliderKey(msg tea.KeyMsg, all bool) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal

	case key.Matches(msg, m.keys.Confirm):
		val := clampInt(m.sliderVal, 0, 100)
		m.mode = ModeNormal
		if all {
			return m, setAllBrightness(m.client, m.displays, val)
		}
		if m.selected >= 0 && m.selected < len(m.displays) {
			d := m.displays[m.selected]
			m.displays[m.selected].Brightness = val
			return m, setBrightness(m.client, d, val)
		}

	case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.Dec):
		m.sliderVal = clampInt(m.sliderVal-m.cfg.Steps.Small, 0, 100)

	case key.Matches(msg, m.keys.Right), key.Matches(msg, m.keys.Inc):
		m.sliderVal = clampInt(m.sliderVal+m.cfg.Steps.Small, 0, 100)

	case key.Matches(msg, m.keys.DecMed):
		m.sliderVal = clampInt(m.sliderVal-m.cfg.Steps.Medium, 0, 100)

	case key.Matches(msg, m.keys.IncMed):
		m.sliderVal = clampInt(m.sliderVal+m.cfg.Steps.Medium, 0, 100)
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

	m.debounceSeq[d.Index]++
	seq := m.debounceSeq[d.Index]
	m.anyDebouncing = true

	return m, debounceCmd(d.Index, d.Brightness, seq, m.cfg.Guardrails.DebounceMs)
}

// handleDebounceFire fires ddcutil only if the sequence still matches.
func (m Model) handleDebounceFire(msg debounceFireMsg) (tea.Model, tea.Cmd) {
	if m.debounceSeq[msg.displayIdx] != msg.seq {
		return m, nil
	}
	delete(m.debounceSeq, msg.displayIdx)
	m.anyDebouncing = len(m.debounceSeq) > 0

	i := m.findDisplayByIndex(msg.displayIdx)
	if i < 0 {
		return m, nil
	}
	return m, setBrightness(m.client, m.displays[i], msg.value)
}

// handleTick processes auto-refresh ticks.
func (m Model) handleTick() (tea.Model, tea.Cmd) {
	var nextTick tea.Cmd
	if m.cfg.Display.RefreshIntervalMs > 0 {
		nextTick = autoRefreshCmd(m.cfg.Display.RefreshIntervalMs)
	}
	if m.mode != ModeNormal || m.executor.IsBusy() || m.anyDebouncing || m.loading {
		return m, nextTick
	}
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
			if updated[i].Disconnected {
				continue
			}
			cur, max, err := client.GetBrightness(ctx, updated[i])
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					updated[i].TimedOut = true
					changed = true
				}
				continue
			}
			updated[i].TimedOut = false
			if cur != updated[i].Brightness || max != updated[i].MaxVal {
				updated[i].Brightness = cur
				updated[i].MaxVal = max
				changed = true
			}
		}
		if !changed {
			return nil
		}
		return refreshDoneMsg{displays: updated}
	}
}

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

func isTimeout(err error) bool {
	return err != nil && errors.Is(err, context.DeadlineExceeded)
}

func isBusy(err error) bool {
	return err != nil && errors.Is(err, ddc.ErrBusy)
}
