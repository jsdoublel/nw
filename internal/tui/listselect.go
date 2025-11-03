package tui

import (
	"strings"

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

func MakeListSelector(a *ApplicationTUI, title string, items []list.Item, delegate list.ItemDelegate) *ListSelector {
	list := list.New(items, delegate, listPaneWidth, listPaneHeight)
	list.Title = title
	list.Styles.Title = list.Styles.Title.Background(addListTitleColor)
	list.SetShowHelp(false)
	list.SetShowFilter(false)
	list.SetFilteringEnabled(false)
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
	view := ls.list.View()
	if len(ls.list.Items()) == 0 {
		_, plural := ls.list.StatusBarItemName()
		placeholder := ls.list.Styles.NoItems.Render("No " + plural + ".")
		view = strings.Replace(view, placeholder, "", 1)
	}
	return lsSty.Width(listPaneWidth).Height(listPaneHeight).Render(view)
}

func (ls *ListSelector) Focus(focused bool) {
	ls.focused = focused
}
