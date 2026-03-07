package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pym/luma/internal/config"
	"github.com/pym/luma/internal/ddc"
)

// Mode represents the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSlider
	ModeInput
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

	// input state
	inputBuf string

	// debounce: pending value per display index
	pendingVal map[int]int
	debouncing map[int]bool

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

type refreshDoneMsg struct {
	displays []ddc.Display
}

// New constructs the root model.
func New(cfg config.Config) Model {
	exec := ddc.NewExecutor(cfg.Guardrails.MaxDdcutilProcs, cfg.Guardrails.CommandTimeoutS)
	client := ddc.NewClient(exec)
	return Model{
		cfg:        cfg,
		client:     client,
		executor:   exec,
		loading:    true,
		pendingVal: make(map[int]int),
		debouncing: make(map[int]bool),
		styles:     newStyles(cfg.Theme.AccentColor),
		keys:       defaultKeyMap(),
	}
}

// Init triggers display detection on startup.
func (m Model) Init() tea.Cmd {
	return detectDisplays(m.client)
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
			cur, max, gerr := client.GetBrightness(ctx, displays[i].Index)
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
		cur, max, err := client.GetBrightness(ctx, d.Index)
		return brightnessUpdatedMsg{displayIdx: d.Index, current: cur, max: max, err: err}
	}
}

// setBrightness sets brightness and then re-fetches.
func setBrightness(client *ddc.Client, d ddc.Display, value int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.SetBrightness(ctx, d.Index, value)
		if err != nil {
			return brightnessUpdatedMsg{displayIdx: d.Index, err: fmt.Errorf("set: %w", err)}
		}
		cur, max, err := client.GetBrightness(ctx, d.Index)
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
			cur, max, err := client.GetBrightness(ctx, updated[i].Index)
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
