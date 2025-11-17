package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
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
	case YesNoResponseMsg:
		p.app.screens.pop()
		return p.app, func() tea.Msg { return p.callback(msg.response) }
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Left):
			p.selected = true
		case key.Matches(msg, keys.Right):
			p.selected = false
		case msg.Type == tea.KeyEnter:
			return p, func() tea.Msg { return YesNoResponseMsg{p.selected} }
		case key.Matches(msg, keys.Yes):
			return p, func() tea.Msg { return YesNoResponseMsg{true} }
		case key.Matches(msg, keys.No) || msg.Type == tea.KeyEsc:
			return p, func() tea.Msg { return YesNoResponseMsg{false} }
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
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, fmt.Sprintf("%s   %s", yes, no))
	question := questionStyle.Render(p.question)
	return yesNoStyle.Render(lipgloss.JoinVertical(lipgloss.Center, question, buttons))
}

// Response to prompt: Yes [true] No [false]
type YesNoResponseMsg struct {
	response bool
}

func (a *ApplicationTUI) AskYesNo(question string, callback func(bool) tea.Msg) {
	yesNoModel := &YesNoPrompt{question: question, selected: true, callback: callback, app: a}
	callingModel := a.screens.cur()
	overlayModel := overlay.New(yesNoModel, callingModel, overlay.Center, overlay.Center, 0, 0)
	a.screens.push(overlayModel)
}
