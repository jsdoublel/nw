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

type addListPane int // panes in the add list window

const (
	addListSearchList addListPane = iota
	addListViewLists
)

type AddListsScreen struct {
	panes []tea.Model
	focus addListPane
	app   *ApplicationTUI
}

func MakeAddListScreen(a *ApplicationTUI) *AddListsScreen {
	return &AddListsScreen{
		panes: []tea.Model{MakeSearchListPane(a), MakeViewListPane(a)},
		focus: addListSearchList,
		app:   a,
	}
}

func (al *AddListsScreen) Init() tea.Cmd { return nil }

func (al *AddListsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := al.panes[al.focus].Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			if al.focus == addListViewLists {
				return m, GoBack
			}
		case key.Matches(msg, keys.MoveRight):
			al.focusView()
		case key.Matches(msg, keys.MoveLeft):
			al.focusSearch()
		}
	case UpdateScreenMsg:
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
	return al.panes[addListSearchList]
}

func (al *AddListsScreen) viewPane() tea.Model {
	return al.panes[addListViewLists]
}

func (al *AddListsScreen) focusView() {
	al.focus = addListViewLists
	sp, ok := al.searchPane().(*SearchModel)
	if !ok {
		panic("First pane on add list screen should be *SearchModel")
	}
	sp.Unfocus()
	vp, ok := al.viewPane().(*ListSelector)
	if !ok {
		panic("Second pane on add list screen should be *ListSelector")
	}
	vp.Focus()
}

func (al *AddListsScreen) focusSearch() {
	al.focus = addListSearchList
	sp, ok := al.searchPane().(*SearchModel)
	if !ok {
		panic("First pane on add list screen should be *SearchModel")
	}
	sp.Focus()
	vp, ok := al.viewPane().(*ListSelector)
	if !ok {
		panic("Second pane on add list screen should be *ListSelector")
	}
	vp.Unfocus()
}

// ----- Search Pane

type searchListsItem struct {
	fl *app.FilmList
}

func (li searchListsItem) FilterValue() string {
	return li.fl.Name
}

func (li searchListsItem) Title() string {
	return li.fl.Name
}

func (li searchListsItem) Description() string {
	desc := fmt.Sprintf("%d films", li.fl.NumFilms)
	if len(li.fl.Desc) != 0 {
		desc += fmt.Sprintf(" :: %s", li.fl.Desc)
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
	li, ok := item.(*searchListsItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", ls.SelectedItem()))
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter || key.Matches(msg, keys.Delete) {
			if !d.app.IsListTracked(li.fl.Url) {
				if err := d.app.AddList(li.fl); err != nil {
					log.Print(err.Error())
					return nil
				}
				return UpdateScreen
			} else {
				d.app.AskYesNo(fmt.Sprintf("Stop tracking list %s?", li.Title()), func(b bool) tea.Msg {
					return removeListMsg{ok: b}
				})
			}
		}
	case removeListMsg:
		if !d.app.IsListTracked(li.fl.Url) {
			panic("should not ask to stop tracking untracked list")
		}
		if msg.ok {
			if err := d.app.RemoveList(li.fl); err != nil {
				log.Print(err.Error())
				return nil
			}
			return UpdateScreen
		}
	}
	return nil
}

func (d searchListsDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	li, ok := listItem.(*searchListsItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", listItem))
	}
	dd := d.DefaultDelegate
	if d.app.IsListTracked(li.fl.Url) {
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
		items = append(items, &searchListsItem{lh})
	}
	inputChangeAction := func(s string) tea.Cmd {
		return func() tea.Msg { return UpdateSearchFilterMsg{filter: s} }
	}
	queryEnterAction := func(s string) {
		if err := a.AddListFromUrl(s); !errors.Is(err, app.ErrInvalidUrl) {
			log.Printf("could not add query as url, %s", err)
		}
	}
	return MakeSearchModel(
		a,
		items,
		"Enter URL or search lists...",
		searchListsDelegate{listStyleDelegate(), a},
		inputChangeAction,
		queryEnterAction,
	)
}
