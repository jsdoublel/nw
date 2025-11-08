package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const minCast = 5

type NewFilmDetailsMsg struct {
	film app.Film
}

type FilmDetailsModel struct {
	film    *app.FilmRecord
	focused bool
	app     *ApplicationTUI
	err     error // records error if film details could not be retrieved
}

func (fd *FilmDetailsModel) Init() tea.Cmd { return nil }

func (fd *FilmDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (fd *FilmDetailsModel) View() string {
	if fd.err != nil {
		return filmDetailsErrStyle.Width(listPaneWidth).Render(fd.errorText())
	}
	return filmDetailsStyle.Width(listPaneWidth).Render(fd.renderDetails())
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
	colWidth := listPaneWidth/2 - 1
	rightText := lipgloss.NewStyle().Width(colWidth).Render(fd.film.Details.Overview)
	castLimit := max(minCast, lipgloss.Height(rightText)-4)
	var b strings.Builder
	if directors := fd.directorLine(); directors != "" {
		b.WriteString(flimDirStyle.Render(directors))
	}
	if runtime := fd.film.Details.Runtime; runtime > 0 {
		b.WriteString(fmt.Sprintf("\n%d minutes", runtime))
	}
	if cast := fd.castLine(castLimit); cast != "" {
		b.WriteString("\n\n")
		b.WriteString(filmCastHeaderStyle.Render("Cast"))
		b.WriteString("\n")
		b.WriteString(cast)
	}
	leftText := lipgloss.NewStyle().Width(colWidth).Render(b.String())
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

func MakeFilmDetailsModel(f *app.Film, a *ApplicationTUI) *FilmDetailsModel {
	fr, err := a.FilmStore.Lookup(*f)
	filmDetailsStyle = filmDetailsStyle.BorderForeground(unfocused)
	return &FilmDetailsModel{
		film:    fr,
		focused: false,
		app:     a,
		err:     err,
	}
}
