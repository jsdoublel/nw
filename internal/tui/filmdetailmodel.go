package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const minCast = 5

type NewFilmDetailsMsg struct {
	film app.Film
}

type FilmDetailsModel struct {
	film           *app.FilmRecord
	focused        bool
	actions        []FilmAction
	selectedAction int
	app            *ApplicationTUI
	err            error // records error if film details could not be retrieved
}

type FilmAction struct {
	label  string
	action func(app.Film) error
}

var filmActions = []FilmAction{
	{label: "Watch", action: func(f app.Film) error { return nil }},
	{label: "Poster", action: func(f app.Film) error { return nil }},
	{label: "Letterboxd", action: func(f app.Film) error { return nil }},
}

func (fd *FilmDetailsModel) Init() tea.Cmd { return nil }

func (fd *FilmDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Right):
			fd.actionRight()
		case key.Matches(msg, keys.Left):
			fd.actionLeft()
		case msg.Type == tea.KeyEnter:
			if err := fd.actions[fd.selectedAction].action(fd.film.Film); err != nil {
				log.Printf("action for film %s failed, %s", fd.film, err)
			}
		}
	}
	return nil, nil
}

func (fd *FilmDetailsModel) View() string {
	if fd.err != nil {
		return filmDetailsStyle.Foreground(red).Width(paneWidth).Render(fd.errorText())
	}
	return filmDetailsStyle.Width(paneWidth).
		Render(lipgloss.JoinVertical(lipgloss.Center, fd.renderDetails(), "", fd.renderActions()))
}

func (fd *FilmDetailsModel) Focus() {
	fd.focused = true
	filmDetailsStyle = filmDetailsStyle.BorderForeground(focused)
}

func (fd *FilmDetailsModel) Unfocus() {
	fd.focused = false
	filmDetailsStyle = filmDetailsStyle.BorderForeground(unfocused)
}

func (fd *FilmDetailsModel) errorText() string {
	if fd.err == nil {
		return ""
	}
	return fmt.Sprintf("Unable to load film details: %s", fd.err)
}

func (fd *FilmDetailsModel) renderDetails() string {
	title := filmTitleStyle.Render(fd.film.String())
	colWidth := paneWidth/2 - 1
	rightText := filmTextStyle.Width(colWidth).Render(fd.film.Details.Overview)
	limitAdj := 2
	var b strings.Builder
	if directors := fd.directorLine(); directors != "" {
		b.WriteString(flimDirStyle.Render(directors))
		limitAdj++
	}
	if runtime := fd.film.Details.Runtime; runtime > 0 {
		b.WriteString(fmt.Sprintf("\n%d minutes", runtime))
		limitAdj++
	}
	castLimit := max(minCast, lipgloss.Height(rightText)-limitAdj)
	if cast := fd.castLine(castLimit); cast != "" {
		b.WriteString("\n\n")
		b.WriteString(filmCastHeaderStyle.Render("Cast"))
		b.WriteString("\n")
		b.WriteString(cast)
	}
	leftText := filmTextStyle.Width(colWidth).Render(b.String())
	return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.JoinHorizontal(lipgloss.Top, leftText, rightText))
}

func (fd *FilmDetailsModel) castLine(limit int) string {
	cast := fd.film.Details.Credits.Cast
	names := make([]string, 0, limit)
	for _, member := range cast {
		if character := strings.TrimSpace(member.Character); character != "" {
			names = append(names, member.Name)
		}
		if len(names) == limit {
			break
		}
	}
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, "\n")
}

func (fd *FilmDetailsModel) directorLine() string {
	crew := fd.film.Details.Credits.Crew
	directors := make([]string, 0)
	for _, member := range crew {
		if strings.EqualFold(member.Job, "Director") {
			directors = append(directors, member.Name)
		}
	}
	if len(directors) == 0 {
		return ""
	}
	return fmt.Sprintf("dir. %s", strings.Join(directors, ", "))
}

func (fd *FilmDetailsModel) renderActions() string {
	sep := "   "
	var b strings.Builder
	for i, a := range fd.actions {
		if fd.selectedAction == i {
			b.WriteString(filmActionSelected.Render(a.label))
		} else {
			b.WriteString(filmActionUnselected.Render(a.label))
		}
		if i != len(fd.actions)-1 {
			b.WriteString(sep)
		}
	}
	return b.String()
}

func (fd *FilmDetailsModel) actionRight() {
	if fd.selectedAction < len(fd.actions) {
		fd.selectedAction++
	}
}

func (fd *FilmDetailsModel) actionLeft() {
	if fd.selectedAction != 0 {
		fd.selectedAction--
	}
}

func MakeFilmDetailsModel(f *app.Film, a *ApplicationTUI) *FilmDetailsModel {
	fr, err := a.FilmStore.Lookup(*f)
	filmDetailsStyle = filmDetailsStyle.BorderForeground(focused)
	return &FilmDetailsModel{
		film:    fr,
		focused: false,
		app:     a,
		actions: filmActions,
		err:     err,
	}
}
