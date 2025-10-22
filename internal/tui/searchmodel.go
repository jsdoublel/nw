package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	searchMode mode = iota
	normalMode
)

var (
	searchInputStyle          = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	unfocusedSearchInputStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5c5c5c"))
	searchListStyle           = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	unfocusedSearchListStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5c5c5c"))
	cursorStyle               = lipgloss.NewStyle()
)

type SearchModel struct {
	input textinput.Model
	list  list.Model
	focus mode
	app   *ApplicationTUI
}

func MakeSearchModel(a *ApplicationTUI, items []list.Item, searchText string, listDelegate list.ItemDelegate) *SearchModel {
	ti := textinput.New()
	ti.Placeholder = searchText
	ti.Cursor.Style = cursorStyle
	ti.Focus()
	frameW, _ := searchInputStyle.GetFrameSize()
	ti.Width = max(max(listPaneWidth-frameW, 0), lipgloss.Width(searchText))

	list := list.New(items, listDelegate, 0, 0)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowFilter(false)
	list.SetShowStatusBar(false)
	list.DisableQuitKeybindings()
	return &SearchModel{input: ti, list: list, focus: searchMode, app: a}
}

func (sm *SearchModel) Init() tea.Cmd {
	return nil
}

func (sm *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	prev := sm.input.Value()
	var ic tea.Cmd
	sm.input, ic = sm.input.Update(msg)
	query := sm.input.Value()
	if prev != query {
		sm.list.SetFilterText(strings.TrimSpace(query))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		frameW, frameH := searchListStyle.GetFrameSize()
		listWidth := max(listPaneWidth-frameW, 0)
		listHeight := max(listPaneHeight-frameH, 0)
		sm.list.SetSize(listWidth, listHeight)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if sm.focus == searchMode {
				sm.switchToNormal()
				return sm, nil
			} else {
				return sm, tea.Quit
			}
		case "i", ":", "/":
			if sm.focus == normalMode {
				sm.switchToSearch()
			}
		}
	}

	var lc tea.Cmd
	sm.list, lc = sm.list.Update(msg)
	return sm, tea.Batch(ic, lc)
}

func (sm *SearchModel) View() string {
	var inSty, listSty lipgloss.Style
	if sm.focus == searchMode {
		inSty = searchInputStyle
		listSty = unfocusedSearchListStyle
	} else {
		inSty = unfocusedSearchInputStyle
		listSty = searchListStyle
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		inSty.Width(listPaneWidth).Render(sm.input.View()),
		listSty.Width(listPaneWidth).Render(sm.list.View()),
	)
}

func (sm *SearchModel) switchToSearch() {
	sm.input.Focus()
	sm.focus = searchMode
}

func (sm *SearchModel) switchToNormal() {
	sm.input.Blur()
	sm.focus = normalMode
}
