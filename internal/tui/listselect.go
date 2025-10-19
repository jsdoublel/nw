package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const (
	listPaneWidth  = 64
	listPaneHeight = 36
)

var lsStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder())

// Window for scrolling and selecting from list
type ListSelector struct {
	list           list.Model
	app            *app.Application
	action         func(it list.Item) error // action on pressing enter
	viewportWidth  int
	viewportHeight int
}

// func MakeListSelector(app *app.Application) *ListSelector {
func MakeListSelector(a *app.Application, items []list.Item, delegate list.ItemDelegate, action func(it list.Item) error) *ListSelector {
	return &ListSelector{
		list:   list.New(items, delegate, 0, 0),
		app:    a,
		action: action,
	}
}

func (lp *ListSelector) Init() tea.Cmd {
	return nil
}

func (lp *ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if err := lp.action(lp.list.SelectedItem()); err != nil {
				log.Print(err.Error())
			}
		}
	case tea.WindowSizeMsg:
		frameW, frameH := lsStyle.GetFrameSize()
		listWidth := max(listPaneWidth-frameW, 0)
		listHeight := max(listPaneHeight-frameH, 0)
		lp.viewportWidth = msg.Width
		lp.viewportHeight = msg.Height
		lp.list.SetSize(listWidth, listHeight)
	}
	var cmd tea.Cmd
	lp.list, cmd = lp.list.Update(msg)
	return lp, cmd
}

func (lp *ListSelector) View() string {
	content := lsStyle.Width(listPaneWidth).Height(listPaneHeight).Render(lp.list.View())
	if lp.viewportWidth == 0 || lp.viewportHeight == 0 {
		return content
	}
	return lipgloss.Place(lp.viewportWidth, lp.viewportHeight, lipgloss.Center, lipgloss.Center, content)
}
