package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const SplashText = " NW Loading...  "

var fakeCursor = spinner.Spinner{
	Frames: []string{string(cursor), " "},
	FPS:    time.Second / 4,
}

type SplashScreenModel struct {
	spinner spinner.Model
	tick    int
	err     error
}

func (ss *SplashScreenModel) Init() tea.Cmd { return nil }

func (ss *SplashScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	ss.spinner, cmd = ss.spinner.Update(msg)
	if msg, ok := msg.(tea.KeyMsg); ok && ss.err != nil && key.Matches(msg, keys.Back) {
		return ss, GoBack
	}
	return ss, cmd
}

func (ss *SplashScreenModel) View() string {
	if ss.err != nil {
		return lipgloss.NewStyle().Foreground(red).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			fmt.Sprintf("error %s", ss.err),
			fmt.Sprintf("press %s to exit.", keys.Back.Help().Key),
		))
	}
	ss.tick++
	ss.spinner.Style = splashSpinnerStyles[ss.tick%len(splashSpinnerStyles)]
	return fmt.Sprintf("%s%s", SplashText[:min(len(SplashText)-1, ss.tick)], ss.spinner.View())
}

func (ss *SplashScreenModel) SetError(err error) {
	ss.err = err
}

func MakeSplashScreen() (tea.Model, tea.Cmd) {
	ss := &SplashScreenModel{
		spinner: spinner.New(),
	}
	ss.spinner.Spinner = fakeCursor
	return ss, ss.spinner.Tick
}
