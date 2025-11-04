package tui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

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
		case key.Matches(msg, keys.Back):
			if al.focus == viewLists {
				return m, GoBack
			}
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
	fl  *app.FilmList
	app *ApplicationTUI
}

func (li viewListItem) FilterValue() string {
	return li.fl.Name
}

func (li viewListItem) Title() string {
	return li.fl.Name
}

func (li viewListItem) Description() string {
	var ordered string
	if li.fl.Ordered {
		ordered = "Ordered"
	} else {
		ordered = "Unordered"
	}
	var suffix string
	nw, err := li.fl.NextWatch()
	switch {
	case errors.Is(err, app.ErrListEmpty):
		suffix = "List Empty"
	case errors.Is(err, app.ErrNoValidFilm):
		suffix = "List Complete"
	default:
		suffix = fmt.Sprintf("Next Watch : %s", nw)
	}
	return fmt.Sprintf("%s :: %s", ordered, suffix)
}

func (d viewListsDelegate) Update(msg tea.Msg, ls *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case TrackedChangedMsg:
		ls.SetItems(creatViewListItems(d.app))
	case tea.KeyMsg:
		li, ok := ls.SelectedItem().(viewListItem)
		if !ok { // SelectedItem will return nil when list is empty
			return nil
		}
		if msg.Type == tea.KeyEnter {
			li.fl.ToggleOrdered()
			return TrackedChanged
		} else if key.Matches(msg, keys.Delete) {
			d.app.AskYesNo(fmt.Sprintf("Stop tracking list %s?", li.Title()), func(b bool) tea.Msg {
				return removeListMsg{ok: b}
			})
		}
	case removeListMsg:
		li, ok := ls.SelectedItem().(viewListItem)
		if !ok {
			panic(fmt.Sprintf("(Add List) viewListDelegate item should be viewListItem, instead item is %T", ls.SelectedItem()))
		}
		if msg.ok {
			if err := d.app.RemoveList(li.fl); err != nil {
				log.Print(err.Error())
				return nil
			}
			return TrackedChanged
		}
	}
	return nil
}

// Create list of view-list items for view-list pane.
func creatViewListItems(a *ApplicationTUI) []list.Item {
	items := make([]list.Item, 0, len(a.TrackedLists))
	for _, v := range a.TrackedLists {
		items = append(items, viewListItem{v, a})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.Compare(items[i].FilterValue(), items[j].FilterValue()) < 0
	})
	return items
}

func MakeViewListPane(a *ApplicationTUI) *ListSelector {
	return MakeListSelector(a, "Tracked Lists", creatViewListItems(a), viewListsDelegate{DefaultDelegate: listStyleDelegate(), app: a})
}

// ----- Search Pane

type searchListsItem struct {
	app.FilmList
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
		if msg.Type == tea.KeyEnter || key.Matches(msg, keys.Delete) {
			if !d.app.IsListTracked(fl.Url) {
				if err := d.app.AddList(&fl.FilmList); err != nil {
					log.Print(err.Error())
					return nil
				}
				return TrackedChanged
			} else {
				d.app.AskYesNo(fmt.Sprintf("Stop tracking list %s?", fl.Title()), func(b bool) tea.Msg {
					return removeListMsg{ok: b}
				})
			}
		}
	case removeListMsg:
		if !d.app.IsListTracked(fl.Url) {
			panic("should not ask to stop tracking untracked list")
		}
		if msg.ok {
			if err := d.app.RemoveList(&fl.FilmList); err != nil {
				log.Print(err.Error())
				return nil
			}
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
	if d.app.IsListTracked(fl.Url) {
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
		items = append(items, &searchListsItem{*lh})
	}
	return MakeSearchModel(a, items, "Enter URL or search lists...", searchListsDelegate{listStyleDelegate(), a}, func(query string) {
		if err := a.AddListFromUrl(query); !errors.Is(err, app.ErrInvalidUrl) {
			log.Printf("could not add query as url, %s", err)
		}
	})
}
