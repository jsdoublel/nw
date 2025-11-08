package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsdoublel/nw/internal/app"
)

const (
	mainScreenDetails mainScreenPane = iota
	mainScreenNW
	mainScreenViewList
)

type focusable interface {
	Focus()
	Unfocus()
}

type mainScreenPane int

type MainScreen struct {
	panes []tea.Model
	focus mainScreenPane
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
		ms.panes[mainScreenDetails].View(),
		ms.panes[mainScreenNW].View(),
		ms.panes[mainScreenViewList].View(),
	)
}

func (ms *MainScreen) focusRight() {
	if m, ok := ms.panes[ms.focus].(focusable); ok {
		m.Unfocus()
	}
	if int(ms.focus) != len(ms.panes)-1 {
		ms.focus++
	}
	if m, ok := ms.panes[ms.focus].(focusable); ok {
		m.Focus()
	}
}

func (ms *MainScreen) focusLeft() {
	if m, ok := ms.panes[ms.focus].(focusable); ok {
		m.Unfocus()
	}
	if int(ms.focus) != 0 {
		ms.focus--
	}
	if m, ok := ms.panes[ms.focus].(focusable); ok {
		m.Focus()
	}
}

func (ms *MainScreen) NewFilmDetails(film app.Film) {
	ms.panes[mainScreenDetails] = MakeFilmDetailsModel(&film, ms.app)
}

func MakeMainScreen(a *ApplicationTUI) *MainScreen {
	return &MainScreen{
		panes: []tea.Model{MakeFilmDetailsModel(a.NWQueue.Stacks[0][0], a), MakeNWModel(a), MakeViewListPane(a)},
		focus: mainScreenNW,
		app:   a,
	}
}
