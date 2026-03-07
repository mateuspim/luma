package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pym/luma/internal/config"
	"github.com/pym/luma/internal/ddc"
)

// Mode represents the current input mode.
type Mode int

const (
	ModeNormal    Mode = iota
	ModeSlider         // per-display slider (Enter from list)
	ModeAllSlider      // set-all slider (a from list)
)

// Model is the root Bubble Tea model.
type Model struct {
	cfg      config.Config
	client   *ddc.Client
	executor *ddc.Executor

	displays []ddc.Display
	selected int
	mode     Mode

	// slider state
	sliderVal int

	// debounce: sequence counter per display index
	// When a step key fires, seq increments. The debounce timer carries
	// the seq at the time it was started; if seq has changed by the time
	// it fires, the timer is stale and is dropped.
	debounceSeq map[int]int

	// anyDebouncing is true if any display has a pending debounce timer.
	anyDebouncing bool

	// spinnerFrame is the current index into spinnerFrames.
	spinnerFrame int

	loading bool
	err     error

	styles Styles
	keys   KeyMap
	width  int
	height int
}

// -- message types --

type displaysLoadedMsg struct {
	displays []ddc.Display
	err      error
}

type brightnessUpdatedMsg struct {
	displayIdx int // ddcutil display index
	current    int
	max        int
	err        error
}

type debounceFireMsg struct {
	displayIdx int
	value      int
	seq        int // to detect stale timers
}

type tickMsg struct{}

type spinnerTickMsg struct{}

// debounceCmd returns a command that fires debounceFireMsg after debounceMs.
func debounceCmd(displayIdx, value, seq, debounceMs int) tea.Cmd {
	return tea.Tick(time.Duration(debounceMs)*time.Millisecond, func(_ time.Time) tea.Msg {
		return debounceFireMsg{displayIdx: displayIdx, value: value, seq: seq}
	})
}

// autoRefreshCmd fires tickMsg after intervalMs.
func autoRefreshCmd(intervalMs int) tea.Cmd {
	return tea.Tick(time.Duration(intervalMs)*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

// spinnerTickCmd fires spinnerTickMsg every 80ms for smooth spinner animation.
func spinnerTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(_ time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

type refreshDoneMsg struct {
	displays []ddc.Display
}

// New constructs the root model.
func New(cfg config.Config) Model {
	exec := ddc.NewExecutor(cfg.Guardrails.MaxDdcutilProcs, cfg.Guardrails.CommandTimeoutS)
	client := ddc.NewClient(exec)
	return Model{
		cfg:         cfg,
		client:      client,
		executor:    exec,
		loading:     true,
		debounceSeq: make(map[int]int),
		styles:      newStyles(cfg.Theme.AccentColor),
		keys:        defaultKeyMap(),
	}
}

// Init triggers display detection on startup and schedules the first refresh tick.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{detectDisplays(m.client), spinnerTickCmd()}
	if m.cfg.Display.RefreshIntervalMs > 0 {
		cmds = append(cmds, autoRefreshCmd(m.cfg.Display.RefreshIntervalMs))
	}
	return tea.Batch(cmds...)
}

// detectDisplays fetches all displays and their brightness values.
func detectDisplays(client *ddc.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		displays, err := client.Detect(ctx)
		if err != nil {
			return displaysLoadedMsg{err: err}
		}
		for i := range displays {
			cur, max, gerr := client.GetBrightness(ctx, displays[i])
			if gerr != nil {
				displays[i].Brightness = -1
				continue
			}
			displays[i].Brightness = cur
			displays[i].MaxVal = max
		}
		return displaysLoadedMsg{displays: displays}
	}
}

// fetchBrightness fetches brightness for a single display.
func fetchBrightness(client *ddc.Client, d ddc.Display) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		cur, max, err := client.GetBrightness(ctx, d)
		return brightnessUpdatedMsg{displayIdx: d.Index, current: cur, max: max, err: err}
	}
}

// setBrightness sets brightness and then re-fetches.
func setBrightness(client *ddc.Client, d ddc.Display, value int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.SetBrightness(ctx, d, value)
		if err != nil {
			return brightnessUpdatedMsg{displayIdx: d.Index, err: fmt.Errorf("set: %w", err)}
		}
		cur, max, err := client.GetBrightness(ctx, d)
		return brightnessUpdatedMsg{displayIdx: d.Index, current: cur, max: max, err: err}
	}
}

// setAllBrightness sets brightness on all displays.
func setAllBrightness(client *ddc.Client, displays []ddc.Display, value int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_ = client.SetBrightnessAll(ctx, displays, value)
		// Re-fetch all
		updated := make([]ddc.Display, len(displays))
		copy(updated, displays)
		for i := range updated {
			cur, max, err := client.GetBrightness(ctx, updated[i])
			if err == nil {
				updated[i].Brightness = cur
				updated[i].MaxVal = max
			}
		}
		return refreshDoneMsg{displays: updated}
	}
}

// findDisplayByIndex returns the slice index for a given ddcutil display index.
func (m *Model) findDisplayByIndex(ddcIdx int) int {
	for i, d := range m.displays {
		if d.Index == ddcIdx {
			return i
		}
	}
	return -1
}
