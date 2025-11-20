package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsdoublel/nw/internal/app"
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
	return lipgloss.JoinVertical(lipgloss.Left, q.question, q.input.View())
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
		log.Fatalf("startup failed with %s\n", err)
	}
	q, _ = result.(*Question)
	return strings.TrimSpace(q.input.Value())
}

func RunStartup(username string) {
	changed := false
	if username == "" && app.Config.Username == "" {
		app.Config.Username = AskQuestion("What is your Letterboxd username?:", "username")
	}
	if app.TMDBClient == nil && app.Config.ApiKey == "" {
		app.Config.ApiKey = AskQuestion("Enter your TMDB api key:", "api key")
		app.ApiInit()
	}
	if changed && strings.ToLower(AskQuestion("Save changes to config?", "y/n"))[0] == 'y' {
		if err := app.SaveConfig(); err != nil {
			log.Printf("failed to update config, %s", err)
		}
	}
}
