package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	mainScreenViewList mainScreenPane = iota
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
	if msg, ok := msg.(tea.KeyMsg); ok {
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
	}
	return m, cmd
}

func (ms *MainScreen) View() string {
	return ms.panes[ms.focus].View()
}

func (ms *MainScreen) focusRight() {
}

func (ms *MainScreen) focusLeft() {
}

func MakeMainScreen(a *ApplicationTUI) *MainScreen {
	return &MainScreen{
		panes: []tea.Model{MakeViewListPane(a)},
		focus: mainScreenViewList,
		app:   a,
	}
}
