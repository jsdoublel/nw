package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Bubbletea model joining two other models; it passes all updates to both and
// is also focusable.
type JoinModel struct {
	models   []focusable       // list of models to be joined
	vertical bool              // join vertical; otherwise horizontal
	pos      lipgloss.Position // lipgloss position for join
}

func (jm *JoinModel) Init() tea.Cmd { return nil }

func (jm *JoinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	for _, model := range jm.models {
		_, c := model.Update(msg)
		cmds = append(cmds, c)
	}
	c := tea.Batch(cmds...)
	return nil, c
}

func (jm *JoinModel) View() string {
	views := make([]string, len(jm.models))
	for _, m := range jm.models {
		views = append(views, m.View())
	}
	if jm.vertical {
		return lipgloss.JoinVertical(jm.pos, views...)
	}
	return lipgloss.JoinHorizontal(jm.pos, views...)
}

func (jm *JoinModel) Focus() {
	for _, m := range jm.models {
		m.Focus()
	}
}

func (jm *JoinModel) Unfocus() {
	for _, m := range jm.models {
		m.Unfocus()
	}
}
