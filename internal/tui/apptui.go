package tui

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

type GoBackMsg struct{}

func GoBack() tea.Msg { return GoBackMsg{} }

type UpdateScreenMsg struct{}

func UpdateScreen() tea.Msg { return UpdateScreenMsg{} }

type ScreenStack []tea.Model

func (ss *ScreenStack) push(m tea.Model) { *ss = append(*ss, m) }
func (ss *ScreenStack) pop()             { *ss = (*ss)[:len(*ss)-1] }
func (ss ScreenStack) cur() tea.Model    { return ss[len(ss)-1] }

// Main model struct that drives NW TUI
type ApplicationTUI struct {
	app.Application
	screens ScreenStack
	width   int
	height  int
}

func RunApplicationTUI(username string) error {
	logf, err := tea.LogToFile(filepath.Join(app.NWDataPath, "nw.log"), "")
	if err != nil {
		return fmt.Errorf("could not set up logging, %w", err)
	}
	defer func() { _ = logf.Close() }()
	log.Print("nw starting...")
	application, err := app.Load(username)
	if err != nil {
		return fmt.Errorf("could not load application data, %w", err)
	}
	defer application.Shutdown()
	if err := application.Init(); err != nil {
		return err
	}
	a := ApplicationTUI{Application: *application}
	p := tea.NewProgram(&a, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (a *ApplicationTUI) Init() tea.Cmd {
	a.screens.push(MakeMainScreen(a))
	return nil
}

func (a *ApplicationTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := a.screens.cur().Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case GoBackMsg:
		if len(a.screens) == 1 {
			return a, tea.Quit
		}
		a.screens.pop()
		return a, UpdateScreen
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Update):
			if err := a.UpdateUserData(); err != nil {
				log.Printf("failed to update user data, %s", err)
			}
			return a, UpdateScreen
		case key.Matches(msg, keys.StopWatch):
			a.StopDiscordRPC()
		}
	}
	return a, cmd
}

func (a *ApplicationTUI) View() string {
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, a.screens.cur().View())
}
