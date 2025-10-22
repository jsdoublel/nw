package tui

import (
	"fmt"
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

var appViewStyle = lipgloss.NewStyle().PaddingTop(8)

type ScreenStack []tea.Model

func (ss *ScreenStack) push(m tea.Model)    { *ss = append(*ss, m) }
func (ss *ScreenStack) pop()                { *ss = (*ss)[:len(*ss)-1] }
func (ss ScreenStack) cur() tea.Model       { return ss[len(ss)-1] }
func (ss *ScreenStack) replace(s tea.Model) { (*ss)[len(*ss)-1] = s }

// Main model struct that drives NW TUI
type ApplicationTUI struct {
	app.Application
	screens ScreenStack
	pending func(bool) tea.Msg
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
	aModel := ApplicationTUI{Application: *application}
	aModel.screens.push(MakeAddListPane(&aModel))
	p := tea.NewProgram(&aModel, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (a *ApplicationTUI) Init() tea.Cmd { return nil }

func (a *ApplicationTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	prev := a.screens.cur() // record top model to check whether we need to update at the end
	cur, cmd := prev.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case YesNoResponse:
		if _, ok := cur.(YesNoPrompt); !ok {
			panic("model sending YesNoResponse should be YesNoPrompt")
		}
		a.screens.pop()
		return a, func() tea.Msg { return a.pending(msg.response) }
	}
	if prev == a.screens.cur() {
		a.screens.replace(cur)
	}
	return a, cmd
}

func (a *ApplicationTUI) View() string {
	content := a.screens.cur().View()
	if a.width == 0 || a.height == 0 {
		return content
	}
	padded := appViewStyle.Render(content)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Top, padded)
}
