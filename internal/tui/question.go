package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Question struct {
	question string
	input    textinput.Model
}

func (q *Question) Init() tea.Cmd {
	return nil
}

func (q *Question) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ti, cmd := q.input.Update(msg)
	q.input = ti
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyEnter {
			return q, tea.Quit
		}
	}
	return q, cmd
}

func (q *Question) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, startupTextStyle.Render("\n"+q.question), startupInputStyle.Render(q.input.View()))
}

func AskQuestion(question, placeholder string) string {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 20
	q := &Question{question: question, input: ti}
	p := tea.NewProgram(q)
	result, err := p.Run()
	if err != nil {
		log.Printf("startup failed with %s\n", err)
		return ""
	}
	q, _ = result.(*Question)
	return strings.TrimSpace(q.input.Value())
}
