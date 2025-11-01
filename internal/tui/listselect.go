package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	listPaneWidth  = 64
	listPaneHeight = 42
)

// Window for scrolling and selecting from list
type ListSelector struct {
	list    list.Model
	focused bool
	app     *ApplicationTUI
}

func MakeListSelector(a *ApplicationTUI, items []list.Item, delegate list.ItemDelegate) *ListSelector {
	list := list.New(items, delegate, listPaneWidth, listPaneHeight)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowFilter(false)
	list.SetShowStatusBar(false)
	list.DisableQuitKeybindings()
	return &ListSelector{
		list: list,
		app:  a,
	}
}

func (ls *ListSelector) Init() tea.Cmd {
	return nil
}

func (ls *ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

func (ls *ListSelector) View() string {
	lsSty := lsStyle
	if !ls.focused {
		lsSty = lsStyle.BorderForeground(unfocusedColor)
	}
	return lsSty.Width(listPaneWidth).Height(listPaneHeight).Render(ls.list.View())
}

func (ls *ListSelector) Focus(focused bool) {
	ls.focused = focused
}
