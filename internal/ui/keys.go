package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all key bindings.
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Inc     key.Binding // +
	Dec     key.Binding // -
	IncMed  key.Binding // ]
	DecMed  key.Binding // [
	IncLrg  key.Binding // }
	DecLrg  key.Binding // {
	Slider  key.Binding // s
	Refresh key.Binding // r
	Confirm key.Binding // Enter
	Cancel  key.Binding // Esc
	Quit    key.Binding // q / ctrl+c
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Inc:     key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "brightness +")),
		Dec:     key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "brightness -")),
		IncMed:  key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "+medium")),
		DecMed:  key.NewBinding(key.WithKeys("["), key.WithHelp("[", "-medium")),
		IncLrg:  key.NewBinding(key.WithKeys("}"), key.WithHelp("}", "+large")),
		DecLrg:  key.NewBinding(key.WithKeys("{"), key.WithHelp("{", "-large")),
		Slider:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "slider")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("↵", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
