package tui

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsdoublel/nw/internal/app"
)

type itemTitle interface {
	Title() string
	FilterValue() string
}

type nwListItem struct {
	film *app.Film
}

func (li nwListItem) Title() string       { return li.film.String() }
func (li nwListItem) FilterValue() string { return "" }

type stackSeparator int

func (li stackSeparator) Title() string       { return "" }
func (li stackSeparator) FilterValue() string { return "" }

type nwItemDelegate struct{}

func (d nwItemDelegate) Height() int  { return 1 }
func (d nwItemDelegate) Spacing() int { return 0 }

// Skips stack separators when scrolling
func (d nwItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	if _, ok := m.SelectedItem().(stackSeparator); !ok {
		return nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Up):
			m.Select(m.Index() - 1)
		case key.Matches(msg, keys.Down):
			m.Select(m.Index() + 1)
		}
	}
	return nil
}

func (d nwItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it := listItem.(itemTitle)
	fn := func(strs ...string) string { return nwItemStyle.Render(strings.Join(strs, "")) }
	if index == m.Index() {
		fn = func(s ...string) string {
			return nwSelectedItemStyle.Render(strings.Join(s, ""))
		}
	}
	if _, ok := listItem.(stackSeparator); ok {
		fn = func(_ ...string) string {
			return nwSeparatorStyle.Render(strings.Repeat("\u2500", listPaneWidth))
		}
	}
	var prefix string
	if index == 0 {
		prefix = " Next Watch: "
	}
	if _, err := fmt.Fprint(w, fn(prefix, "   ", it.Title(), strings.Repeat(" ", listPaneWidth-len(it.Title())-len(prefix)-3))); err != nil {
		log.Printf("error rendering NW queue, %s", err)
	}
}

type NWModel struct {
	list list.Model
	app  *ApplicationTUI
}

func (nw *NWModel) Init() tea.Cmd { return nil }

func (nw *NWModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	nw.list, cmd = nw.list.Update(msg)
	return nil, cmd
}

func (nw *NWModel) View() string {
	return nwStyle.Width(listPaneWidth).Render(nw.list.View())
}

func MakeNWModel(a *ApplicationTUI) *NWModel {
	l := list.New(makeNWItemsList(a), nwItemDelegate{}, listPaneWidth, listPaneHeight)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return &NWModel{
		list: l,
		app:  a,
	}
}

func makeNWItemsList(a *ApplicationTUI) []list.Item {
	items := make([]list.Item, app.StackSize*app.NumberOfStacks+1+app.NumberOfStacks)
	count := 0
	prevI := 0
	for i, j := range a.NWQueue.Positions() {
		if i != prevI {
			items[count] = stackSeparator(i)
			prevI = i
			count++
		}
		items[count] = nwListItem{film: a.NWQueue.Stacks[i][j]}
		count++
	}
	return items
}
