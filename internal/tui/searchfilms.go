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
	var b strings.Builder
	b.WriteString(fi.Title)
	if releaseDate, err := time.Parse("2006-01-02", fi.ReleaseDate); err == nil {
		b.WriteString(fmt.Sprintf(" (%d)", releaseDate.Year()))
	} else {
		log.Printf("error parsing date for film results %s, %s", fi.Title, fi.ReleaseDate)
	}
	return b.String()
}

type filmSearchDelegate struct{ app *ApplicationTUI }

func (d filmSearchDelegate) Height() int  { return 1 }
func (d filmSearchDelegate) Spacing() int { return 0 }

func (d filmSearchDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	li, ok := listItem.(FilmResultItem)
	if !ok {
		panic(fmt.Sprintf("item (type %T) in film results is not FilmResultsItem", li))
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
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || keyMsg.Type != tea.KeyEnter {
		return nil
	}
	if r, ok := m.SelectedItem().(FilmResultItem); ok {
		d.app.screens.push(MakeFilmDetailsModelFromResults(tmdb.MovieResult(r), d.app))
		return nil
	}
	panic(fmt.Sprintf("Film search result (type %T) is not a tmdb.MovieResult", m.SelectedItem()))
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
		query := strings.TrimSpace(s)
		return func() tea.Msg {
			if query == "" {
				return UpdateSearchItemsMsg{items: nil, query: query}
			}
			results, err := app.SearchFilms(query)
			if err != nil {
				text := fmt.Sprintf("film search failed, %s", err)
				log.Print(err)
				return statusMessageMsg{message: Message{text: text, error: true, timeout: time.Second * 10}}
			}
			items := make([]list.Item, len(results))
			for i, r := range results {
				items[i] = FilmResultItem(r)
			}
			return UpdateSearchItemsMsg{items: items, query: query}
		}
	}
	EnterAction := func(s string) {}
	model := *MakeSearchModel(
		a,
		make([]list.Item, 0),
		"Search films...",
		filmSearchDelegate{a},
		inputAction,
		EnterAction,
	)
	model.switchToSearch()
	return &SearchFilms{
		model: model,
		app:   a,
	}
}
