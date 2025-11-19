package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"

	"github.com/jsdoublel/nw/internal/app"
)

var NoRebind = []string{}

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
	About       key.Binding
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
		{k.Update, k.Delete, k.SearchFilms, k.AddList},
		{k.About, k.Back, k.Help, k.Quit},
	}
}

var keys = newKeyMap()

func newKeyMap() keyMap {
	return keyMap{
		Back:        binding(NoRebind, []string{"esc"}, "esc", "back"),
		Quit:        binding(app.Config.Keybinds.Quit, []string{"ctrl+c"}, "ctrl+c", "quit"),
		Help:        binding(NoRebind, []string{"?"}, "?", "toggle help"),
		Left:        binding(app.Config.Keybinds.Left, []string{"left", "h"}, "\u2190/h", "left"),
		Right:       binding(app.Config.Keybinds.Right, []string{"right", "l"}, "\u2192/l", "right"),
		Up:          binding(app.Config.Keybinds.Up, []string{"up", "k"}, "\u2191/k", "up"),
		Down:        binding(app.Config.Keybinds.Down, []string{"down", "j"}, "\u2193/j", "down"),
		MoveLeft:    binding(app.Config.Keybinds.MoveLeft, []string{"ctrl+h"}, "ctrl+h", "move focus left"),
		MoveRight:   binding(app.Config.Keybinds.MoveRight, []string{"ctrl+l"}, "ctrl+l", "move focus right"),
		MoveUp:      binding(app.Config.Keybinds.MoveUp, []string{"ctrl+k"}, "ctrl+k", "move focus up"),
		MoveDown:    binding(app.Config.Keybinds.MoveDown, []string{"ctrl+j"}, "ctrl+j", "move focus down"),
		Search:      binding(NoRebind, []string{"i", ":", "/"}, "i/:", "enter text"),
		Delete:      binding(app.Config.Keybinds.Delete, []string{"ctrl+d"}, "ctrl+d", "delete"),
		Yes:         binding(app.Config.Keybinds.Yes, []string{"y", "Y"}, "y", "Yes"),
		No:          binding(app.Config.Keybinds.No, []string{"n", "N"}, "n", "No"),
		AddList:     binding(app.Config.Keybinds.AddList, []string{"a"}, "a", "add lists"),
		SearchFilms: binding(app.Config.Keybinds.SearchFilms, []string{"/"}, "/", "search films"),
		Update:      binding(app.Config.Keybinds.Update, []string{"ctrl+u"}, "ctrl+u", "update data"),
		StopWatch:   binding(app.Config.Keybinds.StopWatch, []string{"ctrl+w"}, "ctrl+w", "stop watching"),
		About:       binding(app.Config.Keybinds.About, []string{"ctrl+a"}, "ctrl+a", "about"),
	}
}

func binding(cfg []string, fallback []string, defaultHelp string, desc string) key.Binding {
	keys := fallback
	help := defaultHelp
	if len(cfg) != 0 {
		keys = cfg
		help = strings.Join(cfg, "/")
	}
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(help, desc),
	)
}
