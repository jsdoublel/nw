package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var yesNoStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

// Model for yes no question pop-up
type YesNoPrompt struct {
	question string // question to be asked
}

func (p YesNoPrompt) Init() tea.Cmd { return nil }

func (p YesNoPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter", "y", "Y":
			return p, func() tea.Msg { return YesNoResponse{true} }
		case "esc", "n", "N", "q":
			return p, func() tea.Msg { return YesNoResponse{false} }
		}
	}
	return p, nil
}

func (p YesNoPrompt) View() string {
	return yesNoStyle.Render(fmt.Sprintf("\n %s \n\n  [Yes]   [No]\n", p.question))
}

// Response to prompt: Yes [true] No [false]
type YesNoResponse struct {
	response bool
}

func (a *ApplicationTUI) AskYesNo(question string, callback func(bool) tea.Msg) {
	a.screens.push(YesNoPrompt{question: question})
	a.pending = callback
}
