package tui

import (
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsdoublel/nw/internal/app"
)

type nwDeleteFilmMsg struct {
	ok bool
}

func NWDeleteFilm() tea.Msg { return nwDeleteFilmMsg{} }

type itemTitle interface {
	Title() string
	Updated() bool
	FilterValue() string
}

type nwListItem struct {
	film    *app.Film
	updated bool
}

func (li nwListItem) Title() string       { return li.film.String() }
func (li nwListItem) Updated() bool       { return li.updated }
func (li nwListItem) FilterValue() string { return "" }

type stackSeparator int

func (li stackSeparator) Title() string       { return "" }
func (li stackSeparator) Updated() bool       { return false }
func (li stackSeparator) FilterValue() string { return "" }

type nwItemDelegate struct{}

func (d nwItemDelegate) Height() int  { return 1 }
func (d nwItemDelegate) Spacing() int { return 0 }

// Skips stack separators when scrolling. Also, sends updated film details when
// movement occurs
func (d nwItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	_, ssOk := m.SelectedItem().(stackSeparator)
	moved := false // movement key pressed
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Up):
			if ssOk {
				m.Select(m.Index() - 1)
			}
			moved = true
		case key.Matches(msg, keys.Down):
			if ssOk {
				m.Select(m.Index() + 1)
			}
			moved = true
		}
	}
	if li, ok := m.SelectedItem().(nwListItem); moved && ok {
		return func() tea.Msg { return NewFilmDetailsMsg{film: *li.film} }
	}
	return nil
}

func (d nwItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it := listItem.(itemTitle)
	var b strings.Builder
	b.WriteString("   ")
	fn := func(strs ...string) string { return nwItemStyle.Render(strings.Join(strs, "")) }
	if it.Updated() {
		fn = func(s ...string) string {
			return nwUpdatedItemStyle.Render(strings.Join(s, ""))
		}
		if index != 0 {
			b.Reset()
			b.WriteString(" + ")
		}
	}
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
	if index == 0 {
		b.Reset()
		b.WriteString(" Next Watch: ")
	}
	paddingLen := listPaneWidth - utf8.RuneCountInString(it.Title()) - utf8.RuneCountInString(b.String())
	b.WriteString(it.Title())
	var content string
	if paddingLen < 0 {
		content = string(append([]rune(b.String())[:utf8.RuneCountInString(b.String())+paddingLen-2], rune('…')))
	} else {
		b.WriteString(strings.Repeat(" ", paddingLen))
		content = b.String()
	}
	if _, err := fmt.Fprint(w, fn(content)); err != nil {
		log.Printf("error rendering NW queue, %s", err)
	}
}

type NWModel struct {
	list    list.Model
	focused bool
	app     *ApplicationTUI
}

func (nw *NWModel) Init() tea.Cmd { return nil }

func (nw *NWModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	li, ok := nw.list.SelectedItem().(nwListItem)
	if !ok {
		return nil, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keys.Delete) {
			nw.app.AskYesNo(
				fmt.Sprintf("Remove \"%s\" from queue?\nCannot be undone!", li.film),
				func(b bool) tea.Msg { return nwDeleteFilmMsg{ok: b} },
			)
		}
	case nwDeleteFilmMsg:
		if msg.ok {
			if err := nw.app.NWQueue.DeleteFilm(*li.film); err != nil {
				log.Printf("error after deleting film, %s", err)
			}
			return nil, UpdateScreen
		}
	case UpdateScreenMsg:
		nw.list.SetItems(makeNWItemsList(nw.app))
	}
	var cmd tea.Cmd
	nw.list, cmd = nw.list.Update(msg)
	return nil, cmd
}

func (nw *NWModel) View() string {
	return nwStyle.Width(listPaneWidth).Render(nw.list.View())
}

func (nw *NWModel) Focus() {
	nw.focused = true
	nwSeparatorStyle = nwSeparatorStyle.Foreground(focused)
	nwStyle = nwStyle.BorderForeground(focused)
}

func (nw *NWModel) Unfocus() {
	nw.focused = false
	nwSeparatorStyle = nwSeparatorStyle.Foreground(unfocused)
	nwStyle = nwStyle.BorderForeground(unfocused)
}

func MakeNWModel(a *ApplicationTUI) *NWModel {
	l := list.New(makeNWItemsList(a), nwItemDelegate{}, listPaneWidth, listPaneHeight)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return &NWModel{
		list:    l,
		app:     a,
		focused: true,
	}
}

func makeNWItemsList(a *ApplicationTUI) []list.Item {
	items := make([]list.Item, app.StackSize*app.NumberOfStacks+1+app.NumberOfStacks)
	var count, prevI int
	for i, j := range a.NWQueue.Positions() {
		if i != prevI {
			items[count] = stackSeparator(i)
			prevI = i
			count++
		}
		items[count] = nwListItem{film: a.NWQueue.Stacks[i][j], updated: a.NWQueue.LastUpdated(i, j)}
		count++
	}
	return items
}
