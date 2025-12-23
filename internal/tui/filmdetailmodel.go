package tui

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/pkg/browser"

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
	action func(app.FilmRecord) (tea.Cmd, error)
}

func init() { // discrads output from calling OpenURL which messes with View
	browser.Stdout = io.Discard
	browser.Stderr = io.Discard
}

func (fd *FilmDetailsModel) Init() tea.Cmd { return nil }

func (fd *FilmDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Right):
			fd.actionRight()
		case key.Matches(msg, keys.Left):
			fd.actionLeft()
		case msg.Type == tea.KeyEnter && fd.film != nil:
			cmd, err := fd.actions[fd.selectedAction].action(*fd.film)
			if err != nil {
				log.Printf("action \"%s\" failed for film %s, %s", fd.actions[fd.selectedAction].label, fd.film, err)
			}
			return fd, cmd
		case key.Matches(msg, keys.Back):
			return fd, GoBack
		}
	}
	return fd, nil
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
	filmDetailsStyle = filmDetailsStyle.BorderForeground(focusedColor)
}

func (fd *FilmDetailsModel) Unfocus() {
	fd.focused = false
	filmDetailsStyle = filmDetailsStyle.BorderForeground(unfocusedColor)
}

func (fd *FilmDetailsModel) errorText() string {
	if fd.err == nil {
		return ""
	}
	return fmt.Sprintf("Unable to load film details: %s", fd.err)
}

func (fd *FilmDetailsModel) renderDetails() string {
	title := filmTitleStyle.Render(fd.film.String())
	colWidth := paneWidth/2 - 2
	rightText := filmTextStyle.Width(colWidth).Render(fd.film.Details.Overview)
	limitAdj := 2 // adjust cast length limit based on height of text above
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
		names = append(names, member.Name)
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
	if fd.selectedAction < len(fd.actions)-1 {
		fd.selectedAction++
	}
}

func (fd *FilmDetailsModel) actionLeft() {
	if fd.selectedAction > 0 {
		fd.selectedAction--
	}
}

func filmActions(fr app.FilmRecord, a *ApplicationTUI) []FilmAction {
	actions := []FilmAction{
		{label: "Watch", action: func(fr app.FilmRecord) (tea.Cmd, error) { return nil, a.StartDiscordRPC(fr) }},
		{label: "Poster", action: func(fr app.FilmRecord) (tea.Cmd, error) {
			if path, err := app.DownloadPoster(fr); err != nil {
				return statusMessageCmd(Message{text: fmt.Sprintf("error %s", err), error: true}), err
			} else {
				return statusMessageCmd(Message{text: fmt.Sprintf("Poster downloaded to %s", path)}), nil
			}
		}},
	}
	if fr.Url != "" {
		actions = append(actions, FilmAction{
			label:  "Letterboxd",
			action: func(f app.FilmRecord) (tea.Cmd, error) { return nil, browser.OpenURL(f.Url) },
		})
	}
	if fr.Url == "" || app.Config.Features.AlwaysIncludeTMDB {
		actions = append(actions, FilmAction{
			label: "TMDB",
			action: func(f app.FilmRecord) (tea.Cmd, error) {
				return nil, browser.OpenURL(fmt.Sprintf("%s%d", app.TMDBFilmPathPrefix, fr.TMDBID))
			},
		})
	}
	if app.Config.Features.DisableDiscordRPC {
		return actions[1:]
	}
	return actions
}

func MakeFilmDetailsModel(f *app.Film, a *ApplicationTUI) *FilmDetailsModel {
	fr, err := a.FilmStore.Lookup(*f)
	actions := []FilmAction{}
	if fr != nil {
		actions = filmActions(*fr, a)
	}
	filmDetailsStyle = filmDetailsStyle.BorderForeground(focusedColor)
	return &FilmDetailsModel{
		film:    fr,
		focused: false,
		app:     a,
		actions: actions,
		err:     err,
	}
}

func MakeFilmDetailsModelFromResults(f tmdb.MovieResult, a *ApplicationTUI) *FilmDetailsModel {
	releaseYear, _ := app.ReleaseYear(f)
	details, err := app.TMDBFilm(int(f.ID))
	fr := app.FilmRecord{Film: app.Film{Title: details.Title, Year: uint(releaseYear)}, TMDBID: int(details.ID), Details: details}
	filmDetailsStyle = filmDetailsStyle.BorderForeground(focusedColor)
	return &FilmDetailsModel{
		film:    &fr,
		focused: false,
		app:     a,
		actions: filmActions(fr, a),
		err:     err,
	}
}
