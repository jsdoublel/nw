package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model for yes no question pop-up
type YesNoPrompt struct {
	question string // question to be asked
	selected bool   // true is yes
	callback func(bool) tea.Msg
	app      *ApplicationTUI
}

func (p *YesNoPrompt) Init() tea.Cmd { return nil }

func (p *YesNoPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case YesNoResponse:
		if p.app.screens.cur() != p {
			panic("model sending YesNoResponse should be top of stack YesNoPrompt")
		}
		p.app.screens.pop()
		return p.app, func() tea.Msg { return p.callback(msg.response) }
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Left):
			p.selected = true
		case key.Matches(msg, keys.Right):
			p.selected = false
		case msg.Type == tea.KeyEnter:
			return p, func() tea.Msg { return YesNoResponse{p.selected} }
		case key.Matches(msg, keys.Yes):
			return p, func() tea.Msg { return YesNoResponse{true} }
		case key.Matches(msg, keys.No) || msg.Type == tea.KeyEsc:
			return p, func() tea.Msg { return YesNoResponse{false} }
		}
	}
	return p, nil
}

func (p *YesNoPrompt) View() string {
	var yes, no string
	if p.selected {
		yes = yesNoSelected.Render("Yes")
		no = yesNoUnselected.Render("No")
	} else {
		yes = yesNoUnselected.Render("Yes")
		no = yesNoSelected.Render("No")
	}
	sep := lipgloss.NewStyle().Width(3).Render("")
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yes, sep, no)
	question := questionStyle.Render(p.question)
	content := lipgloss.JoinVertical(lipgloss.Center, "", question, "", buttons, "")
	return yesNoStyle.Render(content)
}

// Response to prompt: Yes [true] No [false]
type YesNoResponse struct {
	response bool
}

func (a *ApplicationTUI) AskYesNo(question string, callback func(bool) tea.Msg) {
	a.screens.push(&YesNoPrompt{question: question, selected: true, callback: callback, app: a})
}
