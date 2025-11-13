package tui

import (
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var heightCutoff = int(math.Floor(float64(paneHeight) * 1.8))

// Bubbletea model joining two other models; it passes all updates to both and
// is also focusable.
type JoinModel struct {
	main      focusable
	secondary focusable
	pos       lipgloss.Position // lipgloss position for join
	app       *ApplicationTUI
}

func (jm *JoinModel) Init() tea.Cmd { return nil }

func (jm *JoinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && !jm.bothOnScreen() {
		if msg.Type == tea.KeyEnter {
			jm.app.screens.push(jm.secondary)
		}
	}
	cmds := make([]tea.Cmd, 0, 2)
	switch {
	case jm.bothOnScreen():
		_, sc := jm.secondary.Update(msg)
		cmds = append(cmds, sc)
		fallthrough
	default:
		_, mc := jm.main.Update(msg)
		cmds = append(cmds, mc)
	}
	return jm, tea.Batch(cmds...)
}

func (jm *JoinModel) View() string {
	if jm.app.width < 2*paneWidth && jm.bothOnScreen() {
		return joinTallThin.Height(jm.app.height - 1). // - 1 for status bar
								Render(lipgloss.JoinVertical(jm.pos, jm.secondary.View(), jm.main.View()))
	} else if jm.app.width < 2*paneWidth {
		return jm.main.View()
	}
	return lipgloss.JoinHorizontal(jm.pos, jm.secondary.View(), jm.main.View())
}

func (jm *JoinModel) Focus() {
	jm.secondary.Focus()
	jm.main.Focus()
}

func (jm *JoinModel) Unfocus() {
	jm.secondary.Unfocus()
	jm.main.Unfocus()
}

func (jm *JoinModel) bothOnScreen() bool {
	return jm.app.width >= 2*paneWidth || jm.app.height > heightCutoff
}
