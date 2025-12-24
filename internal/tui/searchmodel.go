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

type UpdateSearchItemsMsg struct {
	items []list.Item
	query string
}
type UpdateSearchFilterMsg struct{ filter string }

// Model with text search input and list of results below
type SearchModel struct {
	ListSelector
	input             textinput.Model
	queryEnterAction  func(string, list.Item) tea.Cmd
	inputChangeAction func(string) tea.Cmd
	defaultMode       mode
	mode              mode
	focused           bool
	inputStyle        lipgloss.Style
	listStyle         lipgloss.Style
	app               *ApplicationTUI
}

func MakeSearchModel(
	a *ApplicationTUI,
	items []list.Item,
	searchText string,
	delegate list.ItemDelegate,
	defaultMode mode,
	inputAction func(string) tea.Cmd,
	queryAction func(string, list.Item) tea.Cmd,
) *SearchModel {
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
	sm := &SearchModel{
		ListSelector:      *list,
		input:             ti,
		inputChangeAction: inputAction,
		queryEnterAction:  queryAction,
		defaultMode:       defaultMode,
		focused:           true,
		inputStyle:        searchInputStyle,
		listStyle:         searchListStyle,
		app:               a,
	}
	if defaultMode == normalMode {
		sm.switchToNormal()
	} else {
		sm.switchToSearch()
	}
	return sm
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

	var ec tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyEnter:
			if sm.mode == searchMode {
				ec = sm.queryEnterAction(query, sm.list.SelectedItem())
			}
			fallthrough
		case key.Matches(msg, keys.MoveDown):
			if sm.mode == searchMode {
				sm.switchToNormal()
				return sm, ec
			}
		case key.Matches(msg, keys.Back):
			if sm.mode == searchMode && sm.input.Value() != "" {
				sm.input.SetValue("")
				sm.list.ResetFilter()
				return sm, sm.inputChangeAction("")
			}
			return sm, GoBack
		case key.Matches(msg, keys.Search), key.Matches(msg, keys.MoveUp):
			sm.switchToSearch()
		}
	case UpdateSearchItemsMsg:
		if strings.TrimSpace(sm.input.Value()) == msg.query {
			sm.list.SetItems(msg.items)
		}
	case UpdateSearchFilterMsg:
		sm.list.SetFilterText(msg.filter)
	case UpdateScreenMsg:
		if sm.defaultMode == normalMode {
			sm.switchToNormal()
		} else {
			sm.switchToSearch()
		}
	}

	var lc tea.Cmd
	sm.list, lc = sm.list.Update(msg)
	return sm, tea.Batch(ic, lc, ac, ec)
}

func (sm *SearchModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		sm.inputStyle.Width(paneWidth).Render(sm.input.View()),
		sm.listStyle.Width(paneWidth).Height(resultPainHeight).
			Render(sm.list.View()),
	)
}

func (sm *SearchModel) switchToSearch() {
	sm.input.Focus()
	sm.mode = searchMode
	sm.listStyle = sm.listStyle.BorderForeground(unfocusedColor)
	sm.inputStyle = sm.inputStyle.BorderForeground(focusedColor)
}

func (sm *SearchModel) switchToNormal() {
	sm.input.Blur()
	sm.mode = normalMode
	sm.listStyle = sm.listStyle.BorderForeground(focusedColor)
	sm.inputStyle = sm.inputStyle.BorderForeground(unfocusedColor)
}

func (sm *SearchModel) Focus() {
	sm.focused = true
	if sm.mode == normalMode {
		sm.listStyle = sm.listStyle.BorderForeground(focusedColor)
	} else {
		sm.inputStyle = sm.inputStyle.BorderForeground(focusedColor)
	}
}

func (sm *SearchModel) Unfocus() {
	sm.focused = false
	sm.listStyle = sm.listStyle.BorderForeground(unfocusedColor)
	sm.inputStyle = sm.inputStyle.BorderForeground(unfocusedColor)
}
