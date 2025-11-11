package tui

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	tmdb "github.com/cyruzin/golang-tmdb"

	"github.com/jsdoublel/nw/internal/app"
)

type SearchFilms struct {
	model SearchModel
	app   *ApplicationTUI
}

func (sf *SearchFilms) Init() tea.Cmd { return nil }

func (sf *SearchFilms) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := sf.model.Update(msg)
	return m, cmd
}

func (sf *SearchFilms) View() string { return sf.model.View() }

type FilmResultItem tmdb.MovieResult

func (fi FilmResultItem) FilterValue() string { return "" }
func (fi FilmResultItem) String() string {
	releaseDate, err := time.Parse("2006-01-02", fi.ReleaseDate)
	if err != nil {
		log.Printf("error parsing date for film results %s, %s", fi.Title, fi.ReleaseDate)
	}
	return fmt.Sprintf("%s (%d)", fi.Title, releaseDate.Year())
}

type filmSearchDelegate struct{}

func (d filmSearchDelegate) Height() int  { return 1 }
func (d filmSearchDelegate) Spacing() int { return 0 }

func (d filmSearchDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	li, ok := listItem.(FilmResultItem)
	if !ok {
		panic("item in film results is not FilmResultsItem")
	}
	out := trimAndPadString(li.String(), paneWidth)
	if index == m.Index() {
		out = filmSearchSelectedStyle.Render(out)
	} else {
		out = filmSearchItemStyle.Render(out)
	}
	if _, err := io.WriteString(w, out); err != nil {
		log.Printf("error rendering film search list, %s", err)
	}
}

func (d filmSearchDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Pads a string to reach given width or, if it is too long trim it, adding an
// ellipse to the end.
func trimAndPadString(s string, width int) string {
	length := utf8.RuneCountInString(s)
	var b strings.Builder
	if length < width {
		b.WriteString(s)
		b.WriteString(strings.Repeat(" ", width-length))
		return b.String()
	}
	return string(append([]rune(s)[:width-1], ellipse))
}

func MakeSearchFilms(a *ApplicationTUI) *SearchFilms {
	inputAction := func(s string) tea.Cmd {
		results, err := app.SearchFilms(s)
		if err != nil {
			log.Print(err.Error())
		}
		items := make([]list.Item, len(results))
		for i, r := range results {
			items[i] = FilmResultItem(r)
		}
		return func() tea.Msg { return UpdateSearchItemsMsg{items: items} }
	}
	EnterAction := func(s string) {}
	model := *MakeSearchModel(
		a,
		make([]list.Item, 0),
		"Search films...",
		filmSearchDelegate{},
		inputAction,
		EnterAction,
	)
	model.switchToSearch()
	return &SearchFilms{
		model: model,
		app:   a,
	}
}
