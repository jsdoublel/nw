package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	searchMode mode = iota
	normalMode

	resultPainHeight = paneHeight - 3
)

type UpdateSearchItemsMsg struct{ items []list.Item }
type UpdateSearchFilterMsg struct{ filter string }

// Model with text search input and list of results below
type SearchModel struct {
	ListSelector
	input             textinput.Model
	queryAction       func(string)
	inputChangeAction func(string) tea.Cmd
	mode              mode
	focused           bool
	app               *ApplicationTUI
}

func MakeSearchModel(a *ApplicationTUI, items []list.Item, searchText string, delegate list.ItemDelegate, inputAction func(string) tea.Cmd, queryAction func(string)) *SearchModel {
	list := MakeListSelector(a, "", items, delegate)
	list.list.SetHeight(resultPainHeight)
	list.list.SetShowStatusBar(false)
	list.list.SetShowTitle(false)
	list.list.SetFilteringEnabled(true)
	ti := textinput.New()
	ti.Placeholder = searchText
	ti.Cursor.Style = cursorStyle
	frameW, _ := searchInputStyle.GetFrameSize()
	ti.Width = paneWidth - frameW - len(ti.Prompt)
	searchListStyle = searchListStyle.BorderForeground(focused)
	searchInputStyle = searchInputStyle.BorderForeground(unfocused)
	return &SearchModel{
		ListSelector:      *list,
		input:             ti,
		inputChangeAction: inputAction,
		queryAction:       queryAction,
		mode:              normalMode,
		focused:           true,
		app:               a,
	}
}

func (sm *SearchModel) Init() tea.Cmd { return nil }

func (sm *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	prev := strings.TrimSpace(sm.input.Value())
	var ic tea.Cmd
	sm.input, ic = sm.input.Update(msg)
	query := strings.TrimSpace(sm.input.Value())
	var ac tea.Cmd
	if prev != query {
		ac = sm.inputChangeAction(query)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyEnter || key.Matches(msg, keys.MoveDown):
			if sm.mode == searchMode {
				sm.queryAction(query)
				sm.switchToNormal()
				return sm, nil
			}
		case key.Matches(msg, keys.Back):
			if sm.mode == searchMode && sm.input.Value() != "" {
				sm.input.SetValue("")
				sm.list.ResetFilter()
				return sm, nil
			} else if sm.mode == searchMode {
				sm.switchToNormal()
				return sm, nil
			}
			return sm, GoBack
		case key.Matches(msg, keys.Search), key.Matches(msg, keys.MoveUp):
			sm.switchToSearch()
		}
	case UpdateSearchItemsMsg:
		sm.list.SetItems(msg.items)
	case UpdateSearchFilterMsg:
		sm.list.SetFilterText(msg.filter)
	}

	var lc tea.Cmd
	sm.list, lc = sm.list.Update(msg)
	return sm, tea.Batch(ic, lc, ac)
}

func (sm *SearchModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		searchInputStyle.Width(paneWidth).Render(sm.input.View()),
		searchListStyle.Width(paneWidth).Height(resultPainHeight).
			Render(sm.list.View()),
	)
}

func (sm *SearchModel) switchToSearch() {
	sm.input.Focus()
	sm.mode = searchMode
	searchListStyle = searchListStyle.BorderForeground(unfocused)
	searchInputStyle = searchInputStyle.BorderForeground(focused)
}

func (sm *SearchModel) switchToNormal() {
	sm.input.Blur()
	sm.mode = normalMode
	searchListStyle = searchListStyle.BorderForeground(focused)
	searchInputStyle = searchInputStyle.BorderForeground(unfocused)
}

func (sm *SearchModel) Focus() {
	sm.focused = true
	if sm.mode == normalMode {
		searchListStyle = searchListStyle.BorderForeground(focused)
	} else {
		searchInputStyle = searchInputStyle.BorderForeground(focused)
	}
}

func (sm *SearchModel) Unfocus() {
	sm.focused = false
	searchListStyle = searchListStyle.BorderForeground(unfocused)
	searchInputStyle = searchInputStyle.BorderForeground(unfocused)
}
