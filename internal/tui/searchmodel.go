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

	resultPainHeight = listPaneHeight - 3
)

// Model with text search input and list of results below
type SearchModel struct {
	ListSelector
	input       textinput.Model
	queryAction func(string)
	mode        mode
	focused     bool
	app         *ApplicationTUI
}

func MakeSearchModel(a *ApplicationTUI, items []list.Item, searchText string, delegate list.ItemDelegate, queryAction func(string)) *SearchModel {
	list := MakeListSelector(a, items, delegate)
	list.list.SetHeight(resultPainHeight)
	ti := textinput.New()
	ti.Placeholder = searchText
	ti.Cursor.Style = cursorStyle
	frameW, _ := searchInputStyle.GetFrameSize()
	ti.Width = max(max(listPaneWidth-frameW-len(ti.Prompt), 0), lipgloss.Width(searchText))
	return &SearchModel{
		ListSelector: *list,
		input:        ti,
		queryAction:  queryAction,
		mode:         normalMode,
		focused:      true,
		app:          a,
	}
}

func (sm *SearchModel) Init() tea.Cmd { return nil }

func (sm *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	prev := sm.input.Value()
	var ic tea.Cmd
	sm.input, ic = sm.input.Update(msg)
	query := sm.input.Value()
	if prev != query {
		sm.list.SetFilterText(strings.TrimSpace(query))
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
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
	}

	var lc tea.Cmd
	sm.list, lc = sm.list.Update(msg)
	return sm, tea.Batch(ic, lc)
}

func (sm *SearchModel) View() string {
	inSty := searchInputStyle
	listSty := searchListStyle
	if sm.mode == searchMode || !sm.focused {
		listSty = searchListStyle.BorderForeground(gray4)
	}
	if sm.mode == normalMode || !sm.focused {
		inSty = searchInputStyle.BorderForeground(gray4)
	}
	return lipgloss.JoinVertical(
		lipgloss.Center,
		inSty.Width(listPaneWidth).Render(sm.input.View()),
		listSty.Width(listPaneWidth).Height(resultPainHeight).Render(sm.list.View()),
	)
}

func (sm *SearchModel) switchToSearch() {
	sm.input.Focus()
	sm.mode = searchMode
}

func (sm *SearchModel) switchToNormal() {
	sm.input.Blur()
	sm.mode = normalMode
}

func (sm *SearchModel) Focus(focused bool) {
	sm.focused = focused
}
