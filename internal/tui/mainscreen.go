package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	mainScreenNW mainScreenPane = iota
	mainScreenViewList
)

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
	case UpdateScreenMsg:
		for _, p := range ms.panes {
			p.Update(msg)
		}
	}
	return m, cmd
}

func (ms *MainScreen) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Center, ms.panes[mainScreenNW].View(), ms.panes[mainScreenViewList].View())
}

func (ms *MainScreen) focusRight() {
	if int(ms.focus) != len(ms.panes)-1 {
		ms.focus++
	}
}

func (ms *MainScreen) focusLeft() {
	if int(ms.focus) != 0 {
		ms.focus--
	}
}

func MakeMainScreen(a *ApplicationTUI) *MainScreen {
	return &MainScreen{
		panes: []tea.Model{MakeNWModel(a), MakeViewListPane(a)},
		focus: mainScreenNW,
		app:   a,
	}
}
