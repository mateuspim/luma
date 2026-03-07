package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all key bindings.
type KeyMap struct {
	Up        key.Binding // ↑/k
	Down      key.Binding // ↓/j
	Left      key.Binding // ← (slider only)
	Right     key.Binding // → (slider only)
	Inc       key.Binding // +/= (±small)
	Dec       key.Binding // -   (±small)
	IncMed    key.Binding // ]   (±medium)
	DecMed    key.Binding // [   (±medium)
	AllSlider key.Binding // a   (set all)
	Confirm   key.Binding // Enter
	Cancel    key.Binding // Esc
	Quit      key.Binding // q / ctrl+c
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k")),
		Down:      key.NewBinding(key.WithKeys("down", "j")),
		Left:      key.NewBinding(key.WithKeys("left")),
		Right:     key.NewBinding(key.WithKeys("right")),
		Inc:       key.NewBinding(key.WithKeys("+", "=")),
		Dec:       key.NewBinding(key.WithKeys("-")),
		IncMed:    key.NewBinding(key.WithKeys("]")),
		DecMed:    key.NewBinding(key.WithKeys("[")),
		AllSlider: key.NewBinding(key.WithKeys("a")),
		Confirm:   key.NewBinding(key.WithKeys("enter")),
		Cancel:    key.NewBinding(key.WithKeys("esc")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c")),
	}
}
