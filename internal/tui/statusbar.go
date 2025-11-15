package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const messageTimeout = 5 * time.Second

type StatusBarModel struct {
	message string
	app     *ApplicationTUI
}

func (sb *StatusBarModel) Init() tea.Cmd { return nil }

func (sb *StatusBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(statusClearMsg); ok {
		sb.message = ""
	}
	return sb, nil
}

func (sb *StatusBarModel) View() string {
	strs := make([]string, 0)
	if sb.app.DiscordRPC.Watching() {
		strs = append(strs, statusBarWatchingStyle.Render(
			fmt.Sprintf("Watching %s, press %s to stop", sb.app.DiscordRPC, keys.StopWatch.Help().Key),
		))
	}
	if sb.message != "" {
		strs = append(strs, statusBarMessageStyle.Render(sb.message))
	}
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}

type statusClearMsg struct{}

func (sb *StatusBarModel) SetMessage(message string) tea.Cmd {
	sb.message = message
	return tea.Tick(messageTimeout, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}
