package tui

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/jsdoublel/bubbletea-overlay"

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
	*app.Application
	screens ScreenStack
	status  StatusBarModel
	help    help.Model
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
	a := ApplicationTUI{Application: application}
	p := tea.NewProgram(&a, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (a *ApplicationTUI) Init() tea.Cmd {
	a.status = StatusBarModel{app: a}
	a.help = help.New()
	return updateUserDataCmd(a, true)
}

func (a *ApplicationTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := a.UpdateRouter(msg)
	switch msg := msg.(type) {
	case userDataLoadedMsg:
		a.screens.pop()          // remove loading screen
		if len(a.screens) == 0 { // we need different behavior on startup vs. update
			a.screens.push(MakeMainScreen(a))
		} else {
			return a, UpdateScreen
		}
	case userDataFailedMsg:
		if ss, ok := a.screens.cur().(*SplashScreenModel); ok {
			ss.SetError(msg.err)
		} else {
			panic("userDataFailedMsg received without splash screen")
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = a.width
		a.checkSize()
	case GoBackMsg:
		if len(a.screens) == 1 {
			return a, tea.Quit
		}
		a.screens.pop()
		return a, UpdateScreen
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Update):
			return a, updateUserDataCmd(a, false)
		case key.Matches(msg, keys.StopWatch):
			a.StopDiscordRPC()
		case key.Matches(msg, keys.Help):
			a.help.ShowAll = !a.help.ShowAll
		}
	}
	return a, tea.Batch(cmds...)
}

func (a *ApplicationTUI) View() string {
	cur := a.screens.cur()
	main := lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, cur.View())
	if _, ok := cur.(*ResizeLockModel); ok {
		return main
	}
	if _, ok := cur.(*SplashScreenModel); ok {
		return main
	}
	compStatus := overlay.Composite(a.status.View(), main, overlay.Left, overlay.Top, 0, 0)
	return overlay.Composite(a.help.View(keys), compStatus, overlay.Left, overlay.Bottom, 0, 0)
}

// Handle update rounting with overlays
func (a *ApplicationTUI) UpdateRouter(msg tea.Msg) []tea.Cmd {
	var c, bc, sc tea.Cmd
	_, c = a.screens.cur().Update(msg) // always returns nil if cur is an overlay
	_, sc = a.status.Update(msg)
	if ov, ok := a.screens.cur().(*overlay.Model); ok {
		_, c = ov.Foreground.Update(msg)
		if msg, ok := msg.(UpdateScreenMsg); ok {
			_, bc = ov.Background.Update(msg)
		}
	}
	return []tea.Cmd{c, bc, sc}
}

func (a *ApplicationTUI) checkSize() {
	toSmall := a.height <= paneHeight || a.width <= paneWidth
	if _, ok := a.screens.cur().(*ResizeLockModel); ok && !toSmall {
		a.screens.pop()
	} else if !ok && toSmall {
		a.screens.push(&ResizeLockModel{a})
	}
}

type userDataLoadedMsg struct{}
type userDataFailedMsg struct{ err error }

func updateUserDataCmd(app *ApplicationTUI, check bool) tea.Cmd {
	if len(app.screens) != 0 { // check to prevent user spamming Update key
		if _, ok := app.screens.cur().(*SplashScreenModel); ok {
			return nil
		}
	}
	splash, cmd := MakeSplashScreen()
	app.screens.push(splash)
	return tea.Batch(cmd, func() tea.Msg {
		if err := app.UpdateUserData(check); err != nil {
			return userDataFailedMsg{err}
		}
		return userDataLoadedMsg{}
	})
}
