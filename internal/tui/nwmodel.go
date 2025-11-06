package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsdoublel/nw/internal/app"
)

type nwListItem struct {
	film *app.Film
}

func (li nwListItem) Title() string {
	return li.film.String()
}

func (li nwListItem) Description() string {
	return ""
}

func (li nwListItem) FilterValue() string {
	return ""
}

type NWModel struct {
	list list.Model
	app  *ApplicationTUI
}

func (nw *NWModel) Init() tea.Cmd {
	return nil
}

func (nw *NWModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (nw *NWModel) View() string {
	return mainStyle.Render(nw.list.View())
}

func MakeNWModel(a *ApplicationTUI) *NWModel {
	return &NWModel{
		list: list.New(makeNWItemsList(a), list.NewDefaultDelegate(), listPaneWidth, listPaneHeight),
		app:  a,
	}
}

func makeNWItemsList(a *ApplicationTUI) []list.Item {
	items := make([]list.Item, app.StackSize*app.NumberOfStacks+1)
	count := 0
	for i, j := range a.NWQueue.Positions() {
		items[count] = nwListItem{film: a.NWQueue.Stacks[i][j]}
		count++
	}
	return items
}
