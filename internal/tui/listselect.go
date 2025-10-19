package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

var lsStyle = lipgloss.NewStyle().Margin(1, 2)

// Window for scrolling and selecting from list
type ListSelector struct {
	list   list.Model
	app    *app.Application
	action func(it list.Item) error // action on pressing enter
}

// func MakeListSelector(app *app.Application) *ListSelector {
func MakeListSelector(app *app.Application, items []list.Item, delegate list.ItemDelegate, action func(it list.Item) error) *ListSelector {
	return &ListSelector{
		list:   list.New(items, delegate, 0, 0),
		app:    app,
		action: action,
	}
}

func (lp *ListSelector) Init() tea.Cmd {
	return nil
}

func (lp *ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return lp, tea.Quit
		case tea.KeyEnter:
			if err := lp.action(lp.list.SelectedItem()); err != nil {
				log.Print(err.Error())
			}
		}
	case tea.WindowSizeMsg:
		h, v := lsStyle.GetFrameSize()
		lp.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	lp.list, cmd = lp.list.Update(msg)
	return lp, cmd
}

func (lp *ListSelector) View() string {
	return lsStyle.Render(lp.list.View())
}
