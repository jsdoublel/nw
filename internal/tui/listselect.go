package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listPaneWidth  = 64
	listPaneHeight = 28
)

var lsStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder())

// Window for scrolling and selecting from list
type ListSelector struct {
	list list.Model
	app  *ApplicationTUI
}

// func MakeListSelector(app *app.Application) *ListSelector {
func MakeListSelector(a *ApplicationTUI, items []list.Item, delegate list.ItemDelegate) *ListSelector {
	return &ListSelector{
		list: list.New(items, delegate, 0, 0),
		app:  a,
	}
}

func (ls *ListSelector) Init() tea.Cmd {
	return nil
}

func (ls *ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		frameW, frameH := lsStyle.GetFrameSize()
		listWidth := max(listPaneWidth-frameW, 0)
		listHeight := max(listPaneHeight-frameH, 0)
		ls.list.SetSize(listWidth, listHeight)
	}
	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

func (ls *ListSelector) View() string {
	return lsStyle.Width(listPaneWidth).Height(listPaneHeight).Render(ls.list.View())
}
