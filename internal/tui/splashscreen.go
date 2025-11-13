package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashScreenModel struct{ err error }

func (ss *SplashScreenModel) Init() tea.Cmd { return nil }

func (ss *SplashScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && key.Matches(msg, keys.Back) {
		return ss, GoBack
	}
	return ss, nil
}

func (ss *SplashScreenModel) View() string {
	if ss.err != nil {
		return lipgloss.NewStyle().Foreground(red).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			fmt.Sprintf("error %s", ss.err),
			fmt.Sprintf("press %s to exit.", keys.Back.Help().Key),
		))
	}
	return "Loading..."
}

func (ss *SplashScreenModel) SetError(err error) {
	ss.err = err
}
