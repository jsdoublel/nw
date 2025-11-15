package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type PopupModel struct {
	text string
}

func (pm *PopupModel) Init() tea.Cmd { return nil }

func (pm *PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && (msg.Type == tea.KeyEnter || key.Matches(msg, keys.Back)) {
		return pm, GoBack
	}
	return pm, nil
}

func (pm *PopupModel) View() string {
	return popupStyle.Render(lipgloss.JoinVertical(lipgloss.Center, pm.text, popupOkStyle.Render("OK")))
}

func (a *ApplicationTUI) Popup(text string) {
	popup := &PopupModel{text: text}
	backdrop := a.screens.cur()
	overlayModel := overlay.New(popup, backdrop, overlay.Center, overlay.Center, 0, 0)
	a.screens.push(overlayModel)
}
