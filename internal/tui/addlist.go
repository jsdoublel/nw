package tui

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

type TrackedChangedMsg struct{} // message sent indicating the tracked lists have changed

func TrackedChanged() tea.Msg { return TrackedChangedMsg{} }

type pane int // panes in the window

const (
	searchList pane = iota
	viewLists
)

type AddListsScreen struct {
	panes []tea.Model
	focus pane
	app   *ApplicationTUI
}

func MakeAddListScreen(a *ApplicationTUI) AddListsScreen {
	return AddListsScreen{
		panes: []tea.Model{MakeSearchListPane(a), MakeViewListPane(a)},
		focus: searchList,
		app:   a,
	}
}

func (al *AddListsScreen) Init() tea.Cmd { return nil }

func (al *AddListsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := al.panes[al.focus].Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.MoveRight):
			al.focusView()
		case key.Matches(msg, keys.MoveLeft):
			al.focusSearch()
		}
	case TrackedChangedMsg:
		m, cmd = al.viewPane().Update(msg)
	}
	return m, cmd
}

func (al *AddListsScreen) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		al.searchPane().View(),
		al.viewPane().View(),
	)
}

func (al *AddListsScreen) searchPane() tea.Model {
	return al.panes[searchList]
}

func (al *AddListsScreen) viewPane() tea.Model {
	return al.panes[viewLists]
}

func (al *AddListsScreen) focusView() {
	al.focus = viewLists
	sp, ok := al.searchPane().(*SearchModel)
	if !ok {
		panic("First pane on add list screen should be *SearchModel")
	}
	sp.Focus(false)
	vp, ok := al.viewPane().(*ListSelector)
	if !ok {
		panic("Second pane on add list screen should be *ListSelector")
	}
	vp.Focus(true)
}

func (al *AddListsScreen) focusSearch() {
	al.focus = searchList
	sp, ok := al.searchPane().(*SearchModel)
	if !ok {
		panic("First pane on add list screen should be *SearchModel")
	}
	sp.Focus(true)
	vp, ok := al.viewPane().(*ListSelector)
	if !ok {
		panic("Second pane on add list screen should be *ListSelector")
	}
	vp.Focus(false)
}

// ----- View Pane

type viewListsDelegate struct {
	list.DefaultDelegate
	app *ApplicationTUI
}

type viewListItem struct {
	app.FilmList
}

func (li viewListItem) FilterValue() string {
	return li.Name
}

func (li viewListItem) Title() string {
	return li.Name
}

func (li viewListItem) Description() string {
	if li.Ordered {
		return "Ordered"
	} else {
		return "Unordered"
	}
}

func (d viewListsDelegate) Update(msg tea.Msg, ls *list.Model) tea.Cmd {
	if _, ok := msg.(TrackedChangedMsg); ok {
		items := make([]list.Item, 0, len(d.app.TrackedLists))
		for _, v := range d.app.TrackedLists {
			items = append(items, viewListItem{*v})
		}
		ls.SetItems(items)
	}
	return nil
}

func MakeViewListPane(a *ApplicationTUI) *ListSelector {
	items := make([]list.Item, 0, len(a.TrackedLists))
	for _, v := range a.TrackedLists {
		items = append(items, viewListItem{*v})
	}
	return MakeListSelector(a, items, viewListsDelegate{DefaultDelegate: listStyleDelegate(), app: a})
}

// ----- Search Pane

type searchListsItem struct {
	app.FilmList
	selected bool
}

func (li searchListsItem) FilterValue() string {
	return li.Name
}

func (li searchListsItem) Title() string {
	return li.Name
}

func (li searchListsItem) Description() string {
	desc := fmt.Sprintf("%d films", li.NumFilms)
	if len(li.Desc) != 0 {
		desc += fmt.Sprintf(" :: %s", li.Desc)
	}
	return desc
}

type searchListsDelegate struct {
	list.DefaultDelegate
	app *ApplicationTUI
}

type removeListMsg struct {
	ok bool
}

func (d searchListsDelegate) Update(msg tea.Msg, ls *list.Model) tea.Cmd {
	item := ls.SelectedItem()
	if item == nil {
		return nil
	}
	fl, ok := item.(*searchListsItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", ls.SelectedItem()))
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if !fl.selected {
				if err := d.app.AddList(&fl.FilmList); err != nil {
					log.Print(err.Error())
					return nil
				}
				fl.selected = true
				return TrackedChanged
			} else {
				d.app.AskYesNo(fmt.Sprintf("Stop tracking list %s?", fl.Title()), func(b bool) tea.Msg {
					return removeListMsg{ok: b}
				})
			}
		}
	case removeListMsg:
		if !fl.selected {
			panic("should not ask to stop tracking untracked list")
		}
		if msg.ok {
			if err := d.app.RemoveList(&fl.FilmList); err != nil {
				log.Print(err.Error())
				return nil
			}
			fl.selected = false
			return TrackedChanged
		}
	}
	return nil
}

func (d searchListsDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	fl, ok := listItem.(*searchListsItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", listItem))
	}
	dd := d.DefaultDelegate
	if fl.selected {
		dd.Styles.NormalTitle = dd.Styles.NormalTitle.
			Foreground(luster).Bold(true)
		dd.Styles.NormalDesc = dd.Styles.NormalDesc.
			Foreground(lack)
	}
	dd.Render(w, m, index, listItem)
}

func MakeSearchListPane(a *ApplicationTUI) *SearchModel {
	items := make([]list.Item, 0, len(a.ListHeaders))
	for _, lh := range a.ListHeaders {
		items = append(items, &searchListsItem{*lh, a.IsListTracked(lh)})
	}
	return MakeSearchModel(a, items, "Enter URL or search lists...", searchListsDelegate{listStyleDelegate(), a}, func(query string) {
		if err := a.AddListFromUrl(query); !errors.Is(err, app.ErrInvalidUrl) {
			log.Printf("could not add query as url, %s", err)
		}
	})
}
