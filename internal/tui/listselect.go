package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

var lsStyle = lipgloss.NewStyle().Margin(1, 2)

type ListSelector struct {
	list list.Model
	app  *app.Application
}

func MakeListSelector(app *app.Application) ListSelector {
	items := make([]list.Item, len(app.User.ListHeaders))
	for i, lh := range app.User.ListHeaders {
		items[i] = lh
	}
	return ListSelector{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
		app:  app,
	}
}

func (ls ListSelector) Init() tea.Cmd {
	return nil
}

func (ls ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			return ls, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := lsStyle.GetFrameSize()
		ls.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

func (ls ListSelector) View() string {
	return lsStyle.Render(ls.list.View())
}
