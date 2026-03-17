package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const (
	mainScreenNWPos = iota
	mainScreenViewListPos
)

type MainScreen struct {
	model *JoinModel
	app   *ApplicationTUI
}

func (ms *MainScreen) Init() tea.Cmd {
	return nil
}

func (ms *MainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := ms.model.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return m, GoBack
		case key.Matches(msg, keys.AddList):
			ms.app.screens.push(MakeAddListScreen(ms.app))
		case key.Matches(msg, keys.SearchFilms):
			ms.app.screens.push(MakeSearchFilms(ms.app))
		}
	case NewFilmDetailsMsg:
		ms.NewFilmDetails(msg.film)
	}
	return m, cmd
}

func (ms *MainScreen) View() string {
	return ms.model.View()
}

func (ms *MainScreen) NewFilmDetails(film app.Film) {
	ms.model.secondary = MakeFilmDetailsModel(&film, ms.app)
}

func MakeMainScreen(a *ApplicationTUI) *MainScreen {
	return &MainScreen{
		model: &JoinModel{
			secondary: MakeFilmDetailsModel(a.NWQueue.Stacks[0][0], a),
			main: &MainScreenPanes{
				panes:    []focusable{MakeNWModel(a), MakeViewListPane(a)},
				focusIdx: mainScreenNWPos,
				app:      a,
			},
			pos: lipgloss.Top,
			app: a,
		},
		app: a,
	}
}

// -------- Main Screen Panes

type focusable interface {
	tea.Model
	Focus() tea.Cmd
	Unfocus()
}

// Panes on the Main Screen which focus can be toggled between
type MainScreenPanes struct {
	panes    []focusable
	focusIdx int // index of pane in focus
	app      *ApplicationTUI
}

func (p *MainScreenPanes) Init() tea.Cmd {
	return nil
}

func (p *MainScreenPanes) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := p.panes[p.focusIdx].Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return m, GoBack
		case key.Matches(msg, keys.MoveRight):
			return m, p.focusRight()
		case key.Matches(msg, keys.MoveLeft):
			return m, p.focusLeft()
		}
	case UpdateScreenMsg:
		for _, p := range p.panes {
			p.Update(msg)
		}
	}
	return m, cmd
}

func (p *MainScreenPanes) View() string {
	if p.app.width < 3*paneWidth { // 3 x paneWidth, as we consider film details
		return p.panes[p.focusIdx].View()
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		p.panes[mainScreenNWPos].View(),
		p.panes[mainScreenViewListPos].View(),
	)
}

func (p *MainScreenPanes) focusRight() tea.Cmd {
	p.panes[p.focusIdx].Unfocus()
	if int(p.focusIdx) != len(p.panes)-1 {
		p.focusIdx++
	}
	return p.panes[p.focusIdx].Focus()
}

func (p *MainScreenPanes) focusLeft() tea.Cmd {
	p.panes[p.focusIdx].Unfocus()
	if int(p.focusIdx) != 0 {
		p.focusIdx--
	}
	return p.panes[p.focusIdx].Focus()
}
