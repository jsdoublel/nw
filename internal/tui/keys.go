package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding
	Search      key.Binding
	Delete      key.Binding
	Yes         key.Binding
	No          key.Binding
	Left        key.Binding
	Right       key.Binding
	Up          key.Binding
	Down        key.Binding
	MoveLeft    key.Binding
	MoveRight   key.Binding
	MoveUp      key.Binding
	MoveDown    key.Binding
	AddList     key.Binding
	SearchFilms key.Binding
	Update      key.Binding
	StopWatch   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Back,
		k.Help,
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.MoveLeft, k.MoveRight, k.MoveUp, k.MoveDown},
		{k.Search, k.Delete, k.Back, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("left/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("right/l", "right"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "down"),
	),
	MoveLeft: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "move focus left"),
	),
	MoveRight: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "move focus right"),
	),
	MoveUp: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "move focus up"),
	),
	MoveDown: key.NewBinding(
		key.WithKeys("ctrl+j"),
		key.WithHelp("ctrl+j", "move focus down"),
	),
	Search: key.NewBinding(
		key.WithKeys("i", ":", "/"),
		key.WithHelp("i/:", "enter text"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "Yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n", "N"),
		key.WithHelp("n", "No"),
	),
	AddList: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add lists"),
	),
	SearchFilms: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search films"),
	),
	Update: key.NewBinding(
		key.WithKeys("U"),
		key.WithHelp("U", "update"),
	),
	StopWatch: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "stop watching"),
	),
}
