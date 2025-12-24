package tui

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsdoublel/nw/internal/app"
)

// Model for viewing and selecting film lists. It is used for both view and
// searching through lists.
type ListSelector struct {
	list    list.Model
	focused bool // window is focused (changes how it's drawn)
	style   lipgloss.Style
	app     *ApplicationTUI
}

func MakeListSelector(a *ApplicationTUI, title string, items []list.Item, delegate list.ItemDelegate) *ListSelector {
	l := list.New(items, delegate, paneWidth, paneHeight)
	l.Title = title
	l.Styles.Title = l.Styles.Title.Background(addListTitleColor).Foreground(textDark)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return &ListSelector{
		list:  l,
		app:   a,
		style: lsStyle.BorderForeground(unfocusedColor),
	}
}

func (ls *ListSelector) Init() tea.Cmd {
	return nil
}

func (ls *ListSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

func (ls *ListSelector) View() string {
	view := ls.list.View()
	if len(ls.list.Items()) == 0 {
		_, plural := ls.list.StatusBarItemName()
		placeholder := ls.list.Styles.NoItems.Render("No " + plural + ".")
		view = strings.Replace(view, placeholder, "", 1)
	}
	return ls.style.Width(paneWidth).Height(paneHeight).Render(view)
}

func (ls *ListSelector) Focus() {
	ls.focused = true
	ls.style = lsStyle.BorderForeground(focusedColor)
}

func (ls *ListSelector) Unfocus() {
	ls.focused = false
	ls.style = lsStyle.BorderForeground(unfocusedColor)
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
		suffix = nw.String()
	}
	return fmt.Sprintf("%s :: %s", ordered, suffix)
}

func (d viewListsDelegate) Update(msg tea.Msg, ls *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case UpdateScreenMsg:
		ls.SetItems(creatViewListItems(d.app))
	case tea.KeyMsg:
		li, ok := ls.SelectedItem().(viewListItem)
		if !ok { // SelectedItem will return nil when list is empty
			return nil
		}
		if msg.Type == tea.KeyEnter {
			li.fl.ToggleOrdered()
			return UpdateScreen
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
			return UpdateScreen
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
	return MakeListSelector(
		a,
		"Tracked Lists",
		creatViewListItems(a),
		viewListsDelegate{DefaultDelegate: listStyleDelegate(), app: a},
	)
}
