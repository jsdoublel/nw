package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const joinDetailsPos = 0

const (
	mainScreenNWPos = iota
	mainScreenViewListPos
)

type focusable interface {
	tea.Model
	Focus()
	Unfocus()
}

type MainScreen struct {
	panes []focusable
	focus int
	app   *ApplicationTUI
}

func (ms *MainScreen) Init() tea.Cmd {
	return nil
}

func (ms *MainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := ms.panes[ms.focus].Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return m, GoBack
		case key.Matches(msg, keys.MoveRight):
			ms.focusRight()
		case key.Matches(msg, keys.MoveLeft):
			ms.focusLeft()
		case key.Matches(msg, keys.AddList):
			ms.app.screens.push(MakeAddListScreen(ms.app))
		}
	case NewFilmDetailsMsg:
		ms.NewFilmDetails(msg.film)
	case UpdateScreenMsg:
		for _, p := range ms.panes {
			p.Update(msg)
		}
	}
	return m, cmd
}

func (ms *MainScreen) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		ms.panes[mainScreenNWPos].View(),
		ms.panes[mainScreenViewListPos].View(),
	)
}

func (ms *MainScreen) focusRight() {
	ms.panes[ms.focus].Unfocus()
	if int(ms.focus) != len(ms.panes)-1 {
		ms.focus++
	}
	ms.panes[ms.focus].Focus()
}

func (ms *MainScreen) focusLeft() {
	ms.panes[ms.focus].Unfocus()
	if int(ms.focus) != 0 {
		ms.focus--
	}
	ms.panes[ms.focus].Focus()
}

func (ms *MainScreen) NewFilmDetails(film app.Film) {
	if jm, ok := ms.panes[mainScreenNWPos].(*JoinModel); ok {
		jm.models[joinDetailsPos] = MakeFilmDetailsModel(&film, ms.app)
		return
	}
	panic("film details not in correct position in JoinModel")
}

func MakeMainScreen(a *ApplicationTUI) *MainScreen {
	return &MainScreen{
		panes: []focusable{&JoinModel{
			models: []focusable{MakeFilmDetailsModel(a.NWQueue.Stacks[0][0], a), MakeNWModel(a)},
			pos:    lipgloss.Top,
		}, MakeViewListPane(a)},
		focus: mainScreenNWPos,
		app:   a,
	}
}
