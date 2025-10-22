package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	searchInputStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	searchListStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	cursorStyle      = lipgloss.NewStyle()
)

type SearchModel struct {
	input textinput.Model
	list  list.Model
	app   *ApplicationTUI
}

func MakeSearchModel(a *ApplicationTUI, items []list.Item, searchText string, listDelegate list.ItemDelegate) *SearchModel {
	ti := textinput.New()
	ti.Placeholder = searchText
	ti.Cursor.Style = cursorStyle
	// ti.Focus()
	frameW, _ := searchInputStyle.GetFrameSize()
	ti.Width = max(max(listPaneWidth-frameW, 0), lipgloss.Width(searchText))
	list := list.New(items, listDelegate, 0, 0)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowFilter(false)
	list.SetShowStatusBar(false)
	return &SearchModel{input: ti, list: list, app: a}
}

func (sm *SearchModel) Init() tea.Cmd {
	return nil
}

func (sm *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		frameW, frameH := searchListStyle.GetFrameSize()
		listWidth := max(listPaneWidth-frameW, 0)
		listHeight := max(listPaneHeight-frameH, 0)
		sm.list.SetSize(listWidth, listHeight)
	}
	var ic tea.Cmd
	sm.input, ic = sm.input.Update(msg)
	var lc tea.Cmd
	sm.list, lc = sm.list.Update(msg)
	return sm, tea.Batch(ic, lc)
}

func (sm *SearchModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		searchInputStyle.Width(listPaneWidth).Render(sm.input.View()),
		searchListStyle.Width(listPaneWidth).Render(sm.list.View()),
	)
}
